package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Dimitriy14/image-resizing/apploader"
	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/services"
	"github.com/urfave/negroni"
)

func main() {
	configPath := flag.String("-config", "config.json", "-config ")
	flag.Parse()

	if *configPath != "" {
		config.FilePath = *configPath
	}

	err := apploader.LoadApplicationServices()
	if err != nil {
		log.Fatal(err)
	}

	middlewareManager := negroni.New()
	middlewareManager.Use(negroni.NewRecovery())
	negroniLogger := negroni.NewLogger()
	negroniLogger.ALogger = logger.NewNegroniLogger(logger.Log)

	middlewareManager.Use(negroniLogger)
	middlewareManager.UseHandler(services.NewRouter())

	server := &http.Server{
		Addr:    config.Conf.ListenURL,
		Handler: middlewareManager,
	}

	logger.Log.Infof("", "Started serving at: %s", config.Conf.ListenURL)
	if err := server.ListenAndServe(); err != nil {
		logger.Log.Errorf("", "==== Resizer stopped due to error: %v", err)
	}
}
