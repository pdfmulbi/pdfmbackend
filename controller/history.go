package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/merge [post]
// @Security BearerAuth
func CreateMergeHistory(w http.ResponseWriter, r *http.Request) {
	// 1. Cek siapa yang login
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 2. Decode Input (Pakai struct Input dari Model)
	var req model.MergeInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}

	// 3. Mapping ke Model Database
	data := model.MergeHistory{
		ID:         primitive.NewObjectID(),
		UserID:     user.ID,
		InputFiles: req.InputFiles,
		OutputFile: req.OutputFile,
		CreatedAt:  time.Now(),
	}

	// 4. Simpan ke database
	_, err = atdb.InsertOneDoc(config.Mongoconn, "merge_history", data)
	if err != nil {
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Beri respon sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Log Merge berhasil disimpan",
		"id":      data.ID,
	})
}

// GetMergeHistory godoc
// @Summary Lihat Riwayat Merge
// @Description Menampilkan daftar riwayat merge user
// @Tags History - Merge
// @Accept json
// @Produce json
// @Success 200 {array} model.MergeHistory
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/merge [get]
// @Security BearerAuth
func GetMergeHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	data, err := atdb.GetAllDoc[[]model.MergeHistory](config.Mongoconn, "merge_history", bson.M{"user_id": user.ID})
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ==========================================
// 2. HANDLER UNTUK COMPRESS HISTORY
// ==========================================

// CreateCompressHistory godoc
// @Summary Simpan Log Compress PDF
// @Description Mencatat riwayat kompresi PDF
// @Tags History - Compress
// @Accept json
// @Produce json
// @Param request body model.CompressInput true "Payload Data Compress"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/compress [post]
// @Security BearerAuth
func CreateCompressHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CompressInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
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
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Log Compress berhasil disimpan"})
}

// GetCompressHistory godoc
// @Summary Lihat Riwayat Compress
// @Description Menampilkan daftar riwayat compress user
// @Tags History - Compress
// @Accept json
// @Produce json
// @Success 200 {array} model.CompressHistory
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/compress [get]
// @Security BearerAuth
func GetCompressHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	data, err := atdb.GetAllDoc[[]model.CompressHistory](config.Mongoconn, "compress_history", bson.M{"user_id": user.ID})
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ==========================================
// 3. HANDLER UNTUK CONVERT HISTORY
// ==========================================

// CreateConvertHistory godoc
// @Summary Simpan Log Convert PDF
// @Description Mencatat riwayat konversi PDF (misal: PDF to Word)
// @Tags History - Convert
// @Accept json
// @Produce json
// @Param request body model.ConvertInput true "Payload Data Convert"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/convert [post]
// @Security BearerAuth
func CreateConvertHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.ConvertInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
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
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Log Convert berhasil disimpan"})
}

// GetConvertHistory godoc
// @Summary Lihat Riwayat Convert
// @Description Menampilkan daftar riwayat convert user
// @Tags History - Convert
// @Accept json
// @Produce json
// @Success 200 {array} model.ConvertHistory
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/convert [get]
// @Security BearerAuth
func GetConvertHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	data, err := atdb.GetAllDoc[[]model.ConvertHistory](config.Mongoconn, "convert_history", bson.M{"user_id": user.ID})
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ==========================================
// 4. HANDLER UNTUK SUMMARY HISTORY
// ==========================================

// CreateSummaryHistory godoc
// @Summary Simpan Log Summary PDF
// @Description Mencatat riwayat ringkasan (summary) PDF
// @Tags History - Summary
// @Accept json
// @Produce json
// @Param request body model.SummaryInput true "Payload Data Summary"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/summary [post]
// @Security BearerAuth
func CreateSummaryHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.SummaryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
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
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Log Summary berhasil disimpan"})
}

// GetSummaryHistory godoc
// @Summary Lihat Riwayat Summary
// @Description Menampilkan daftar riwayat summary user
// @Tags History - Summary
// @Accept json
// @Produce json
// @Success 200 {array} model.SummaryHistory
// @Failure 401 {object} map[string]string
// @Router /pdfm/log/summary [get]
// @Security BearerAuth
func GetSummaryHistory(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	data, err := atdb.GetAllDoc[[]model.SummaryHistory](config.Mongoconn, "summary_history", bson.M{"user_id": user.ID})
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ==========================================
// 5. UNIFIED HISTORY - SEMUA AKTIVITAS USER
// ==========================================

// PENTING: Struct HistoryItem LOKAL SUDAH DIHAPUS.
// Kita sekarang pakai `model.HistoryItem` agar sinkron dengan model/history.go

// GetAllHistory godoc
// @Summary Lihat Semua Riwayat (Gabungan)
// @Description Menggabungkan semua jenis riwayat (Merge, Compress, Convert, Summary) menjadi satu list
// @Tags History - Unified
// @Accept json
// @Produce json
// @Success 200 {object} model.UnifiedHistoryResponse
// @Failure 401 {object} map[string]string
// @Router /pdfm/history/all [get]
// @Security BearerAuth
func GetAllHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// PERBAIKAN: Gunakan struct dari Model, bukan struct lokal
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
		for _, c := range compressData {
			allHistory = append(allHistory, model.HistoryItem{
				ID:          c.ID.Hex(),
				Type:        "compress",
				Description: "Compressed PDF file",
				FileName:    c.FileName,
				Details: map[string]interface{}{
					"original_size":   c.OriginalSize,
					"compressed_size": c.CompressedSize,
					"status":          c.Status,
				},
				CreatedAt: c.CreatedAt,
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

	// Sort by CreatedAt descending (terbaru diatas)
	sort.Slice(allHistory, func(i, j int) bool {
		return allHistory[i].CreatedAt.After(allHistory[j].CreatedAt)
	})

	// Return empty array if no history
	if allHistory == nil {
		allHistory = []model.HistoryItem{}
	}

	// PERBAIKAN: Return pakai struct Response, bukan map manual
	json.NewEncoder(w).Encode(model.UnifiedHistoryResponse{
		Status:  200,
		Message: "History retrieved successfully",
		History: allHistory,
	})
}

// ==========================================
// 6. DELETE HISTORY ITEM
// ==========================================

// PENTING: Struct DeleteHistoryRequest LOKAL SUDAH DIHAPUS.

// DeleteHistory godoc
// @Summary Hapus Riwayat
// @Description Menghapus satu item riwayat berdasarkan ID dan Tipe
// @Tags History - Unified
// @Accept json
// @Produce json
// @Param request body model.DeleteHistoryInput true "Payload Hapus History"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /pdfm/history/delete [delete]
// @Security BearerAuth
func DeleteHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := GetUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// PERBAIKAN: Pakai struct dari Model, bukan struct lokal
	var req model.DeleteHistoryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == "" || req.Type == "" {
		http.Error(w, "ID and type are required", http.StatusBadRequest)
		return
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Determine collection based on type
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
		http.Error(w, "Invalid history type", http.StatusBadRequest)
		return
	}

	// Delete from database (only if belongs to user)
	filter := bson.M{"_id": objectID, "user_id": user.ID}
	result, err := config.Mongoconn.Collection(collectionName).DeleteOne(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "History item not found or not authorized", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "History deleted successfully",
	})
}

// ==========================================
// HELPER: MENGAMBIL USER DARI TOKEN
// ==========================================
func GetUserFromToken(r *http.Request) (model.PdfmUsers, error) {
	// ... (Isi fungsi helper ini sama persis dengan yang lama, biarkan saja)
	authHeader := r.Header.Get("Authorization")
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