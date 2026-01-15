package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// HANDLER UNTUK FEEDBACK (Masukan User)
// ==========================================

// InsertFeedback godoc
// @Summary Mengirim Feedback (User Login)
// @Description User mengirim kritik dan saran (Wajib Login)
// @Tags Feedback
// @Accept json
// @Produce json
// @Param request body model.FeedbackInput true "Payload Feedback"
// @Success 200 {object} model.FeedbackResponse
// @Failure 400 {object} model.ResponseMessage
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/feedback [post]
// @Security BearerAuth
func InsertFeedback(w http.ResponseWriter, r *http.Request) {
	// 1. Setup Header (Standar CORS)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Cek siapa yang login (Sama seperti di history.go)
	user, err := GetUserFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
		return
	}

	// 3. Siapkan wadah data
	var data model.Feedback

	// 4. Decode data dari Frontend
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ResponseMessage{Message: "Data tidak valid"})
		return
	}

	// Validasi Pesan
	if data.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ResponseMessage{Message: "Pesan tidak boleh kosong"})
		return
	}

	// 5. Lengkapi data server-side
	data.ID = primitive.NewObjectID()
	data.CreatedAt = time.Now()

	// 6. PENTING: Isi data diri otomatis dari Token (Biar aman & valid)
	data.UserID = user.ID.Hex() // Simpan ID user

	// Jika frontend tidak mengirim nama/email, pakai data dari akun login
	if data.Name == "" {
		data.Name = user.Name
	}
	if data.Email == "" {
		data.Email = user.Email
	}

	// 7. Simpan ke database "feedback" pakai atdb helper
	_, err = atdb.InsertOneDoc(config.Mongoconn, "feedback", data)
	if err != nil {
		http.Error(w, "Gagal menyimpan feedback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 8. Beri respon sukses (Format sama persis dengan history.go)
	json.NewEncoder(w).Encode(model.FeedbackResponse{
		Message: "Terima kasih atas masukan Anda!",
		ID:      data.ID,
	})
}

// ==========================================
// GET ALL FEEDBACK (Admin Only)
// ==========================================

// GetAllFeedback godoc
// @Summary Melihat Semua Feedback (Admin Only)
// @Description Hanya admin yang bisa melihat daftar feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Success 200 {array} model.Feedback
// @Failure 401 {object} model.ResponseMessage
// @Failure 403 {object} model.ResponseMessage
// @Router /pdfm/feedback [get]
// @Security BearerAuth
func GetAllFeedback(w http.ResponseWriter, r *http.Request) {
	// 1. Setup Header CORS
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Cek Admin Authentication
	user, err := GetUserFromToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
		return
	}

	// 3. Pastikan user adalah admin
	if !user.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(model.ResponseMessage{Message: "Forbidden: Admin access required"})
		return
	}

	// 4. Ambil semua feedback dari database
	var feedbacks []model.Feedback
	feedbacks, err = atdb.GetAllDoc[[]model.Feedback](config.Mongoconn, "feedback", bson.M{})
	if err != nil {
		http.Error(w, "Gagal mengambil data feedback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Return data feedback
	json.NewEncoder(w).Encode(feedbacks)
}
