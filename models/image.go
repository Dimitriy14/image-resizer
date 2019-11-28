package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// ResizeParams contains resized data
type ResizeParams struct {
	With   uint `json:"with"`
	Height uint `json:"height"`
}

// Images contains links for original and resized image
type Images struct {
	ID       uuid.UUID `json:"id"        gorm:"primary_key; column:id"`
	Original string    `json:"original"  gorm:"column:original"`
	Resized  string    `json:"resized"   gorm:"column:resized"`
	UserID   uuid.UUID `json:"-"         gorm:"column:user_id"`
}

func (i Images) TableName() string {
	return "images"
}

func (i *Images) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("id", uuid.New())
}
