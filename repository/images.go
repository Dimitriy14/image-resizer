package repository

import (
	"github.com/Dimitriy14/image-resizing/models"
	"github.com/google/uuid"
)

func (r *repoImpl) GetAllImages(userID uuid.UUID) ([]models.Images, error) {
	var images []models.Images
	err := r.db.Session.Where("user_id = ?", userID).Find(&images).Error
	return images, err
}

func (r *repoImpl) GetImageByID(userID, imageID uuid.UUID) (models.Images, error) {
	var image models.Images
	err := r.db.Session.Where("user_id = ? AND id= ?", userID, imageID).Find(&image).Error
	return image, err
}

func (r *repoImpl) SaveImage(img models.Images) (models.Images, error) {
	err := r.db.Session.Save(&img).Error
	return img, err
}

func (r *repoImpl) UpdateImage(img models.Images) (models.Images, error) {
	err := r.db.Session.Model(&models.Images{}).Update(&img).Error
	return img, err
}
