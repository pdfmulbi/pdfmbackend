package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/model"
)

func executeRequest(t *testing.T, handler fiber.Handler, method, url string, body []byte, token string) *http.Response {
	app := fiber.New()
	app.Add(method, url, handler)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// Test Skenario untuk InsertFeedback
func TestInsertFeedback(t *testing.T) {
	// PENTING: Instansiasi objek h agar bisa memanggil method InsertFeedback
	h := &FeedbackHandler{}

	// Skenario 1: Gagal karena tidak ada Token (Unauthorized)
	t.Run("Unauthorized Access", func(t *testing.T) {
		feedbackData := model.Feedback{Message: "Sangat membantu!"}
		body, _ := json.Marshal(feedbackData)

		// Kita panggil h.InsertFeedback (OOP style)
		rr := executeRequest(t, h.InsertFeedback, http.MethodPost, "/pdfm/feedback", body, "")

		if rr.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %v", rr.StatusCode)
		}
	})

	// Skenario 2: Gagal karena Method bukan POST
	t.Run("Method Not Allowed", func(t *testing.T) {
		rr := executeRequest(t, h.InsertFeedback, http.MethodGet, "/pdfm/feedback", nil, "sample_token")

		if rr.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %v", rr.StatusCode)
		}
	})

	// Skenario 3: Gagal karena Pesan Kosong (Bad Request)
	t.Run("Empty Message Validation", func(t *testing.T) {
		feedbackData := model.Feedback{Message: ""}
		body, _ := json.Marshal(feedbackData)

		rr := executeRequest(t, h.InsertFeedback, http.MethodPost, "/pdfm/feedback", body, "sample_token")

		// Logika pengecekan status tetap sama
		if rr.StatusCode != http.StatusBadRequest && rr.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 400 or 401, got %v", rr.StatusCode)
		}
	})
}

// Test Skenario untuk GetAllFeedback
func TestGetAllFeedback(t *testing.T) {
	h := &FeedbackHandler{}

	// Skenario 1: Mencoba akses tanpa role Admin (Forbidden)
	t.Run("Access by Non-Admin", func(t *testing.T) {
		rr := executeRequest(t, h.GetAllFeedback, http.MethodGet, "/pdfm/feedback", nil, "regular_user_token")

		if rr.StatusCode != http.StatusForbidden && rr.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 403 or 401, got %v", rr.StatusCode)
		}
	})

	// Skenario 2: Akses dengan Method yang salah
	t.Run("Invalid Method", func(t *testing.T) {
		rr := executeRequest(t, h.GetAllFeedback, http.MethodPost, "/pdfm/feedback", nil, "admin_token")

		if rr.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %v", rr.StatusCode)
		}
	})
}
