package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/gocroot/model"
)

// ==========================================
// 1. TESTING GET NOTIFICATIONS
// ==========================================

func TestGetNotifications(t *testing.T) {
	h := &NotificationHandler{} // Instansiasi Objek (OOP)
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetNotifications)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetNotifications_Options(t *testing.T) {
	h := &NotificationHandler{}
	req, _ := http.NewRequest("OPTIONS", "/pdfm/notifications", nil)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetNotifications)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for OPTIONS, got %v", resp.StatusCode)
	}
}

// ==========================================
// 2. TESTING ADD NOTIFICATION
// ==========================================

func TestAddNotification(t *testing.T) {
	h := &NotificationHandler{}
	token := setupMockToken(t)
	payload := model.NotificationRequest{
		Type:     "info",
		Message:  "Pesan notifikasi baru",
		Icon:     "bell",
		FileName: "dokumen.pdf",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/pdfm/notifications", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.AddNotification)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201, got %v", resp.StatusCode)
	}
}

func TestAddNotification_InvalidBody(t *testing.T) {
	h := &NotificationHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("POST", "/pdfm/notifications", bytes.NewBufferString("invalid-json"))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.AddNotification)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid body, got %v", resp.StatusCode)
	}
}

// ==========================================
// 3. TESTING MARK AS READ & CLEAR
// ==========================================

func TestMarkAllAsRead(t *testing.T) {
	h := &NotificationHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("PUT", "/pdfm/notifications/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.MarkAllAsRead)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestClearNotifications(t *testing.T) {
	h := &NotificationHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("DELETE", "/pdfm/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.ClearNotifications)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 4. TESTING HELPER FUNCTIONS
// ==========================================

func TestGetUserIDFromToken_Expired(t *testing.T) {
	h := &NotificationHandler{}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Set("Authorization", "Bearer expired_token_sample")
	defer app.ReleaseCtx(c)

	// Panggil h.GetUserIDFromToken karena sekarang adalah method
	_, err := h.GetUserIDFromToken(c)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestGetMongoCollection(t *testing.T) {
	h := &NotificationHandler{}
	col := h.GetMongoCollection("notifications")
	if col.Name() != "notifications" {
		t.Errorf("Expected collection name 'notifications', got %s", col.Name())
	}
}
