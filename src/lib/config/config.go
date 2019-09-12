package config

import (
	"log"

	"github.com/dsbezerra/amenic/src/lib/env"
)

// DBType ...
type DBType string

const (
	// MongoDB name
	MongoDB DBType = "MongoDB"

	// DefaultDBConnection is the default URI used to connect to the database
	DefaultDBConnection = "mongodb://localhost/amenic"

	// DefaultDBLoggingConnection is the default URI used to connect to the logging database
	DefaultDBLoggingConnection = "mongodb://localhost/amenic-logs"

	DefaultMessageBrokerType = "amqp"
	DefaultAMQPMessageBroker = "amqp://guest:guest@localhost:5672"
)

// ServiceConfig ...
type ServiceConfig struct {
	DBType              DBType `json:"database_type"`
	DBConnection        string `json:"database_connection"`
	DBLoggingType       DBType `json:"logging_database_type"`
	DBLoggingConnection string `json:"logging_database_connection"`
	RESTEndpoint        string `json:"rest_endpoint"`
	// RESTTLSEndpoint    string
	IsProduction           bool   `json:"is_production"`
	MessageBrokerType      string `json:"message_broker_type"`
	AMQPMessageBroker      string `json:"amqp_message_broker"`
	ImageServiceConnection string `json:"imageservice_connection"`
}

// LoadConfiguration initializes the required configuration
// for this service
func LoadConfiguration() (*ServiceConfig, error) {
	// Start config with default values
	config := &ServiceConfig{
		DBType:              MongoDB,
		DBConnection:        DefaultDBConnection,
		DBLoggingType:       MongoDB,
		DBLoggingConnection: DefaultDBLoggingConnection,
		RESTEndpoint:        "localhost:8000", // Default REST endpoint.
		IsProduction:        false,
		MessageBrokerType:   DefaultMessageBrokerType,
		AMQPMessageBroker:   DefaultAMQPMessageBroker,
	}

	// This loads all environment variables defined in the .env file
	vars, err := env.LoadEnv()
	if err != nil {
		return config, err
	}

	connection := vars["DATABASE"]
	if connection == "" {
		log.Fatal("Main database URI is missing")
	}

	config.AMQPMessageBroker = vars["AMQP_URL"]
	config.RESTEndpoint = vars["LISTEN_URL"]
	// NOTE: If we use other database type, update this.
	config.DBConnection = connection
	config.DBLoggingConnection = vars["LOG_DATABASE"]
	config.IsProduction = vars["MODE"] == "release"
	config.ImageServiceConnection = vars["CLOUDINARY_URL"]

	return config, err
}
