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
	Type     string `json:"type"`
	Message  string `json:"message"`
	Icon     string `json:"icon"`
	FileName string `json:"file_name"`
}

// NotificationResponse is the response for getting notifications
type NotificationResponse struct {
	Status        int            `json:"status"`
	Message       string         `json:"message"`
	Notifications []Notification `json:"notifications"`
}
