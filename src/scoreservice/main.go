package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dsbezerra/amenic/src/contracts"
	"github.com/dsbezerra/amenic/src/lib/config"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic/src/scoreservice/listener"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	ServiceName = "Score"
)

var (
	ctx = &Context{
		Service: ServiceName,
		Log:     logrus.WithFields(logrus.Fields{"App": ServiceName}),
	}
)

// Context ...
type Context struct {
	Service  string
	Log      *logrus.Entry
	Config   *config.ServiceConfig
	Data     persistence.DataAccessLayer
	Emitter  messagequeue.EventEmitter
	Listener messagequeue.EventListener
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
	ctx.Listener = eventListener

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

	// Start event processor.
	p := listener.EventProcessor{
		Data:          data,
		Log:           ctx.Log,
		EventListener: eventListener,
	}
	go p.ProcessEvents()

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	c := cron.NewWithLocation(loc)

	// Setup possible cron jobs
	{
		tasks, err := data.GetTasks(data.DefaultQuery().
			AddCondition("service", ServiceName))
		if err != nil {
			ctx.Log.Warnf("couldn't retrieve tasks for service %s", ServiceName)
		} else if tasks != nil && len(tasks) > 0 {
			for _, t := range tasks {
				ctx.Log.Infof("Setting up task %s", t.Name)
				for _, spec := range t.Cron {
					ctx.Log.Infof("Spec %s", spec)
					c.AddFunc(spec, func() {
						switch t.Type {
						case models.TaskCheckOpeningMovies:
							// NOTE: We emit here because our event processor do extra things to keep our tasks collection synced.
							ctx.Emitter.Emit(&contracts.EventCommandDispatched{
								TaskID:           t.ID,
								Name:             t.Name,
								Type:             t.Type,
								DispatchTime:     time.Now().UTC(),
								ExecutionTimeout: time.Second * 5,
							})

						default:
							ctx.Log.Infof("Unknown task type %s", t.Type)
						}
					})
				}
				ctx.Log.Infof("Task %s setup completed!", t.Name)
			}

			c.Start()
		}
	}

	// Wait for a signal
	sig := <-sigCh
	ctx.Log.WithField("signal", sig).Info("Signal received. Shutting down.")

	c.Stop()
}
