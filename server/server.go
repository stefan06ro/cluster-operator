package server

import (
	"context"
	"net/http"
	"sync"

	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/cluster-operator/service"
)

// Config represents the configuration used to construct server object.
type Config struct {
	// Dependencies
	Service *service.Service

	//Settings
	MicroServerConfig microserver.Config
}

// DefaultConfig provides a default configuration to create a new server object.
func DefaultConfig() Config {
	return Config{
		// Dependencies
		Service: nil,

		// Settings
		MicroServerConfig: microserver.DefaultConfig(),
	}
}

// New creates a new server object with given configuration.
func New(config Config) (microserver.Server, error) {
	newServer := &server{
		// Dependencies
		logger: config.MicroServerConfig.Logger,

		// Internals
		bootOnce:     sync.Once{},
		config:       config.MicroServerConfig,
		serviceName:  config.MicroServerConfig.ServiceName,
		shutdownOnce: sync.Once{},
	}

	// Apply internals to the micro server config.
	newServer.config.Endpoints = []microserver.Endpoint{}
	newServer.config.ErrorEncoder = newServer.newErrorEncoder()

	return newServer, nil
}

type server struct {
	// Dependencies
	logger micrologger.Logger

	// Internals
	bootOnce     sync.Once
	config       microserver.Config
	serviceName  string
	shutdownOnce sync.Once
}

func (s *server) Boot() {
	s.bootOnce.Do(func() {
		// Insert here custom boot logic for server/endpoint/middleware if needed.
	})
}

func (s *server) Config() microserver.Config {
	return s.config
}

func (s *server) Shutdown() {
	s.shutdownOnce.Do(func() {
		// Insert here custom shutdown logic for server/endpoint/middleware if needed.
	})
}

func (s *server) newErrorEncoder() kithttp.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		rErr := err.(microserver.ResponseError)
		uErr := rErr.Underlying()

		rErr.SetCode(microserver.CodeInternalError)
		rErr.SetMessage(uErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
