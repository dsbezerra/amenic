package main

import (
	"net/http"
	"os"
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/dsbezerra/amenic/src/apiservice/listener"
	"github.com/dsbezerra/amenic/src/apiservice/v1"
	"github.com/dsbezerra/amenic/src/apiservice/v2"
	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/config"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	ServiceName = "API"
)

var (
	ctx = &Context{
		Service: ServiceName,
		Log:     logrus.WithFields(logrus.Fields{"App": ServiceName}),
	}
)

// Stats is used to get information about this service
type Stats struct {
	StartupTime time.Time `json:"startup_time"`
}

// Context ...
type Context struct {
	Service string
	Config  *config.ServiceConfig
	Stats   *Stats
	Data    persistence.DataAccessLayer
	Emitter messagequeue.EventEmitter
	Log     *logrus.Entry
}

func main() {
	settings, err := config.LoadConfiguration()
	if err != nil {
		ctx.Log.Fatal(err)
	}
	ctx.Config = settings

	tasks, err := config.LoadTasks()
	if err != nil {
		ctx.Log.Fatal(err)
	}

	conn, err := amqp.Dial(settings.AMQPMessageBroker)
	if err != nil {
		ctx.Log.Fatal(err)
	}

	eventEmitter, err := messagequeue.NewAMQPEventEmitter(conn, "events")
	if err != nil {
		ctx.Log.Fatal(err)
	}
	ctx.Emitter = eventEmitter

	eventListener, err := messagequeue.NewAMQPEventListener(conn, "events", ServiceName)
	if err != nil {
		ctx.Log.Fatal(err)
	}

	// Let's setup our logging database
	go setupLoggingDatabase(settings.DBLoggingConnection)

	// Let's setup our main database
	ctx.Log.Info("Setting up database...")

	data, err := mongolayer.NewMongoDAL(settings.DBConnection)
	if err != nil {
		ctx.Log.Fatal(err)
	}
	data.Setup()
	defer data.Close()

	ctx.Data = data
	ctx.Log.Info("Database setup completed!")

	// Ensure our tasks are saved in database.
	(data.(*mongolayer.MongoDAL)).EnsureTasksExists(tasks)

	// Initialize app context
	ctx.Stats = initStats()

	// Start event processor.
	p := listener.EventProcessor{
		Data:          data,
		Log:           ctx.Log,
		EventListener: eventListener,
	}
	go p.ProcessEvents()

	// Setup possible cron jobs
	{
		tasks, err := data.GetTasks(data.DefaultQuery().
			AddCondition("service", ServiceName))
		if err != nil {
			ctx.Log.Warnf("couldn't retrieve tasks for service %s", ServiceName)
		} else if tasks != nil && len(tasks) > 0 {
			loc, _ := time.LoadLocation("America/Sao_Paulo")

			c := cron.NewWithLocation(loc)
			defer c.Stop()

			for _, t := range tasks {
				ctx.Log.Infof("Setting up task %s", t.Name)
				for _, spec := range t.Cron {
					ctx.Log.Infof("Spec %s", spec)
					c.AddFunc(spec, func() {
						switch t.Type {
						case models.TaskCreateStatic:
							RunStaticTask(t)

							// NOTE(diego): This is a special case to make sure we have updated now_playing and
							// showtimes as soon as we enter thursday.
							if time.Now().Weekday() == time.Thursday {
								RunNowPlayingScraper()
							}
						default:
							ctx.Log.Infof("Unknown task type %s", t.Type)
						}
					})
				}

				if t.RunAtStart {
					RunStaticTask(t)
				}

				ctx.Log.Infof("Task %s setup completed!", t.Name)
			}

			c.Start()
		}
	}

	// Now let's build the main router for this API
	router := ctx.buildRouter()
	router.Run(settings.RESTEndpoint)
}

func setupLoggingDatabase(URI string) {
	if URI == "" {
		ctx.Log.Warnf("Logging database URI was not defined. Logs will not be persisted!")
		return
	}

	// ctx.Log.Info("Setting up logging database...")
	// hooker, err := mgorus.NewHooker(URI, "amenic-logs", "entries")
	// if err != nil {
	// 	ctx.Log.Fatal(err)
	// }
	// ctx.Log.Info("Logging database setup completed!")
	// ctx.Log.Logger.AddHook(hooker)
}

func initStats() *Stats {
	stats := &Stats{
		StartupTime: time.Now().UTC(),
	}
	return stats
}

func (ctx *Context) buildRouter() *gin.Engine {
	if ctx.Config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.SecureJsonPrefix(")]}',\n")

	// Setup middlewares.
	ctx.Log.WithField("ROUTER", "middlewares").Info("Setting up middlewares...")
	router.Use(helmet.Default())
	router.Use(
		rest.Init(),
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	// Setup static routes with an arbitrary cache control.
	router.Use(middlewares.StaticWithCache("/", static.LocalFile("./static", false)))

	// Setup route handlers.
	ctx.Log.WithField("ROUTER", "handlers").Info("Setting up route handlers...")
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello!")
	})

	v1.AddRoutes(router, ctx.Data, ctx.Emitter)
	v2.AddRoutes(router, ctx.Data, ctx.Emitter)
	return router
}

// Mode ...
func (ctx *Context) Mode() string {
	return os.Getenv("AMENIC_MODE")
}

// RunStaticTask ...
func RunStaticTask(t models.Task) {
	args := models.ParseArgs(t.Args)
	// NOTE: We emit here because our event processor do extra things to keep our tasks collection synced.
	ctx.Emitter.Emit(&contracts.EventStaticDispatched{
		Name:             t.Type,
		TaskID:           t.ID,
		Type:             args["type"],
		CinemaID:         args["theater"],
		DispatchTime:     time.Now().UTC(),
		ExecutionTimeout: time.Second * 5,
	})
}

// RunNowPlayingScraper emits an command dispatched event for the start_scraper
// task with now_playing type and ignore_last_run defined to true.
//
// This will be catched by event_listener in scraper service and run the proper
// scrapers.
func RunNowPlayingScraper() {
	nt := "start_scraper"
	ctx.Emitter.Emit(&contracts.EventCommandDispatched{
		Name: nt,
		Type: nt,
		Args: []string{
			"-type", "now_playing",
			"-ignore_last_run", "true",
		},
	})
}
