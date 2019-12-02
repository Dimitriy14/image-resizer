package usecases

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/Dimitriy14/image-resizing/models"
	"github.com/disintegration/imaging"

	"github.com/stretchr/testify/assert"
)

func TestNewImageResizer(t *testing.T) {
	assert.NotNil(t, NewImageResizer(), "NewImageResizer should not be nil")
}

func TestResiserImpl_Resize(t *testing.T) {
	img := imaging.New(200, 200, color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 1,
	})

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, img, imaging.Format(1)); err != nil {
		t.Errorf("Cannot encode img: %s", err)
	}

	testCases := []struct {
		name         string
		params       models.ResizeParams
		imageContent []byte
		wantError    bool
	}{
		{
			name: "Good case",
			params: models.ResizeParams{
				With:   100,
				Height: 200,
			},
			imageContent: buf.Bytes(),
		},
		{
			name: "Nil image case",
			params: models.ResizeParams{
				With:   100,
				Height: 200,
			},
			imageContent: nil,
			wantError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewImageResizer()

			resized, err := s.Resize(tc.imageContent, tc.params)
			if err != nil {
				if tc.wantError {
					t.Skipf("Expected error: %s", err)
				}

				t.Errorf("got unexpected error: %s", err)
			}

			reader := bytes.NewReader(resized)
			image, _, err := image.DecodeConfig(reader)
			if err != nil {
				t.Fatalf("cannot decode image err: %s", err)
			}

			if (image.Width != int(tc.params.With)) || (image.Height != int(tc.params.Height)) {
				t.Fatalf("want image with width %d and height %d but got with width %d and height %d", tc.params.With, tc.params.Height, image.Width, image.Height)
			}
		})
	}

}
