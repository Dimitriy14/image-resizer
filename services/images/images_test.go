package images

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Dimitriy14/image-resizing/models"

	"github.com/Dimitriy14/image-resizing/services/common"

	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	assert.Empty(t, NewService(nil, nil, nil, nil), "NewService shouldn't be empty")
}

func TestServiceImpl_GetAllImages(t *testing.T) {
	log := logger.NewMokLogger()
	logger.Log = log

	testCases := []struct {
		name         string
		expCode      int
		getImagesErr error
	}{
		{
			name:    "Good case",
			expCode: http.StatusOK,
		},
		{
			name:         "Getting images error case",
			expCode:      http.StatusInternalServerError,
			getImagesErr: errors.New("ERROR"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bucket := mocks.NewMockStorage(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			resizer := mocks.NewMockResizer(ctrl)

			repo.EXPECT().GetAllImages(gomock.Any()).Return(nil, tc.getImagesErr)

			s := NewService(log, bucket, repo, resizer)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://foo", nil)
			s.GetAllImages(rr, req)

			resp := rr.Result()

			assert.Equal(t, tc.expCode, resp.StatusCode, "unexpected status code")
		})
	}
}

func TestServiceImpl_ResizeNewImage(t *testing.T) {
	log := logger.NewMokLogger()
	logger.Log = log

	testCases := []struct {
		name          string
		width         string
		height        string
		expCode       int
		saveImagesErr error
		resizeErr     error
		uploadErr     error
	}{
		{
			name:    "Good case",
			width:   "100",
			height:  "100",
			expCode: http.StatusCreated,
		},
		{
			name:    "Invalid width",
			width:   "-1",
			height:  "100",
			expCode: http.StatusBadRequest,
		},
		{
			name:    "Invalid height",
			width:   "100",
			height:  "-1",
			expCode: http.StatusBadRequest,
		},
		{
			name:      "Resize error case",
			width:     "100",
			height:    "100",
			expCode:   http.StatusInternalServerError,
			resizeErr: errors.New("RESIZE ERROR"),
		},
		{
			name:      "Upload error case",
			width:     "100",
			height:    "100",
			expCode:   http.StatusInternalServerError,
			uploadErr: errors.New("UPLOAD ERROR"),
		},
		{
			name:          "Saving error case",
			width:         "100",
			height:        "100",
			expCode:       http.StatusInternalServerError,
			saveImagesErr: errors.New("SAVING ERROR"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bucket := mocks.NewMockStorage(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			resizer := mocks.NewMockResizer(ctrl)

			repo.EXPECT().SaveImage(gomock.Any()).Return(models.Images{}, tc.saveImagesErr).AnyTimes()
			bucket.EXPECT().UploadWithOriginal(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", tc.uploadErr).AnyTimes()
			resizer.EXPECT().Resize(gomock.Any(), gomock.Any()).Return([]byte{}, tc.resizeErr).AnyTimes()

			s := NewService(log, bucket, repo, resizer)

			req := newMultipartRequest(t, tc.width, tc.height)
			rr := httptest.NewRecorder()
			s.ResizeNewImage(rr, req)

			resp := rr.Result()

			assert.Equal(t, tc.expCode, resp.StatusCode, "unexpected status code")
		})
	}
}

func newMultipartRequest(t *testing.T, width, height string) *http.Request {
	buf := bytes.NewBuffer([]byte{})
	mw := multipart.NewWriter(buf)
	_, err := mw.CreateFormFile(image, "image.jpg")
	if err != nil {
		t.Errorf("cannot create form file: %s", err)
	}

	err = mw.WriteField("width", width)
	if err != nil {
		t.Errorf("cannot create form file: %s", err)
	}
	err = mw.WriteField("height", height)
	if err != nil {
		t.Errorf("cannot create form file: %s", err)
	}

	common.CloseWithErrCheck(mw, "multipart form")

	req := httptest.NewRequest(http.MethodPost, "http://foo", buf)
	req.Header.Add("Content-type", mw.FormDataContentType())
	return req
}

func TestServiceImpl_ResizeExistedImage(t *testing.T) {
	log := logger.NewMokLogger()
	logger.Log = log
	imgID := uuid.New()

	type errorCases struct {
		getImageErr error
		updateErr   error
		resizeErr   error
		uploadErr   error
		downloadErr error
		deleteErr   error
	}

	testCases := []struct {
		name    string
		body    []byte
		expCode int
		id      string
		errors  errorCases
	}{
		{
			name:    "Good case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusOK,
			id:      imgID.String(),
		},
		{
			name:    "Invalid ID case",
			expCode: http.StatusBadRequest,
			id:      "invalid id",
		},
		{
			name:    "Invalid ID case",
			body:    []byte("invalid body"),
			expCode: http.StatusBadRequest,
			id:      imgID.String(),
		},
		{
			name:    "Not found case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusNotFound,
			id:      imgID.String(),
			errors: errorCases{
				getImageErr: gorm.ErrRecordNotFound,
			},
		},
		{
			name:    "Getting image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusInternalServerError,
			id:      imgID.String(),
			errors: errorCases{
				getImageErr: errors.New("ERROR"),
			},
		},
		{
			name:    "Downloading image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusInternalServerError,
			id:      imgID.String(),
			errors: errorCases{
				downloadErr: errors.New("ERROR"),
			},
		},
		{
			name:    "Resizing image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusInternalServerError,
			id:      imgID.String(),
			errors: errorCases{
				resizeErr: errors.New("ERROR"),
			},
		},
		{
			name:    "Uploading image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusInternalServerError,
			id:      imgID.String(),
			errors: errorCases{
				uploadErr: errors.New("ERROR"),
			},
		},
		{
			name:    "Update image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusInternalServerError,
			id:      imgID.String(),
			errors: errorCases{
				updateErr: errors.New("ERROR"),
			},
		},
		{
			name:    "Deleting image error case",
			body:    []byte(`{"width":100, "height":100}`),
			expCode: http.StatusOK,
			id:      imgID.String(),
			errors: errorCases{
				deleteErr: errors.New("ERROR"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			bucket := mocks.NewMockStorage(ctrl)
			repo := mocks.NewMockRepository(ctrl)
			resizer := mocks.NewMockResizer(ctrl)

			repo.EXPECT().UpdateImage(gomock.Any()).Return(models.Images{ID: imgID}, tc.errors.updateErr).AnyTimes()
			repo.EXPECT().GetImageByID(gomock.Any(), imgID).Return(models.Images{ID: imgID}, tc.errors.getImageErr).AnyTimes()
			bucket.EXPECT().Upload(gomock.Any(), gomock.Any()).Return("", tc.errors.uploadErr).AnyTimes()
			bucket.EXPECT().Download(gomock.Any()).Return([]byte{}, tc.errors.downloadErr).AnyTimes()
			bucket.EXPECT().DeleteImage(gomock.Any()).Return(tc.errors.deleteErr).AnyTimes()
			resizer.EXPECT().Resize(gomock.Any(), gomock.Any()).Return([]byte{}, tc.errors.resizeErr).AnyTimes()

			s := NewService(log, bucket, repo, resizer)

			req := httptest.NewRequest(http.MethodPost, "http://foo", bytes.NewBuffer(tc.body))
			req = mux.SetURLVars(req, map[string]string{
				"id": tc.id,
			})
			rr := httptest.NewRecorder()
			s.ResizeExistedImage(rr, req)

			resp := rr.Result()

			assert.Equal(t, tc.expCode, resp.StatusCode, "unexpected status code")
		})
	}
}
