package images

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/logger"
	"github.com/Dimitriy14/image-resizing/models"
	"github.com/Dimitriy14/image-resizing/repository"
	"github.com/Dimitriy14/image-resizing/services/common"
	"github.com/Dimitriy14/image-resizing/storage"
	"github.com/Dimitriy14/image-resizing/usecases"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

const (
	image        = "image"
	maxImageSize = 10 << 24 // max image size is 10MB
	formWidth    = "width"
	formHeight   = "height"
)

// Service provides functionality to retrieving and saving images
type Service interface {
	GetAllImages(w http.ResponseWriter, r *http.Request)
	ResizeNewImage(w http.ResponseWriter, r *http.Request)
	ResizeExistedImage(w http.ResponseWriter, r *http.Request)
}

// NewService creates new service
func NewService(log logger.Logger, bucket storage.Storage, repo repository.Repository, resizer usecases.ImageResizer) Service {
	return &serviceImpl{
		log:           log,
		bucket:        bucket,
		repo:          repo,
		resizer:       resizer,
		awsStorageUrl: config.Conf.AWS.ImageStorageURL,
	}
}

type serviceImpl struct {
	log           logger.Logger
	bucket        storage.Storage
	repo          repository.Repository
	resizer       usecases.ImageResizer
	awsStorageUrl string
}

func (s *serviceImpl) GetAllImages(w http.ResponseWriter, r *http.Request) {
	uid := common.GetUserIDFromCtx(r.Context())

	s.log.Debugf("Started retrieving all images for user %q", uid)

	images, err := s.repo.GetAllImages(uid)
	if err != nil {
		s.log.Errorf("cannot retrieve images from form: %s", err)
		common.SendInternalServerError(w, "cannot retrieve image", err)
		return
	}

	s.log.Debugf("Successfully retrieved all images for user %q", uid)

	common.RenderJSON(w, images)
}

func (s *serviceImpl) ResizeNewImage(w http.ResponseWriter, r *http.Request) {
	uid := common.GetUserIDFromCtx(r.Context())

	s.log.Debugf("Started resizing image for user %q", uid)

	fileContent, filename, params, err := extractFormData(r)
	if err != nil {
		s.log.Errorf("cannot extract data from request due to: %s", err)
		common.SendError(w, http.StatusBadRequest, "invalid input data", err)
		return
	}

	resizedImg, err := s.resizer.Resize(fileContent, params)
	if err != nil {
		s.log.Errorf("cannot resize image due to: %s", err)
		common.SendInternalServerError(w, "image cannot be resized", err)
		return
	}

	original, resized, err := s.bucket.UploadWithOriginal(filepath.Ext(filename), fileContent, resizedImg)
	if err != nil {
		s.log.Errorf("cannot upload images due to: %s", err)
		common.SendInternalServerError(w, "cannot upload images", err)
		return
	}

	img, err := s.repo.SaveImage(models.Images{
		ID:       uuid.New(),
		Original: original,
		Resized:  resized,
		UserID:   uid,
	})
	if err != nil {
		s.log.Errorf("cannot save images due to: %s", err)
		common.SendInternalServerError(w, "cannot save images", err)
		return
	}

	s.log.Debugf("Successfully resized and saved image for user %q", uid)

	common.RenderJSONCreated(w, &img)
}

func (s *serviceImpl) ResizeExistedImage(w http.ResponseWriter, r *http.Request) {
	var (
		uid    = common.GetUserIDFromCtx(r.Context())
		id     = mux.Vars(r)["id"]
		params models.ResizeParams
	)

	s.log.Debugf("Started resizing already existed image for user %q")

	imageID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorf("cannot parse image id (%s) from request due to: %s", id, err)
		common.SendError(w, http.StatusBadRequest, "invalid image id", err)
		return
	}

	img, err := s.repo.GetImageByID(uid, imageID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.log.Errorf("cannot find image with id (%q) for user (%q) due to: %s", imageID, uid, err)
			common.SendNotFound(w, "image id is not found: %s", err)
			return
		}

		s.log.Errorf("cannot retrieve image with id (%q) for user (%q) due to: %s", imageID, uid, err)
		common.SendInternalServerError(w, "cannot retrieve image due to db problems", err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		s.log.Errorf("cannot extract data from request due to: %s", err)
		common.SendError(w, http.StatusBadRequest, "invalid input data", err)
		return
	}

	imageContent, err := s.bucket.Download(img.Original)
	if err != nil {
		s.log.Errorf("cannot download image from s3 due to: %s", err)
		common.SendInternalServerError(w, "invalid input data", err)
		return
	}

	resizedImgContent, err := s.resizer.Resize(imageContent, params)
	if err != nil {
		s.log.Errorf("cannot resize image with id (%s) for user (%s) due to: %s", err)
		common.SendInternalServerError(w, "cannot resize this image", err)
		return
	}

	newResizeLink, err := s.bucket.Upload(filepath.Ext(img.Resized), resizedImgContent)
	if err != nil {
		s.log.Errorf("cannot resize image with id (%q) for user (%q) due to: %s", err)
		common.SendInternalServerError(w, "invalid input data", err)
		return
	}

	newImg, err := s.repo.UpdateImage(models.Images{
		ID:       imageID,
		Original: img.Original,
		Resized:  newResizeLink,
		UserID:   uid,
	})
	if err != nil {
		s.log.Errorf("cannot save images due to: %s", err)
		common.SendInternalServerError(w, "cannot save images", err)
		return
	}

	//user doesn't have to wait till his old resized image will be deleted
	go s.deleteImage(img.Resized)
	s.log.Debugf("Successfully resized and saved image for user %q", uid)

	common.RenderJSON(w, &newImg)
}

func (s *serviceImpl) deleteImage(addr string) {
	if err := s.bucket.DeleteImage(addr); err != nil {
		s.log.Errorf("got an error while deleting image from S3 with addr: %s", addr)
	}
}

func extractFormData(r *http.Request) ([]byte, string, models.ResizeParams, error) {
	err := r.ParseMultipartForm(int64(maxImageSize))
	if err != nil {
		return nil, "", models.ResizeParams{}, fmt.Errorf("parsing multipart form error: %s", err)
	}

	file, head, err := r.FormFile(image)
	if err != nil {
		return nil, "", models.ResizeParams{}, fmt.Errorf("cannot retrieve image from multipart form: %s", err)
	}

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", models.ResizeParams{}, fmt.Errorf("cannot read image content: %s", err)
	}

	width, err := strconv.ParseUint(r.FormValue(formWidth), 10, 64)
	if err != nil {
		return nil, "", models.ResizeParams{}, fmt.Errorf("converting width to uint error: %s", err)
	}

	height, err := strconv.ParseUint(r.FormValue(formHeight), 10, 64)
	if err != nil {
		return nil, "", models.ResizeParams{}, fmt.Errorf("converting height to uint error: %s", err)
	}

	return fileContent, head.Filename, models.ResizeParams{
		With:   uint(width),
		Height: uint(height),
	}, nil
}
