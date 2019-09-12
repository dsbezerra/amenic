package main

import (
	helmet "github.com/danielkov/gin-helmet"
	"github.com/dsbezerra/amenic/src/adminservice/rest"
	"github.com/dsbezerra/amenic/src/lib/config"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	log = logrus.WithFields(logrus.Fields{"App": "Admin"})
)

// Context ...
type Context struct {
	Config  *config.ServiceConfig
	Data    persistence.DataAccessLayer
	Emitter messagequeue.EventEmitter
}

func main() {
	settings, err := config.LoadConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := amqp.Dial(settings.AMQPMessageBroker)
	if err != nil {
		log.Fatal(err)
	}

	eventEmitter, err := messagequeue.NewAMQPEventEmitter(conn, "events")
	if err != nil {
		log.Fatal(err)
	}

	// eventListener, err := messagequeue.NewAMQPEventListener(conn, "events", "api")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db, err := mongolayer.NewMongoDAL(settings.DBConnection)
	if err != nil {
		log.Fatal(err)
	}
	db.Setup()
	defer db.Close()

	log.Info("Database setup completed!")

	// Initialize app context
	ctx := &Context{
		Config:  settings,
		Data:    db,
		Emitter: eventEmitter,
	}
	router := ctx.buildRouter()

	// Serve API
	rest.ServeAPI(router, db, eventEmitter)
	router.Run(settings.RESTEndpoint)
}

// BuildRouter ...
func (ctx *Context) buildRouter() *gin.Engine {
	if ctx.Config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SecureJsonPrefix(")]}',\n")
	r.Use(helmet.Default())
	r.Use(
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	return r
}
