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
// 1. TESTING MERGE HISTORY
// ==========================================

func TestCreateMergeHistory(t *testing.T) {
	h := &HistoryHandler{} // Instansiasi Objek
	token := setupMockToken(t)
	payload := model.MergeInput{InputFiles: []string{"a.pdf", "b.pdf"}, OutputFile: "merged.pdf"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/pdfm/log/merge", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.CreateMergeHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetMergeHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/log/merge", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetMergeHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 2. TESTING COMPRESS HISTORY
// ==========================================

func TestCreateCompressHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	payload := model.CompressInput{FileName: "c.pdf", OriginalSize: 100, CompressedSize: 50, Status: "Ok"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/pdfm/log/compress", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.CreateCompressHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetCompressHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/log/compress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetCompressHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 3. TESTING CONVERT HISTORY
// ==========================================

func TestCreateConvertHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	payload := model.ConvertInput{FileName: "doc.pdf", SourceFormat: "pdf", TargetFormat: "docx"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/pdfm/log/convert", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.CreateConvertHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetConvertHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/log/convert", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetConvertHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 4. TESTING SUMMARY HISTORY
// ==========================================

func TestCreateSummaryHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	payload := model.SummaryInput{FileName: "sum.pdf", SummaryText: "Summary", Language: "id"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/pdfm/log/summary", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.CreateSummaryHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestGetSummaryHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/log/summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetSummaryHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

// ==========================================
// 5. TESTING UNIFIED & UTILITY FUNCTIONS
// ==========================================

func TestGetAllHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	req, _ := http.NewRequest("GET", "/pdfm/history/all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.GetAllHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %v", resp.StatusCode)
	}
}

func TestDeleteHistory(t *testing.T) {
	h := &HistoryHandler{}
	token := setupMockToken(t)
	payload := model.DeleteHistoryInput{ID: "invalid", Type: "merge"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("DELETE", "/pdfm/history/delete", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	app := fiber.New()
	app.Add(req.Method, req.URL.Path, h.DeleteHistory)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %v", resp.StatusCode)
	}
}

func TestGetUserFromTokenMissingHeader(t *testing.T) {
	h := &HistoryHandler{}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	// Panggil h.GetUserFromToken
	_, err := h.GetUserFromToken(c)
	if err == nil {
		t.Error("Expected error for missing header, got nil")
	}
}

func TestGetUserFromTokenInvalidFormat(t *testing.T) {
	h := &HistoryHandler{}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Set("Authorization", "InvalidFormat token")
	defer app.ReleaseCtx(c)

	_, err := h.GetUserFromToken(c)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}
