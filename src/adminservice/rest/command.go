package rest

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultExecutionTimeout make sure we will not run the command if
	// the time passed since emit is greater than 5 seconds.
	DefaultExecutionTimeout = time.Second * 5
)

// DefaultEventArgs ...
var DefaultEventArgs = map[string]interface{}{
	"execution_timeout": DefaultExecutionTimeout,
}

// RunningCommands defines a custom type to a map[string]command just to make possible to use methods in maps.
type RunningCommands struct {
	sync.RWMutex
	m map[string]models.Command
}

// running command list
var running = RunningCommands{
	m: make(map[string]models.Command),
}

// CommandList is the list of all commands supported by this route.
var commandList = []models.CommandInfo{
	models.CommandInfo{
		Name:        models.CommandCreateStatic,
		Description: "Creates all static JSON files or the one(s) specified in args.",
		PossibleArgs: []models.CommandArg{
			models.CommandArg{
				Name:        "-type",
				Description: "which type of static file to create (can be one of home, now_playing, upcoming)",
				Required:    false,
			},
		},
	},
	models.CommandInfo{
		Name:        models.CommandClearStatic,
		Description: "Clears all static JSON files or the one(s) specified in args.",
		PossibleArgs: []models.CommandArg{
			models.CommandArg{
				Name:        "-type",
				Description: "which type of static file to clear (can be one of home, now_playing, upcoming)",
				Required:    false,
			},
		},
	},
	models.CommandInfo{
		Name:        models.CommandClearShowtimes,
		Description: "Clears all showtimes or showtimes for the specified theaters in args.",
		PossibleArgs: []models.CommandArg{
			models.CommandArg{
				Name:        "-theater",
				Description: "which theater (id) to clear showtimes",
				Required:    false,
			},
		},
	},
	models.CommandInfo{
		Name:        models.CommandStartScraper,
		Description: "Starts all scrapers or the one(s) specified in args.",
		PossibleArgs: []models.CommandArg{
			models.CommandArg{
				Name:        "-theater",
				Description: "which theater (id) to start scraper",
				Required:    false,
			},
			models.CommandArg{
				Name:        "-type",
				Description: "which type of scraper to start (can be one of prices, now_playing or upcoming)",
				Required:    false,
			},
			models.CommandArg{
				Name:        "-ignore_last_run",
				Description: "whether we should ignore last run results or not",
				Required:    false,
			},
		},
	},
	models.CommandInfo{
		Name:        models.CommandCheckOpeningMovies,
		Description: "Checks which movies are new to the movies or the one(s) specified in args.",
		PossibleArgs: []models.CommandArg{
			models.CommandArg{
				Name:        "-theater",
				Description: "which theater (id) to check",
				Required:    false,
			},
		},
	},
	models.CommandInfo{
		Name:         models.CommandSyncScores,
		Description:  "Sync scores (Rotten Tomatoes and IMDb) of now playing movies",
		PossibleArgs: []models.CommandArg{},
	},
}

// CommandService ...
type CommandService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeCommands ...
func (rs *Service) ServeCommands(r *gin.Engine) {
	s := &CommandService{rs.data, rs.emitter}

	commands := r.Group("/commands")
	commands.GET("/", s.GetAll)
	commands.POST("/", s.RunCommand)
}

// GetAll ...
func (s *CommandService) GetAll(c *gin.Context) {
	apiutil.SendSuccess(c, commandList)
}

// RunCommand ...
func (s *CommandService) RunCommand(c *gin.Context) {
	cmd := models.Command{}
	if err := c.ShouldBind(&cmd); err != nil {
		apiutil.SendBadRequest(c)
		return
	}

	if !cmd.IsValid() {
		apiutil.SendSuccessOrError(c, "invalid command", nil)
		return
	}

	if running.isRunning(cmd) {
		apiutil.SendSuccessOrError(c, "command already running", nil)
		return
	}

	go run(s, cmd)

	apiutil.SendSuccess(c, "command started")
}

// Run runs all given commands sequentially.
func (s *CommandService) Run(commands ...models.Command) {
	size := len(commands)

	if size == 0 {
		// Do nothing.
	} else if size == 1 {
		run(s, commands[0])
	} else {
		for _, cmd := range commands {
			run(s, cmd)
		}
	}
}

// RunSingle ...
func run(s *CommandService, cmd models.Command) {
	running.add(cmd)

	printRunningList()

	name := cmd.Name
	args := models.ParseArgs(cmd.Args)

	switch name {
	case models.CommandCreateStatic, models.CommandClearStatic:
		event := &contracts.EventStaticDispatched{
			Name:             cmd.Name,
			Type:             args["type"],
			CinemaID:         args["theater"],
			DispatchTime:     time.Now().UTC(),
			ExecutionTimeout: DefaultExecutionTimeout,
		}

		ok, ID := models.GenerateTaskID(cmd.Name, cmd.Args)
		if ok {
			event.TaskID = ID
		}

		s.emitter.Emit(event)

	case models.CommandCheckOpeningMovies, models.CommandSyncScores:
		event := &contracts.EventCommandDispatched{
			Name:             cmd.Name,
			Type:             cmd.Name, // Command name == event type here
			DispatchTime:     time.Now().UTC(),
			ExecutionTimeout: DefaultExecutionTimeout,
		}

		ok, ID := models.GenerateTaskID(cmd.Name, cmd.Args)
		if ok {
			event.TaskID = ID
		}

		s.emitter.Emit(event)

	case models.CommandStartScraper:
		s.emitter.Emit(&contracts.EventCommandDispatched{
			Name:             cmd.Name,
			Type:             cmd.Name, // Command name == event type here
			Args:             cmd.Args,
			DispatchTime:     time.Now().UTC(),
			ExecutionTimeout: DefaultExecutionTimeout,
		})
	}

	running.remove(cmd)
}

func (r *RunningCommands) isRunning(c models.Command) bool {
	r.RLock()
	_, ok := r.m[c.Hash()]
	r.RUnlock()
	return ok
}

func (r *RunningCommands) add(c models.Command) {
	if c.Name == "" {
		return
	}

	r.Lock()
	r.m[c.Hash()] = c
	r.Unlock()
}

func (r *RunningCommands) remove(c models.Command) {
	if c.Name == "" {
		return
	}

	r.Lock()
	delete(r.m, c.Hash())
	r.Unlock()
}

func printRunningList() {
	running.RLock()

	fmt.Printf("\n------------------>   RUNNING COMMANDS   <------------------]\n\n")
	fmt.Printf("command_name\t\t-\tcommand_args\n")
	for _, c := range running.m {
		args := strings.Join(c.Args, " ")
		if args == "" {
			args = "(no args)"
		}

		fmt.Printf("%s\t\t-\t[ %s ]\n", c.Name, args)
	}
	fmt.Printf("\n\n")

	running.RUnlock()
}

// NewCommand creates a new command with the given name and args.
func NewCommand(name string, args []string) *models.Command {
	result := &models.Command{
		Name: name,
		Args: args,
	}

	if !result.IsValid() {
		return nil
	}

	return result
}
