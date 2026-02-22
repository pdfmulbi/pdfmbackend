package controller

import (
	"errors"
	"bytes"    
	"encoding/json" 
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

// CreateMergeHistory godoc
// @Summary Simpan Log Merge PDF
// @Description Mencatat riwayat penggabungan PDF ke database
// @Tags History - Merge
// @Accept json
// @Produce json
// @Param request body model.MergeInput true "Payload Data Merge"
// @Success 200 {object} model.HistoryActionResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/merge [post]
// @Security BearerAuth
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

	// PERBAIKAN: Gunakan HistoryActionResponse
	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Merge berhasil disimpan",
		ID:      data.ID,
	})
}

// GetMergeHistory godoc
// @Summary Lihat Riwayat Merge
// @Description Menampilkan daftar riwayat merge user
// @Tags History - Merge
// @Accept json
// @Produce json
// @Success 200 {array} model.MergeHistory
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/merge [get]
// @Security BearerAuth
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

// CreateCompressHistory godoc
// @Summary Simpan Log Compress PDF
// @Tags History - Compress
// @Accept json
// @Produce json
// @Param request body model.CompressInput true "Payload Data Compress"
// @Success 200 {object} model.HistoryActionResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/compress [post]
// @Security BearerAuth
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

	// PERBAIKAN: Gunakan HistoryActionResponse (Sekarang Compress juga return ID)
	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Compress berhasil disimpan",
		ID:      data.ID,
	})
}

// GetCompressHistory godoc
// @Summary Lihat Riwayat Compress
// @Tags History - Compress
// @Accept json
// @Produce json
// @Success 200 {array} model.CompressHistory
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/compress [get]
// @Security BearerAuth
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

// CreateConvertHistory godoc
// @Summary Simpan Log Convert PDF
// @Tags History - Convert
// @Accept json
// @Produce json
// @Param request body model.ConvertInput true "Payload Data Convert"
// @Success 200 {object} model.HistoryActionResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/convert [post]
// @Security BearerAuth
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

	// PERBAIKAN: Gunakan HistoryActionResponse
	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Convert berhasil disimpan",
		ID:      data.ID,
	})
}

// GetConvertHistory godoc
// @Summary Lihat Riwayat Convert
// @Tags History - Convert
// @Accept json
// @Produce json
// @Success 200 {array} model.ConvertHistory
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/convert [get]
// @Security BearerAuth
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
// 4. HANDLER UNTUK SUMMARY HISTORY
// ==========================================

// CreateSummaryHistory godoc
// @Summary Simpan Log Summary PDF
// @Tags History - Summary
// @Accept json
// @Produce json
// @Param request body model.SummaryInput true "Payload Data Summary"
// @Success 200 {object} model.HistoryActionResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/summary [post]
// @Security BearerAuth
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

	// PERBAIKAN: Gunakan HistoryActionResponse
	return c.Status(fiber.StatusOK).JSON(model.HistoryActionResponse{
		Message: "Log Summary berhasil disimpan",
		ID:      data.ID,
	})
}

// GetSummaryHistory godoc
// @Summary Lihat Riwayat Summary
// @Tags History - Summary
// @Accept json
// @Produce json
// @Success 200 {array} model.SummaryHistory
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/log/summary [get]
// @Security BearerAuth
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

// SummarizePDF godoc
// @Summary Rangkum Dokumen PDF (AI Gemini)
// @Description Merangkum teks dokumen menggunakan Google Gemini AI 1.5 Flash
// @Tags History - Summary
// @Accept json
// @Produce json
// @Param request body model.SummaryRequest true "Payload Teks PDF"
// @Success 200 {object} model.SummaryResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/ai/summary [post]
// @Security BearerAuth
func (h *HistoryHandler) SummarizePDF(c *fiber.Ctx) error {
	// 1. Cek Auth & Ambil Data User Terbaru
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	// 2. CEK PEMBATASAN KUOTA
	// Pastikan field 'SummaryQuota' sudah ada di model.PdfmUsers dan database Anda
	if user.SummaryQuota <= 0 {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{
			Message: "Kuota rangkuman Anda telah habis. Silakan hubungi admin untuk isi ulang.",
		})
	}

	// 3. Ambil Input
	var req struct {
		Content  string `json:"content"`
		FileName string `json:"file_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	// 4. Ambil API Key & Konfigurasi Gemini
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Konfigurasi AI belum siap"})
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey
	prompt := "Rangkum teks dokumen berikut ini secara profesional dan poin-poin penting dalam Bahasa Indonesia: " + req.Content
	
	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{
						"text": prompt,
					},
				},
			},
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	
	// 5. Kirim Request ke Google Gemini
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(model.ResponseMessage{Message: "Gagal menghubungi layanan AI"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal memproses AI"})
	}

	if len(geminiResp.Candidates) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(model.ResponseMessage{Message: "AI tidak memberikan respon"})
	}

	summaryResult := geminiResp.Candidates[0].Content.Parts[0].Text

	// 6. UPDATE DATABASE (POTONG KUOTA & SIMPAN HISTORY)
	// Kurangi kuota user sebanyak 1
	update := bson.M{"$inc": bson.M{"summary_quota": -1}}
	_, err = config.Mongoconn.Collection("users").UpdateOne(c.Context(), bson.M{"_id": user.ID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal memperbarui kuota user"})
	}

	// Simpan data riwayat
	historyData := model.SummaryHistory{
		ID:          primitive.NewObjectID(),
		UserID:      user.ID,
		FileName:    req.FileName,
		SummaryText: summaryResult,
		Language:    "Indonesian",
		CreatedAt:   time.Now(),
	}
	atdb.InsertOneDoc(config.Mongoconn, "summary_history", historyData)

	// 7. Return Hasil & Sisa Kuota
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

// GetAllHistory godoc
// @Summary Lihat Semua Riwayat (Gabungan)
// @Description Menggabungkan semua jenis riwayat (Merge, Compress, Convert, Summary) menjadi satu list
// @Tags History - Unified
// @Accept json
// @Produce json
// @Success 200 {object} model.UnifiedHistoryResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/history/all [get]
// @Security BearerAuth
func (h *HistoryHandler) GetAllHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	var allHistory []model.HistoryItem

	// 1. Ambil Merge History
	mergeData, err := atdb.GetAllDoc[[]model.MergeHistory](config.Mongoconn, "merge_history", bson.M{"user_id": user.ID})
	if err == nil && mergeData != nil {
		for _, m := range mergeData {
			fileCount := len(m.InputFiles)
			allHistory = append(allHistory, model.HistoryItem{
				ID:          m.ID.Hex(),
				Type:        "merge",
				Description: "Merged " + string(rune(fileCount+'0')) + " PDF files",
				FileName:    m.OutputFile,
				Details:     map[string]interface{}{"input_files": m.InputFiles},
				CreatedAt:   m.CreatedAt,
			})
		}
	}

	// 2. Ambil Compress History
	compressData, err := atdb.GetAllDoc[[]model.CompressHistory](config.Mongoconn, "compress_history", bson.M{"user_id": user.ID})
	if err == nil && compressData != nil {
		for _, cm := range compressData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          cm.ID.Hex(),
				Type:        "compress",
				Description: "Compressed PDF file",
				FileName:    cm.FileName,
				Details: map[string]interface{}{
					"original_size":   cm.OriginalSize,
					"compressed_size": cm.CompressedSize,
					"status":          cm.Status,
				},
				CreatedAt: cm.CreatedAt,
			})
		}
	}

	// 3. Ambil Convert History
	convertData, err := atdb.GetAllDoc[[]model.ConvertHistory](config.Mongoconn, "convert_history", bson.M{"user_id": user.ID})
	if err == nil && convertData != nil {
		for _, cv := range convertData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          cv.ID.Hex(),
				Type:        "convert",
				Description: "Converted " + cv.SourceFormat + " to " + cv.TargetFormat,
				FileName:    cv.FileName,
				Details: map[string]interface{}{
					"source_format": cv.SourceFormat,
					"target_format": cv.TargetFormat,
				},
				CreatedAt: cv.CreatedAt,
			})
		}
	}

	// 4. Ambil Summary History
	summaryData, err := atdb.GetAllDoc[[]model.SummaryHistory](config.Mongoconn, "summary_history", bson.M{"user_id": user.ID})
	if err == nil && summaryData != nil {
		for _, s := range summaryData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          s.ID.Hex(),
				Type:        "summary",
				Description: "Generated PDF summary",
				FileName:    s.FileName,
				Details: map[string]interface{}{
					"language": s.Language,
				},
				CreatedAt: s.CreatedAt,
			})
		}
	}

	// Sort by CreatedAt descending
	sort.Slice(allHistory, func(i, j int) bool {
		return allHistory[i].CreatedAt.After(allHistory[j].CreatedAt)
	})

	if allHistory == nil {
		allHistory = []model.HistoryItem{}
	}

	// PERBAIKAN: Gunakan UnifiedHistoryResponse
	return c.Status(fiber.StatusOK).JSON(model.UnifiedHistoryResponse{
		Status:  200,
		Message: "History retrieved successfully",
		History: allHistory,
	})
}

// ==========================================
// 6. DELETE HISTORY ITEM
// ==========================================

// DeleteHistory godoc
// @Summary Hapus Riwayat
// @Description Menghapus satu item riwayat berdasarkan ID dan Tipe
// @Tags History - Unified
// @Accept json
// @Produce json
// @Param request body model.DeleteHistoryInput true "Payload Hapus History"
// @Success 200 {object} model.ResponseMessage
// @Failure 400 {object} model.ResponseMessage
// @Router /pdfm/history/delete [delete]
// @Security BearerAuth
func (h *HistoryHandler) DeleteHistory(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	var req model.DeleteHistoryInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Invalid request body"})
	}

	if req.ID == "" || req.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "ID and type are required"})
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Invalid ID format"})
	}

	var collectionName string
	switch req.Type {
	case "merge":
		collectionName = "merge_history"
	case "compress":
		collectionName = "compress_history"
	case "convert":
		collectionName = "convert_history"
	case "summary":
		collectionName = "summary_history"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Invalid history type"})
	}

	filter := bson.M{"_id": objectID, "user_id": user.ID}
	result, err := config.Mongoconn.Collection(collectionName).DeleteOne(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete")
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(model.ResponseMessage{Message: "History item not found or not authorized"})
	}

	// PERBAIKAN: Gunakan ResponseMessage
	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "History deleted successfully"})
}

// GetUserFromToken (Helper ini harus tetap ada jika belum ada di file lain dalam package yg sama)
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
		return model.PdfmUsers{}, err
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		return model.PdfmUsers{}, errors.New("token sudah kadaluarsa")
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	return user, err
}
