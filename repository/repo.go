package repository

import (
	"github.com/Dimitriy14/image-resizing/clients/postgres"
	"github.com/Dimitriy14/image-resizing/models"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=../mocks/mock-repo.go -mock_names=Repository=MockRepository -package=mocks github.com/Dimitriy14/image-resizing/repository Repository
type Repository interface {
	GetAllImages(userID uuid.UUID) ([]models.Images, error)
	GetImageByID(userID, imageID uuid.UUID) (models.Images, error)
	SaveImage(models.Images) (models.Images, error)
	UpdateImage(models.Images) (models.Images, error)
}

type repoImpl struct {
	db *postgres.PGClient
}

func NewRepository(client *postgres.PGClient) Repository {
	return &repoImpl{db: client}
}
