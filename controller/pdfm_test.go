package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/gofiber/fiber/v2"
	"testing"

	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// 1. AUTHENTICATION & SESSION
// ==========================================

func TestRegisterHandler(t *testing.T) {
	h := &UserHandler{} // Instansiasi Objek (OOP Style)
	data := model.RegisterInput{
		Name:     "Tester Akun",
		Email:    "tester_akun@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/pdfm/register", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.RegisterHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 200 or 409, got %v", resp.StatusCode)
	}
}

func TestGetUser(t *testing.T) {
	h := &UserHandler{}
	data := model.LoginInput{
		Email:    "tester_history@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/pdfm/login", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetUser)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 200 or 401, got %v", resp.StatusCode)
	}
}

func TestLogoutHandler(t *testing.T) {
	h := &UserHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("POST", "/pdfm/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.LogoutHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 2. USER PROFILE & PHOTO
// ==========================================

func TestGetOneUser(t *testing.T) {
	h := &UserHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/getone/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetOneUser)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestUploadProfilePhotoHandler(t *testing.T) {
	h := &UserHandler{}
	token := setupMockToken(t)
	payload := model.UploadProfilePhotoInput{
		ProfilePhoto: "data:image/png;base64,iVBORw0KGgo...",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/pdfm/profile/photo", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.UploadProfilePhotoHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetProfilePhotoHandler(t *testing.T) {
	h := &UserHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/profile/photo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetProfilePhotoHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 3. ADMIN USER MANAGEMENT
// ==========================================

func TestGetUsers(t *testing.T) {
	h := &UserHandler{}
	req, _ := http.NewRequest("GET", "/pdfm/get/users", nil)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetUsers)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetOneUserAdmin(t *testing.T) {
	h := &UserHandler{}
	req, _ := http.NewRequest("GET", "/pdfm/getoneadmin/users?name=Tester", nil)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetOneUserAdmin)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %v", resp.StatusCode)
	}
}

func TestCreateUser(t *testing.T) {
	h := &UserHandler{}
	payload := model.RegisterInput{Name: "Manual Admin", Email: "admin_create@ex.com", Password: "123"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/pdfm/create/users", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.CreateUser)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 200 or 409, got %v", resp.StatusCode)
	}
}

func TestUpdateUser(t *testing.T) {
	h := &UserHandler{}
	payload := model.UpdateUserInput{
		ID:    primitive.NewObjectID().Hex(),
		Name:  "Update Tester",
		Email: "update@ex.com",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", "/pdfm/update/users", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.UpdateUser)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %v", resp.StatusCode)
	}
}

func TestDeleteUser(t *testing.T) {
	h := &UserHandler{}
	payload := model.DeleteUserInput{ID: primitive.NewObjectID().Hex()}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("DELETE", "/pdfm/delete/users", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.DeleteUser)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 200 or 400, got %v", resp.StatusCode)
	}
}

// ==========================================
// 4. PAYMENT & INVOICES
// ==========================================

func TestConfirmPaymentHandler(t *testing.T) {
	h := &UserHandler{}
	payload := model.PaymentInput{Name: "User Tester History", Amount: 75000}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/pdfm/payment", bytes.NewBuffer(body))
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.ConfirmPaymentHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %v", resp.StatusCode)
	}
}

func TestGetInvoicesHandler(t *testing.T) {
	h := &UserHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/invoices", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetInvoicesHandler)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}
