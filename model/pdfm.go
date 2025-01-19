package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PdfmUsers struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password" json:"password"`
	IsSupport     bool               `bson:"isSupport" json:"isSupport"`
	LastMergeTime time.Time          `bson:"lastMergeTime" json:"lastMergeTime"`
	MergeCount    int                `bson:"mergeCount" json:"mergeCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Invoice struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Amount        int                `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"` // e.g., "Paid", "Pending", "Failed"
	Details       string             `bson:"details" json:"details"`
	PaymentMethod string             `bson:"paymentMethod" json:"paymentMethod"` // e.g., "QRIS"
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

type Token struct {
	Token     string    `bson:"token"`     // Token unik
	Email     string    `bson:"email"`     // Email pengguna
	ExpiresAt time.Time `bson:"expiresAt"` // Waktu kedaluwarsa
}
