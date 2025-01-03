package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Untuk Login
type PdfmAdminRegistration struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Fullname        string             `json:"fullname" bson:"fullname" validate:"required"`
	Email           string             `json:"email" bson:"email" validate:"required,email"`
	Password        string             `json:"password" bson:"password" validate:"required,min=8"`
	ConfirmPassword string             `json:"confirm_password" bson:"confirm_password" validate:"required,eqfield=Password"`
}

type PdfmUser struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"email" validate:"required,email"`
	Password string             `json:"password" bson:"password" validate:"required,min=8"`
}
