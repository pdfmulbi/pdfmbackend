package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetNotifications retrieves all notifications for the authenticated user
// GetNotifications godoc
// @Summary Lihat Notifikasi
// @Description Mengambil daftar notifikasi milik user, diurutkan dari yang terbaru
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} model.NotificationResponse
// @Failure 401 {object} model.NotificationActionResponse
// @Router /pdfm/notifications [get]
// @Security BearerAuth
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user from token
	userID, err := GetUserIDFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		// PERBAIKAN: Gunakan struct Response
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 401, Message: "Unauthorized: " + err.Error()})
		return
	}

	// Get MongoDB collection
	collection := GetMongoCollection("notifications")

	// Find all notifications for this user, sorted by created_at descending
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50)
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID}, opts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 500, Message: "Failed to fetch notifications"})
		return
	}
	defer cursor.Close(context.Background())

	var notifications []model.Notification
	if err = cursor.All(context.Background(), &notifications); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 500, Message: "Failed to parse notifications"})
		return
	}

	// Return empty array if no notifications found
	if notifications == nil {
		notifications = []model.Notification{}
	}

	// Response ini sudah benar pakai struct
	json.NewEncoder(w).Encode(model.NotificationResponse{
		Status:        http.StatusOK,
		Message:       "Notifications retrieved successfully",
		Notifications: notifications,
	})
}

// AddNotification creates a new notification for the authenticated user
// AddNotification godoc
// @Summary Buat Notifikasi (System)
// @Description Menambahkan notifikasi baru (biasanya trigger dari sistem)
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body model.NotificationRequest true "Payload Notifikasi"
// @Success 201 {object} model.NotificationActionResponse
// @Failure 400 {object} model.NotificationActionResponse
// @Router /pdfm/notifications [post]
// @Security BearerAuth
func AddNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := GetUserIDFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 401, Message: "Unauthorized"})
		return
	}

	var req model.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 400, Message: "Invalid request body"})
		return
	}

	if req.Type == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 400, Message: "Type and message are required"})
		return
	}

	notification := model.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      req.Type,
		Message:   req.Message,
		Icon:      req.Icon,
		IsRead:    false,
		CreatedAt: time.Now(),
		FileName:  req.FileName,
	}

	collection := GetMongoCollection("notifications")
	_, err = collection.InsertOne(context.Background(), notification)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 500, Message: "Failed to create notification"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	// PERBAIKAN: Gunakan NotificationActionResponse
	json.NewEncoder(w).Encode(model.NotificationActionResponse{
		Status:  http.StatusCreated,
		Message: "Notification created successfully",
		ID:      notification.ID.Hex(),
	})
}

// MarkAllAsRead marks all notifications as read for the authenticated user
// MarkAllAsRead godoc
// @Summary Tandai Semua Dibaca
// @Description Mengubah status semua notifikasi user menjadi 'read' (sudah dibaca)
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} model.NotificationActionResponse
// @Router /pdfm/notifications/read [put]
// @Security BearerAuth
func MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := GetUserIDFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 401, Message: "Unauthorized"})
		return
	}

	collection := GetMongoCollection("notifications")
	_, err = collection.UpdateMany(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 500, Message: "Failed to mark as read"})
		return
	}

	// PERBAIKAN: Gunakan NotificationActionResponse
	json.NewEncoder(w).Encode(model.NotificationActionResponse{
		Status:  http.StatusOK,
		Message: "All notifications marked as read",
	})
}

// ClearNotifications deletes all notifications for the authenticated user
// ClearNotifications godoc
// @Summary Hapus Semua Notifikasi
// @Description Membersihkan seluruh riwayat notifikasi milik user
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} model.NotificationActionResponse
// @Router /pdfm/notifications [delete]
// @Security BearerAuth
func ClearNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := GetUserIDFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 401, Message: "Unauthorized"})
		return
	}

	collection := GetMongoCollection("notifications")
	_, err = collection.DeleteMany(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.NotificationActionResponse{Status: 500, Message: "Failed to clear notifications"})
		return
	}

	// PERBAIKAN: Gunakan NotificationActionResponse
	json.NewEncoder(w).Encode(model.NotificationActionResponse{
		Status:  http.StatusOK,
		Message: "All notifications cleared",
	})
}

// GetUserIDFromToken extracts user ID from the Authorization header token
// Uses the existing Bearer token authentication system from pdfm.go
func GetUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return primitive.NilObjectID, errors.New("missing token")
	}

	// Validate Bearer token format
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || !strings.HasPrefix(authHeader, bearerPrefix) {
		return primitive.NilObjectID, errors.New("invalid token format")
	}
	token := authHeader[len(bearerPrefix):]

	// Validate token in database
	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid token")
	}

	// Check if token is expired
	if tokenData.ExpiresAt.Before(time.Now()) {
		return primitive.NilObjectID, errors.New("token expired")
	}

	// Get user from email in token
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		return primitive.NilObjectID, errors.New("user not found")
	}

	return user.ID, nil
}

// GetMongoCollection returns a MongoDB collection
func GetMongoCollection(collectionName string) *mongo.Collection {
	return config.Mongoconn.Collection(collectionName)
}

// CreateNotificationForUser creates a notification for a specific user by their ID
// This is useful for internal services to create notifications
func CreateNotificationForUser(userID primitive.ObjectID, notifType, message, icon, fileName string) error {
	notification := model.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      notifType,
		Message:   message,
		Icon:      icon,
		IsRead:    false,
		CreatedAt: time.Now(),
		FileName:  fileName,
	}

	collection := GetMongoCollection("notifications")
	_, err := collection.InsertOne(context.Background(), notification)
	return err
}