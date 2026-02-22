package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HistoryHandler struct{}

// ==========================================
// 1. HANDLER UNTUK MERGE HISTORY
// ==========================================

func (h *HistoryHandler) CreateMergeHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	var req model.MergeInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	data := model.MergeHistory{
		ID:         primitive.NewObjectID(),
		UserID:     user.ID,
		InputFiles: req.InputFiles,
		OutputFile: req.OutputFile,
		CreatedAt:  time.Now(),
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "merge_history", data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan data")
	}

	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Merge berhasil disimpan",
		ID:      data.ID,
	})
}

func (h *HistoryHandler) GetMergeHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}
	data, err := atdb.GetAllDoc[[]model.MergeHistory](config.Mongoconn, "merge_history", bson.M{"user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch data")
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// ==========================================
// 2. HANDLER UNTUK COMPRESS HISTORY
// ==========================================

func (h *HistoryHandler) CreateCompressHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}

	var req model.CompressInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	data := model.CompressHistory{
		ID:             primitive.NewObjectID(),
		UserID:         user.ID,
		FileName:       req.FileName,
		OriginalSize:   req.OriginalSize,
		CompressedSize: req.CompressedSize,
		Status:         req.Status,
		CreatedAt:      time.Now(),
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "compress_history", data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan data")
	}

	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Compress berhasil disimpan",
		ID:      data.ID,
	})
}

func (h *HistoryHandler) GetCompressHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}
	data, err := atdb.GetAllDoc[[]model.CompressHistory](config.Mongoconn, "compress_history", bson.M{"user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch data")
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// ==========================================
// 3. HANDLER UNTUK CONVERT HISTORY
// ==========================================

func (h *HistoryHandler) CreateConvertHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}

	var req model.ConvertInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	data := model.ConvertHistory{
		ID:           primitive.NewObjectID(),
		UserID:       user.ID,
		FileName:     req.FileName,
		SourceFormat: req.SourceFormat,
		TargetFormat: req.TargetFormat,
		CreatedAt:    time.Now(),
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "convert_history", data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan data")
	}

	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Convert berhasil disimpan",
		ID:      data.ID,
	})
}

func (h *HistoryHandler) GetConvertHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}
	data, err := atdb.GetAllDoc[[]model.ConvertHistory](config.Mongoconn, "convert_history", bson.M{"user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch data")
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// ==========================================
// 4. HANDLER UNTUK SUMMARY & AI GEMINI
// ==========================================

func (h *HistoryHandler) CreateSummaryHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}

	var req model.SummaryInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	data := model.SummaryHistory{
		ID:          primitive.NewObjectID(),
		UserID:      user.ID,
		FileName:    req.FileName,
		SummaryText: req.SummaryText,
		Language:    req.Language,
		CreatedAt:   time.Now(),
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "summary_history", data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan data")
	}

	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Summary berhasil disimpan",
		ID:      data.ID,
	})
}

func (h *HistoryHandler) GetSummaryHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized"})
	}
	data, err := atdb.GetAllDoc[[]model.SummaryHistory](config.Mongoconn, "summary_history", bson.M{"user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch data")
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// SummarizePDF merangkum teks menggunakan Gemini AI 1.5 Flash
func (h *HistoryHandler) SummarizePDF(c *fiber.Ctx) error {
	// 1. Validasi Token & Ambil User Terbaru
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{
			Message: "Sesi tidak valid atau telah berakhir: " + err.Error(),
		})
	}

	// 2. CEK PEMBATASAN KUOTA
	if user.SummaryQuota <= 0 {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{
			Message: "Kuota harian Anda telah habis (Sisa: 0).",
		})
	}

	// 3. Parsing Body Request
	var req struct {
		Content  string `json:"content"`
		FileName string `json:"file_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Format JSON tidak valid"})
	}

	// 4. Konfigurasi Gemini API
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{
			Message: "Error: GEMINI_API_KEY tidak ditemukan di environment server.",
		})
	}

	// Menggunakan model 1.5-flash
	url := "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent?key=" + apiKey
	prompt := "Rangkum teks dokumen berikut secara profesional dalam poin-poin penting menggunakan Bahasa Indonesia: " + req.Content
	
	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{"text": prompt},
				},
			},
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	
	// 5. Eksekusi Request ke Google dengan Timeout 30 Detik
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.ResponseMessage{
			Message: "Gagal menghubungi Google AI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Validasi Respon API Google
	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(model.ResponseMessage{
			Message: "Google AI API Error: " + string(body),
		})
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal memproses data AI"})
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(model.ResponseMessage{Message: "AI tidak memberikan respon konten"})
	}

	summaryResult := geminiResp.Candidates[0].Content.Parts[0].Text

	// 6. UPDATE DATABASE (POTONG KUOTA & SIMPAN LOG)
	update := bson.M{
		"$inc": bson.M{"summary_quota": -1},
		"$set": bson.M{"updatedAt": time.Now()},
	}
	_, err = config.Mongoconn.Collection("users").UpdateOne(c.Context(), bson.M{"_id": user.ID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal memotong kuota"})
	}

	historyData := model.SummaryHistory{
		ID:          primitive.NewObjectID(),
		UserID:      user.ID,
		FileName:    req.FileName,
		SummaryText: summaryResult,
		Language:    "Indonesian",
		CreatedAt:   time.Now(),
	}
	atdb.InsertOneDoc(config.Mongoconn, "summary_history", historyData)

	// 7. Kirim Hasil
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Berhasil merangkum dokumen",
		"summary":         summaryResult,
		"id":              historyData.ID,
		"remaining_quota": user.SummaryQuota - 1,
	})
}

// ==========================================
// 5. UNIFIED HISTORY
// ==========================================

func (h *HistoryHandler) GetAllHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	var allHistory []model.HistoryItem

	// Ambil Merge Data
	mergeData, err := atdb.GetAllDoc[[]model.MergeHistory](config.Mongoconn, "merge_history", bson.M{"user_id": user.ID})
	if err == nil && mergeData != nil {
		for _, m := range mergeData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          m.ID.Hex(),
				Type:        "merge",
				Description: "Menggabungkan PDF",
				FileName:    m.OutputFile,
				CreatedAt:   m.CreatedAt,
			})
		}
	}

	// Ambil Summary Data
	summaryData, err := atdb.GetAllDoc[[]model.SummaryHistory](config.Mongoconn, "summary_history", bson.M{"user_id": user.ID})
	if err == nil && summaryData != nil {
		for _, s := range summaryData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          s.ID.Hex(),
				Type:        "summary",
				Description: "Meringkas PDF",
				FileName:    s.FileName,
				CreatedAt:   s.CreatedAt,
			})
		}
	}

	// Urutkan Terbaru
	sort.Slice(allHistory, func(i, j int) bool {
		return allHistory[i].CreatedAt.After(allHistory[j].CreatedAt)
	})

	return c.Status(fiber.StatusOK).JSON(model.UnifiedHistoryResponse{
		Status:  200,
		Message: "History retrieved successfully",
		History: allHistory,
	})
}

// ==========================================
// 6. DELETE & HELPERS
// ==========================================

func (h *HistoryHandler) DeleteHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	var req model.DeleteHistoryInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Invalid request"})
	}

	objectID, _ := primitive.ObjectIDFromHex(req.ID)
	
	var coll string
	switch req.Type {
	case "merge": coll = "merge_history"
	case "summary": coll = "summary_history"
	default: coll = "merge_history"
	}

	filter := bson.M{"_id": objectID, "user_id": user.ID}
	config.Mongoconn.Collection(coll).DeleteOne(c.Context(), filter)

	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "History deleted"})
}

func (h *HistoryHandler) GetUserFromToken(c *fiber.Ctx) (model.PdfmUsers, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return model.PdfmUsers{}, errors.New("token tidak ditemukan")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return model.PdfmUsers{}, errors.New("format token salah")
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil {
		return model.PdfmUsers{}, errors.New("sesi tidak ditemukan")
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		return model.PdfmUsers{}, errors.New("sesi kadaluarsa")
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	return user, err
}