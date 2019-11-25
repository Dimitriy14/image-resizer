package images

import "net/http"

type Service interface {
	GetAllImages(w http.ResponseWriter, r *http.Request)
	ResizeNewImage(w http.ResponseWriter, r *http.Request)
	ResizeExistedImage(w http.ResponseWriter, r *http.Request)
}

type serviceImpl struct {
}
