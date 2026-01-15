package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a user activity notification
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"`           // merge, compress, convert, summary
	Message   string             `bson:"message" json:"message"`
	Icon      string             `bson:"icon" json:"icon"`
	IsRead    bool               `bson:"is_read" json:"is_read"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	FileName  string             `bson:"file_name,omitempty" json:"file_name"`
}

// NotificationRequest is the request body for creating a notification
type NotificationRequest struct {
	Type     string `json:"type" example:"info"`
	Message  string `json:"message" example:"File PDF berhasil digabungkan"`
	Icon     string `json:"icon" example:"check-circle"`
	FileName string `json:"file_name" example:"laporan_final.pdf"`
}

// NotificationResponse is the response for getting notifications
type NotificationResponse struct {
	Status        int            `json:"status" example:"200"`
	Message       string         `json:"message" example:"Notifications retrieved successfully"`
	Notifications []Notification `json:"notifications"`
}

// Untuk Swagger
