package logger

import (
	"os"

	"github.com/Dimitriy14/image-resizing/config"
	"github.com/sirupsen/logrus"
)

var Log Logger

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// Load loads logger
func Load() error {
	var (
		output  = os.Stdout
		logFile = config.Conf.LogFile
	)

	if logFile != "" {
		logFile, err := os.Create(logFile)
		if err != nil {
			return err
		}
		output = logFile
	}

	logLvl, err := logrus.ParseLevel(config.Conf.LogLevel)
	if err != nil {
		return err
	}

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(output)
	log.SetLevel(logLvl)

	Log = log

	return nil
}
