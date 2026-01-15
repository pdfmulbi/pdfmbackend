package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PdfmUsers struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	Password     string             `bson:"password" json:"password"`
	IsAdmin      bool               `bson:"isAdmin" json:"isAdmin"`
	IsSupport    bool               `bson:"isSupport" json:"isSupport"`
	ProfilePhoto string             `bson:"profilePhoto,omitempty" json:"profilePhoto,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
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

type Feedback struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id,omitempty" json:"user_id"` // Opsional
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Message   string             `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type LoginLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	IPAddress string             `bson:"ip_address" json:"ip_address"`
	UserAgent string             `bson:"user_agent" json:"user_agent"`
	LoginAt   time.Time          `bson:"login_at" json:"login_at"`
}

type Token struct {
	Token     string    `bson:"token"`
	Email     string    `bson:"email"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

//================================================//
// 					Untuk SWAGGER 
//================================================//
type RegisterInput struct {
    Name     string `json:"name" example:"pipo"`
    Email    string `json:"email" example:"example@gmail.com"`
    Password string `json:"password" example:"rahasia123"`
}

type LoginInput struct {
    Email    string `json:"email" example:"example@gmail.com"`
    Password string `json:"password" example:"rahasia123"`
}

type PaymentInput struct {
    Name   string `json:"name" example:"pipo"`
    Amount int    `json:"amount" example:"50000"`
}

type UpdateUserInput struct {
    ID        string `json:"id" example:"65a..."`
    Name      string `json:"name" example:"pipo"`
    Email     string `json:"email" example:"example@gmail.com"`
    Password  string `json:"password" example:"passbaru"`
    IsSupport bool   `json:"isSupport" example:"false"`
}

type DeleteUserInput struct {
    ID string `json:"id" example:"65a423..."`
}

type UploadProfilePhotoInput struct {
    ProfilePhoto string `json:"profilePhoto" example:"data:image/png;base64,iVBORw0KGgo..."`
}

type FeedbackInput struct {
    Name    string `json:"name" example:"Budi"`
    Email   string `json:"email" example:"budi@gmail.com"`
    Message string `json:"message" example:"Aplikasi ini sangat mantap!"`
    Rating  int    `json:"rating" example:"5"`
}