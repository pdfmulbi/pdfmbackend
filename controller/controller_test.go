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
	// Data registrasi yang sesuai dengan validasi
	data := model.PdfmUsers{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Konversi data ke JSON
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Gagal meng-encode JSON: %v", err)
	}

	// Jalankan permintaan dengan fungsi helper
	rr := executeRequest(t, controller.RegisterHandler, http.MethodPost, "/pdfm/register", body)

	// Periksa status respons
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	// Periksa isi respons
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Gagal decode respons JSON: %v", err)
	}

	// Validasi isi respons
	if response["message"] != "Registrasi berhasil" {
		t.Errorf("Expected message 'Registrasi berhasil', got '%s'", response["message"])
	}
}

// Test GetUser (Login)
func TestGetUser(t *testing.T) {
	data := model.PdfmUsers{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.GetUser, http.MethodPost, "/pdfm/login", body)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}

// Test MergePDFHandler
func TestMergePDFHandler(t *testing.T) {
	reqURL := "/pdfm/merge?user_id=" + primitive.NewObjectID().Hex()
	rr := executeRequest(t, controller.MergePDFHandler, http.MethodPost, reqURL, nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
		t.Errorf("expected status OK or Forbidden, got %v", rr.Code)
	}
}

// Test GetUsers
func TestGetUsers(t *testing.T) {
	rr := executeRequest(t, controller.GetUsers, http.MethodGet, "/pdfm/users", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
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
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}

// Test GetOneUser
func TestGetOneUser(t *testing.T) {
	reqURL := "/pdfm/users/details?id=" + primitive.NewObjectID().Hex()
	rr := executeRequest(t, controller.GetOneUser, http.MethodGet, reqURL, nil)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("expected status OK or NotFound, got %v", rr.Code)
	}
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
	data := model.PdfmUsers{
		ID:   primitive.NewObjectID(),
		Name: "Updated Name",
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.UpdateUser, http.MethodPut, "/pdfm/users", body)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	data := model.PdfmUsers{
		ID: primitive.NewObjectID(),
	}
	body, _ := json.Marshal(data)

	rr := executeRequest(t, controller.DeleteUser, http.MethodDelete, "/pdfm/users", body)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}
