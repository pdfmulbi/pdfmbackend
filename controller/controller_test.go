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

// Test GetUser
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

// Test MergePDFHandler
func TestMergePDFHandler(t *testing.T) {
	url := "/pdfm/merge?user_id=" + primitive.NewObjectID().Hex()
	rr := executeRequest(t, controller.MergePDFHandler, http.MethodPost, url, nil)

	if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
		t.Errorf("Expected status OK or Forbidden, got %v", rr.Code)
	}
}

// Test GetUsers
func TestGetUsers(t *testing.T) {
	rr := executeRequest(t, controller.GetUsers, http.MethodGet, "/pdfm/users", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

// Test GetOneUser
func TestGetOneUser(t *testing.T) {
	url := "/pdfm/users/details?name=Test+User"
	rr := executeRequest(t, controller.GetOneUser, http.MethodGet, url, nil)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status OK or NotFound, got %v", rr.Code)
	}
}

// Test CreateUser
func TestCreateUser(t *testing.T) {
	data := model.PdfmUsers{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.CreateUser, http.MethodPost, "/pdfm/users", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
    existingID := "6788b3e5b1e4cb696abc705c" // Ganti dengan ID valid di database
    data := map[string]interface{}{
        "id":        existingID,
        "name":      "Updated Name",
        "email":     "updated@example.com",
        "password":  "newpassword123",
        "isSupport": true,
    }
    body, _ := json.Marshal(data)

    rr := executeRequest(t, controller.UpdateUser, http.MethodPut, "/pdfm/users", body)

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

	rr := executeRequest(t, controller.DeleteUser, http.MethodDelete, "/pdfm/users", body)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}
