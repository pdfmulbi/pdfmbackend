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

// Invoice struct dengan userId sebagai identifier unik
// userId adalah ObjectID dari user yang membuat invoice
type Invoice struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`               // UNIQUE identifier - ObjectID user
	UserEmail     string             `bson:"userEmail" json:"userEmail"`         // Email user untuk reference
	UserName      string             `bson:"userName" json:"userName"`           // Nama user untuk display
	Amount        int                `bson:"amount" json:"amount"`               // Nominal pembayaran
	Status        string             `bson:"status" json:"status"`               // "Paid", "Pending", "Failed"
	Details       string             `bson:"details" json:"details"`             // Keterangan
	PaymentMethod string             `bson:"paymentMethod" json:"paymentMethod"` // Metode pembayaran
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`         // Tanggal pembuatan
}

type Token struct {
	Token     string    `bson:"token"`
	Email     string    `bson:"email"`
	ExpiresAt time.Time `bson:"expiresAt"`
}
