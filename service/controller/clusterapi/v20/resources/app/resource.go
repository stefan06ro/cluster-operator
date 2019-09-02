package app

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "appv19"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	Provider string
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	provider string
}

// New creates a new chartconfig service.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
