package apploader

import (
	"fmt"

	"github.com/Dimitriy14/image-resizing/clients/bucket"
	"github.com/Dimitriy14/image-resizing/clients/postgres"
	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/pkg/errors"
)

// LoaderList is a collection of Load() functions
type LoaderList []struct {
	name string
	load func() error
}

var basicLoaders = LoaderList{
	{"config", config.Load}, //config should be loaded first
	{"logger", logger.Load},
}

var clientLoaders = LoaderList{
	{"database", postgres.Load},
	{"bucket", bucket.Load},
}

func LoadApplicationServices() error {
	err := executeLoaders(basicLoaders)
	if err != nil {
		return err
	}

	err = executeLoaders(clientLoaders)
	if err != nil {
		return err
	}
	return nil
}

func executeLoaders(loaders LoaderList) error {
	for _, loader := range loaders {
		err := loader.load()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to execute %s.Load()", loader.name))
		}
	}
	return nil
}
