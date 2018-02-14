package service

import (
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
)

func Test_Service_New(t *testing.T) {
	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config is invalid.
		{
			config: func() Config {
				return Config{}
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test a production-like config is valid.
		{
			config: func() Config {
				config := Config{}

				config.Logger = microloggertest.New()

				config.Flag = flag.New()
				config.Viper = viper.New()

				config.Description = "test"
				config.GitCommit = "test"
				config.ProjectName = "test"
				config.Source = "test"

				config.Viper.Set(config.Flag.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				config.Viper.Set(config.Flag.Service.Kubernetes.InCluster, "false")

				return config
			},
			expectedErrorHandler: nil,
		},
	}

	for index, test := range tests {
		_, err := New(test.config())

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned: %#v", index, err)
			}
			if !test.expectedErrorHandler(err) {
				t.Fatalf("%v: incorrect error returned: %#v", index, err)
			}
		}
	}
}
