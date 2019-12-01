package usecases

import (
	"bytes"
	"image"

	"github.com/Dimitriy14/image-resizing/models"
	"github.com/disintegration/imaging"
)

//go:generate mockgen -destination=../mocks/mock-resizer.go -mock_names=ImageResizer=MockResizer -package=mocks github.com/Dimitriy14/image-resizing/usecases ImageResizer
type ImageResizer interface {
	Resize(imageContent []byte, params models.ResizeParams) ([]byte, error)
}

func NewImageResizer() ImageResizer {
	return &resiserImpl{}
}

type resiserImpl struct {
}

func (r *resiserImpl) Resize(imageContent []byte, params models.ResizeParams) ([]byte, error) {
	var (
		imageReader = bytes.NewReader(imageContent)
	)

	img, formatName, err := image.Decode(imageReader)
	if err != nil {
		return nil, err
	}

	format, err := imaging.FormatFromExtension(formatName)
	if err != nil {
		return nil, err
	}

	i := imaging.Resize(img, int(params.With), int(params.Height), imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err = imaging.Encode(buf, i, format); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
