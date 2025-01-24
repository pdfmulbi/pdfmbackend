package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocroot/controller"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper function to execute a request
func executeRequest(t *testing.T, handler http.HandlerFunc, method, url string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// Test RegisterHandler
func TestRegisterHandler(t *testing.T) {
	data := model.PdfmUsers{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to encode JSON: %v", err)
	}

	rr := executeRequest(t, controller.RegisterHandler, http.MethodPost, "/pdfm/register", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if response["message"] != "Registrasi berhasil" {
		t.Errorf("Expected message 'Registrasi berhasil', got '%s'", response["message"])
	}
}

// Test Login
func TestGetUser(t *testing.T) {
	data := model.PdfmUsers{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.GetUser, http.MethodPost, "/pdfm/login", body)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

// Test Logout
func TestLogoutHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/pdfm/logout", nil)
	req.Header.Set("Authorization", "Bearer sample_token")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.LogoutHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if response["message"] != "Logout berhasil" {
		t.Errorf("Unexpected message: %s", response["message"])
	}
}

// Test GetUsers
func TestGetUsers(t *testing.T) {
	rr := executeRequest(t, controller.GetUsers, http.MethodGet, "/pdfm/get/users", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

// Test GetOneUser
func TestGetOneUser(t *testing.T) {
	url := "/pdfm/getone/users?name=Test+User"
	rr := executeRequest(t, controller.GetOneUser, http.MethodGet, url, nil)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status OK or NotFound, got %v", rr.Code)
	}
}

// Test GetOneUserAdmin
func TestGetOneUserAdmin(t *testing.T) {
	url := "/pdfm/getoneadmin/users?id=678baf981b52a5a0c34d16be"
	rr := executeRequest(t, controller.GetOneUserAdmin, http.MethodGet, url, nil)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status OK or NotFound, got %v", rr.Code)
	}
}

// Test CreateUser
func TestCreateUser(t *testing.T) {
	data := model.PdfmUsers{
		Email: "newuser@example.com",
		Name:  "New User_Lah",
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.CreateUser, http.MethodPost, "/pdfm/create/users", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
	existingID := "678be27e03a8b7bbb3ee3077" // Ganti dengan ID valid di database
	data := map[string]interface{}{
		"id":        existingID,
		"name":      "Updated Name",
		"email":     "updated@example.com",
		"password":  "newpassword123",
		"isSupport": true,
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.UpdateUser, http.MethodPut, "/pdfm/update/users", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v. Response: %s", rr.Code, rr.Body.String())
	}
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	data := map[string]interface{}{
		"id": primitive.NewObjectID().Hex(),
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.DeleteUser, http.MethodDelete, "/pdfm/delete/users", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

func TestConfirmPaymentHandler(t *testing.T) {
	data := map[string]interface{}{
		"name":   "Test User",
		"amount": 100000,
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.ConfirmPaymentHandler, http.MethodPost, "/pdfm/confirm/payment", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v. Response: %s", rr.Code, rr.Body.String())
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if response["message"] != "Payment confirmed, user updated, and invoice created successfully" {
		t.Errorf("Unexpected message: %s", response["message"])
	}
}

func TestGetInvoicesHandler(t *testing.T) {
	rr := executeRequest(t, controller.GetInvoicesHandler, http.MethodGet, "/pdfm/get/invoices", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var invoices []model.Invoice
	if err := json.Unmarshal(rr.Body.Bytes(), &invoices); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(invoices) == 0 {
		t.Log("No invoices found (this may be expected if the database is empty).")
	}
}
