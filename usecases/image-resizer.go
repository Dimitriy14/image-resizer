package usecases

import (
	"image"

	"github.com/Dimitriy14/image-resizing/models"
)

type ImageResizer interface {
	Resize(image image.Image, params models.ResizeParams) error
}

type resiserImpl struct {
}

func (r *resiserImpl) Resize(imageContent image.Image, params models.ResizeParams) error {
	return nil
}
