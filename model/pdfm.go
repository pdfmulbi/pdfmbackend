package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PdfmUsers struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	IsAdmin   bool               `bson:"isAdmin" json:"isAdmin"`
	IsSupport bool               `bson:"isSupport" json:"isSupport"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Invoice struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Amount        int                `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"` //"Paid", "Pending", "Failed"
	Details       string             `bson:"details" json:"details"`
	PaymentMethod string             `bson:"paymentMethod" json:"paymentMethod"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

type Token struct {
	Token     string    `bson:"token"`
	Email     string    `bson:"email"`
	ExpiresAt time.Time `bson:"expiresAt"`
}