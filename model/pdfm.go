package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Untuk Register
type PdfmRegistration struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name          	string             `bson:"name" json:"name"`
	Email           string             `json:"email" bson:"email" validate:"required,email"`
	Password        string             `json:"password" bson:"password" validate:"required,min=8"`
	ConfirmPassword string             `json:"confirm_password" bson:"confirm_password" validate:"required,eqfield=Password"`
}

type PdfmUsers struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password" json:"password"`
	IsPremium     bool               `bson:"isPremium" json:"isPremium"`
	LastMergeTime time.Time          `bson:"lastMergeTime" json:"lastMergeTime"`
	MergeCount    int                `bson:"mergeCount" json:"mergeCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
