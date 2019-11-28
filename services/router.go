package services

import (
	"net/http"

	"github.com/rs/cors"
	"github.com/urfave/negroni"

	"github.com/Dimitriy14/image-resizing/clients/bucket"
	"github.com/Dimitriy14/image-resizing/clients/postgres"
	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/middlewares"
	"github.com/Dimitriy14/image-resizing/repository"
	"github.com/Dimitriy14/image-resizing/services/images"
	"github.com/Dimitriy14/image-resizing/storage/aws"
	"github.com/Dimitriy14/image-resizing/usecases"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	repo := repository.NewRepository(postgres.Client)
	uploader := aws.NewStorage(bucket.Client)
	resizer := usecases.NewImageResizer()
	imageService := images.NewService(logger.Log, uploader, repo, resizer)

	router := mux.NewRouter().StrictSlash(true).PathPrefix(config.Conf.BasePath).Subrouter()
	v1router := router.PathPrefix("/v1").Subrouter()

	v1router.Use(middlewares.CheckUser)
	v1router.HandleFunc("/images", imageService.GetAllImages).Methods(http.MethodGet)
	v1router.HandleFunc("/images", imageService.ResizeNewImage).Methods(http.MethodPost)
	v1router.HandleFunc("/images/{id}", imageService.ResizeExistedImage).Methods(http.MethodPut)

	var corsRouter = mux.NewRouter()
	{
		corsRouter.PathPrefix(config.Conf.BasePath).Handler(negroni.New(
			cors.New(cors.Options{
				AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
			}),
			negroni.Wrap(router),
		))
	}

	return corsRouter
}
