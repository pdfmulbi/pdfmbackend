package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "golang.org/x/crypto/bcrypt"
)

// RegisterHandler menghandle permintaan registrasi.
// @Summary Pendaftaran Akun Baru
// @Description User mendaftarkan diri dengan Nama, Email, dan Password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterInput true "Payload Register"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /pdfm/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// PERBAIKAN: Gunakan model.RegisterInput sesuai Swagger
	var req model.RegisterInput
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi field wajib
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, Email, dan Password wajib diisi", http.StatusBadRequest)
		return
	}

	// Mapping ke struct database
	registrationData := model.PdfmUsers{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password, // TODO: Hash password ini untuk keamanan!
		IsAdmin:   false,
		IsSupport: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Simpan data ke database
	_, err = atdb.InsertOneDoc(config.Mongoconn, "users", registrationData)
	if err != nil {
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Registrasi berhasil"})
}

// GetUser menangani login dan menghasilkan token sederhana
// @Summary Login Pengguna
// @Description Masuk ke sistem untuk mendapatkan Token Akses
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginInput true "Payload login"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /pdfm/login [post]
func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// PERBAIKAN: Gunakan model.LoginInput sesuai Swagger
	var req model.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Cari pengguna di database
	filter := bson.M{"email": req.Email, "password": req.Password}
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	}

	// Buat token unik (UUID)
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	tokenData := model.Token{
		Token:     token,
		Email:     user.Email,
		ExpiresAt: expiresAt,
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "tokens", tokenData)
	if err != nil {
		http.Error(w, "Gagal menyimpan token", http.StatusInternalServerError)
		return
	}

	// Autologing background
	go func() {
		loginLog := model.LoginLog{
			ID:        primitive.NewObjectID(),
			UserID:    user.ID,
			Name:      user.Name,
			Email:     user.Email,
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
			LoginAt:   time.Now(),
		}
		atdb.InsertOneDoc(config.Mongoconn, "login_logs", loginLog)
	}()

	response := map[string]interface{}{
		"token":    token,
		"userName": user.Name,
		"isAdmin":  user.IsAdmin,
		"message":  "Login berhasil",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler godoc
// @Summary Keluar Aplikasi (Logout)
// @Description Menghapus token akses dari database
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /pdfm/logout [post]
// @Security BearerAuth
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Token tidak ditemukan", http.StatusBadRequest)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Format token tidak valid", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(bearerPrefix):]

	_, err := atdb.DeleteOneDoc(config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Logout berhasil"}`))
}

// GetUsers godoc
// @Summary Ambil Semua Data User (Admin)
// @Description Mengambil list semua pengguna yang terdaftar
// @Tags User Management
// @Accept json
// @Produce json
// @Success 200 {array} model.PdfmUsers
// @Router /pdfm/get/users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := atdb.GetAllDoc[[]model.PdfmUsers](config.Mongoconn, "users", bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetOneUserAdmin godoc
// @Summary Cari Satu User (Admin)
// @Description Mencari user berdasarkan Query Param ID atau Name
// @Tags User Management
// @Accept json
// @Produce json
// @Param id query string false "User ID"
// @Param name query string false "User Name"
// @Success 200 {object} model.PdfmUsers
// @Router /pdfm/getoneadmin/users [get]
func GetOneUserAdmin(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var filter bson.M
	
	if id != "" {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		filter = bson.M{"_id": objectID}
	} else {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "Missing user identifier", http.StatusBadRequest)
			return
		}
		filter = bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	}

	fmt.Printf("Filter: %+v\n", filter)

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetOneUser godoc
// @Summary Cek Profil Saya
// @Description Mengambil data user yang sedang login berdasarkan Token
// @Tags User Profile
// @Accept json
// @Produce json
// @Success 200 {object} model.PdfmUsers
// @Failure 401 {object} map[string]string
// @Router /pdfm/getone/users [get]
// @Security BearerAuth
func GetOneUser(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// CreateUser godoc
// @Summary Tambah User Manual (Admin)
// @Description Membuat user baru secara langsung (bypass register)
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.RegisterInput true "Create Payload"
// @Success 200 {object} model.PdfmUsers
// @Router /pdfm/create/users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// PERBAIKAN: Gunakan model.RegisterInput
	var req model.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mapping ke struct DB
	newUser := model.PdfmUsers{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	count, err := atdb.GetCountDoc(config.Mongoconn, "users", bson.M{"email": newUser.Email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "users", newUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}

// UpdateUser godoc
// @Summary Update Data User
// @Description Memperbarui data user (nama, password, dll)
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.UpdateUserInput true "Update Payload"
// @Success 200 {string} string "User updated successfully"
// @Router /pdfm/update/users [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// PERBAIKAN: Gunakan model.UpdateUserInput
	var req model.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and Email cannot be empty", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}
	pipeline := bson.M{
		"$set": bson.M{
			"name":      req.Name,
			"email":     req.Email,
			"password":  req.Password,
			"isSupport": req.IsSupport,
			"updatedAt": time.Now(),
		},
	}

	result, err := atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, []bson.M{pipeline})
	if err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("User updated successfully")
}

// DeleteUser godoc
// @Summary Hapus User
// @Description Menghapus user berdasarkan ID
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.DeleteUserInput true "Payload Hapus"
// @Success 200 {string} string "User deleted successfully"
// @Router /pdfm/delete/users [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// PERBAIKAN: Gunakan model.DeleteUserInput
	var req model.DeleteUserInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	if _, err := atdb.DeleteOneDoc(config.Mongoconn, "users", bson.M{"_id": objectID}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("User deleted successfully")
}

// ConfirmPaymentHandler godoc
// @Summary Konfirmasi Pembayaran
// @Description Mengubah status user menjadi Supporter setelah bayar
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body model.PaymentInput true "Payload Payment"
// @Success 200 {object} map[string]interface{}
// @Router /pdfm/payment [post]
func ConfirmPaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// PERBAIKAN: Gunakan model.PaymentInput
	var req model.PaymentInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount < 1 {
		http.Error(w, "Minimal donasi adalah Rp1", http.StatusBadRequest)
		return
	}

	filter := bson.M{"name": req.Name}
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[ConfirmPaymentHandler] User found: ID=%s, Name=%s, Email='%s'", user.ID.Hex(), user.Name, user.Email)

	pipeline := []bson.M{
		{"$set": bson.M{
			"isSupport": true,
			"updatedAt": time.Now(),
		}},
	}

	_, err = atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, pipeline)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	invoice := model.Invoice{
		ID:            primitive.NewObjectID(),
		Name:          user.Name,
		Email:         user.Email,
		Amount:        req.Amount,
		Status:        "Paid",
		Details:       "Support Payment",
		PaymentMethod: "QRIS",
		CreatedAt:     time.Now(),
	}

	insertedID, err := atdb.InsertOneDoc(config.Mongoconn, "invoices", invoice)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
		http.Error(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[ConfirmPaymentHandler] Invoice created successfully with ID: %s", insertedID.Hex())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Pembayaran telah dilakukan, terima kasih!",
		"invoiceId":   invoice.ID,
		"invoiceDate": invoice.CreatedAt,
		"amountPaid":  invoice.Amount,
	})
}

// GetInvoicesHandler godoc
// @Summary Lihat Invoice Saya
// @Description Melihat riwayat pembayaran pengguna yang login
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {array} model.Invoice
// @Router /pdfm/invoices [get]
// @Security BearerAuth
func GetInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var filter bson.M
	if user.Email != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"name": user.Name},
			},
		}
	} else {
		filter = bson.M{"name": user.Name}
	}

	invoices, err := atdb.GetAllDoc[[]model.Invoice](config.Mongoconn, "invoices", filter)
	if err != nil {
		http.Error(w, "Oops! We couldn't fetch the invoices.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

// UploadProfilePhotoHandler handles uploading profile photo (Base64)
// UploadProfilePhotoHandler godoc
// @Summary Upload Foto Profil
// @Description Mengganti foto profil (Format Base64)
// @Tags User Profile
// @Accept json
// @Produce json
// @Param request body model.UploadProfilePhotoInput true "Payload Foto Base64"
// @Success 200 {object} map[string]string
// @Router /pdfm/profile/photo [post]
// @Security BearerAuth
func UploadProfilePhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// PERBAIKAN: Gunakan model.UploadProfilePhotoInput
	var req model.UploadProfilePhotoInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.ProfilePhoto == "" {
		http.Error(w, "Profile photo is required", http.StatusBadRequest)
		return
	}

	filter := bson.M{"email": tokenData.Email}
	pipeline := []bson.M{
		{"$set": bson.M{
			"profilePhoto": req.ProfilePhoto,
			"updatedAt":    time.Now(),
		}},
	}

	result, err := atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, pipeline)
	if err != nil {
		http.Error(w, "Failed to update profile photo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile photo updated successfully",
	})
}

// GetProfilePhotoHandler returns the profile photo for authenticated user
// GetProfilePhotoHandler godoc
// @Summary Lihat Foto Profil
// @Description Mengambil string Base64 foto profil user
// @Tags User Profile
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /pdfm/profile/photo [get]
// @Security BearerAuth
func GetProfilePhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profilePhoto": user.ProfilePhoto,
	})
}