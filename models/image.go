package models

import (
	"github.com/google/uuid"
)

type ResizeParams struct {
	With   uint `json:"with"`
	Height uint `json:"height"`
}

// Images contains links for original and resized image
type Images struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Original string    `json:"original"`
	Resized  string    `json:"resized"`
	UserID   uuid.UUID
}

// User describes user
type User struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"     validate:"required"`
	Password    string    `json:"password" validate:"required,password"`
	PhoneNumber string    `json:"phone"    validate:"required"`
}
