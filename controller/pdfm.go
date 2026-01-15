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
	"github.com/kimseokgis/backend-ai/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "golang.org/x/crypto/bcrypt"
)

// Register
// RegisterHandler menghandle permintaan registrasi.
// @Summary Pendaftaran Akun Baru
// @Description User mendaftarkan diri dengan Nama, Email, dan Password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterInput "Payload Register"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /pdfm/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var registrationData model.PdfmUsers

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&registrationData)
	if err != nil {
		http.Error(w, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi field wajib
	if registrationData.Name == "" || registrationData.Email == "" || registrationData.Password == "" {
		http.Error(w, "Name, Email, dan Password wajib diisi", http.StatusBadRequest)
		return
	}

	// // Hash password sebelum menyimpan ke database
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	// if err != nil {
	// 	http.Error(w, "Gagal memproses password: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// registrationData.Password = string(hashedPassword)

	// Set nilai default untuk field lainnya
	registrationData.ID = primitive.NewObjectID()
	registrationData.IsAdmin = false
	registrationData.IsSupport = false
	registrationData.CreatedAt = time.Now()
	registrationData.UpdatedAt = time.Now()

	// Simpan data ke database
	_, err = atdb.InsertOneDoc(config.Mongoconn, "users", registrationData)
	if err != nil {
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respon sukses
	response := map[string]string{"message": "Registrasi berhasil"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login
// GetUser menangani login dan menghasilkan token sederhana
// GetUser godoc
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

	// Decode data login dari request body
	var loginDetails model.PdfmUsers
	if err := json.NewDecoder(r.Body).Decode(&loginDetails); err != nil {
		http.Error(w, "Data tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Cari pengguna di database berdasarkan email dan password
	filter := bson.M{"email": loginDetails.Email, "password": loginDetails.Password}
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	}

	// Buat token unik
	token := uuid.New().String()

	// Tentukan waktu kedaluwarsa (misalnya 24 jam)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Simpan token ke database
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

	// Membau autologing untuk mengcek keaktifkan pengguna
	go func() { // Pakai 'go func' biar jalan di background & gak bikin lemot login
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

// Logout
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

	// Hapus token dari database
	_, err := atdb.DeleteOneDoc(config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Logout berhasil"}`))
}

// CRUD
// Get All Users
// GetUsers godoc
// @Summary Ambil Semua Data User (Admin)
// @Description Mengambil list semua pengguna yang terdaftar
// @Tags User Management
// @Accept json
// @Produce json
// @Success 200 {array} model.PdfmUsers
// @Router /pdfm/get/users [get]
func GetUsers(respw http.ResponseWriter, req *http.Request) {
	users, err := atdb.GetAllDoc[[]model.PdfmUsers](config.Mongoconn, "users", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusOK, users)
}

// Get User By ID or Name for  Admin Dahsboard
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
func GetOneUserAdmin(respw http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	var filter bson.M
	if id != "" {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			helper.WriteJSON(respw, http.StatusBadRequest, "Invalid user ID format (GetOneUser)")
			return
		}
		filter = bson.M{"_id": objectID}
	} else {
		name := req.URL.Query().Get("name")
		if name == "" {
			helper.WriteJSON(respw, http.StatusBadRequest, "Missing user identifier")
			return
		}
		// Use case-insensitive regex for name matching
		filter = bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	}

	fmt.Printf("Filter: %+v\n", filter) // Log filter for debugging

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		fmt.Printf("Error: %v\n", err) // Log error for debugging
		helper.WriteJSON(respw, http.StatusNotFound, "User not found")
		return
	}

	helper.WriteJSON(respw, http.StatusOK, user)
}

// Get User Token
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
func GetOneUser(respw http.ResponseWriter, req *http.Request) {
	// Ambil token dari header Authorization
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		helper.WriteJSON(respw, http.StatusUnauthorized, "Missing token")
		return
	}

	// Validasi format token (Bearer <token>)
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		helper.WriteJSON(respw, http.StatusUnauthorized, "Invalid token format")
		return
	}
	token := authHeader[len(bearerPrefix):]

	// Validasi token di database
	var tokenData model.Token
	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		helper.WriteJSON(respw, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Ambil data pengguna berdasarkan email yang terkait dengan token
	var user model.PdfmUsers
	user, err = atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		helper.WriteJSON(respw, http.StatusNotFound, "User not found")
		return
	}

	// Kembalikan data user dalam respons
	helper.WriteJSON(respw, http.StatusOK, user)
}

// Create User
// CreateUser godoc
// @Summary Tambah User Manual (Admin)
// @Description Membuat user baru secara langsung (bypass register)
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.RegisterInput true "Create Payload"
// @Success 200 {object} model.PdfmUsers
// @Router /pdfm/create/users [post]
func CreateUser(respw http.ResponseWriter, req *http.Request) {
	var newUser model.PdfmUsers
	if err := json.NewDecoder(req.Body).Decode(&newUser); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	newUser.ID = primitive.NewObjectID()
	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = time.Now()

	// Check for duplicate email
	count, err := atdb.GetCountDoc(config.Mongoconn, "users", bson.M{"email": newUser.Email})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	if count > 0 {
		helper.WriteJSON(respw, http.StatusConflict, "Email already exists")
		return
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "users", newUser); err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}

	helper.WriteJSON(respw, http.StatusOK, newUser)
}

// Update User
// UpdateUser godoc
// @Summary Update Data User
// @Description Memperbarui data user (nama, password, dll)
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.UpdateUserInput true "Update Payload"
// @Success 200 {string} string "User updated successfully"
// @Router /pdfm/update/users [put]
func UpdateUser(respw http.ResponseWriter, req *http.Request) {
	var updateUser struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		IsSupport bool   `json:"isSupport"`
	}

	// Decode JSON body
	if err := json.NewDecoder(req.Body).Decode(&updateUser); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validasi ID
	if updateUser.ID == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "User ID is required")
		return
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(updateUser.ID)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Validasi data lainnya
	if updateUser.Name == "" || updateUser.Email == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Name and Email cannot be empty")
		return
	}

	// Buat filter dan pipeline update
	filter := bson.M{"_id": objectID}
	pipeline := bson.M{
		"$set": bson.M{
			"name":      updateUser.Name,
			"email":     updateUser.Email,
			"password":  updateUser.Password,
			"isSupport": updateUser.IsSupport,
			"updatedAt": time.Now(),
		},
	}

	// Perform update operation
	result, err := atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, []bson.M{pipeline})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, "Failed to update user: "+err.Error())
		return
	}

	// Periksa apakah ada dokumen yang diperbarui
	if result.MatchedCount == 0 {
		helper.WriteJSON(respw, http.StatusNotFound, "User not found")
		return
	}

	helper.WriteJSON(respw, http.StatusOK, "User updated successfully")
}

// Delete User
// DeleteUser godoc
// @Summary Hapus User
// @Description Menghapus user berdasarkan ID
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.DeleteUserInput true "Payload Hapus"
// @Success 200 {string} string "User deleted successfully"
// @Router /pdfm/delete/users [delete]
func DeleteUser(respw http.ResponseWriter, req *http.Request) {
	var user struct {
		ID string `json:"id"`
	}

	// Decode JSON body to temporary struct
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	// Convert ID from string to primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Delete document by ObjectID
	if _, err := atdb.DeleteOneDoc(config.Mongoconn, "users", bson.M{"_id": objectID}); err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}

	helper.WriteJSON(respw, http.StatusOK, "User deleted successfully")
}

// ConfirmPaymentHandler handles the payment confirmation and updates user status.
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

	var paymentData struct {
		Name   string `json:"name"`   // Nama pengguna
		Amount int    `json:"amount"` // Nominal pembayaran
	}

	if err := json.NewDecoder(r.Body).Decode(&paymentData); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi jumlah donasi minimum
	if paymentData.Amount < 1 {
		http.Error(w, "Minimal donasi adalah Rp1", http.StatusBadRequest)
		return
	}

	// Filter untuk mencari pengguna berdasarkan nama
	filter := bson.M{"name": paymentData.Name}

	// Pastikan pengguna ada
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// DEBUG: Log user data yang ditemukan
	log.Printf("[ConfirmPaymentHandler] User found: ID=%s, Name=%s, Email='%s'", user.ID.Hex(), user.Name, user.Email)

	// Validasi: Jika user tidak punya email, coba cari berdasarkan token
	if user.Email == "" {
		log.Printf("[ConfirmPaymentHandler] WARNING: User has no email in database!")
	}

	// // Periksa apakah pengguna sudah menjadi supporter
	// if user.IsSupport {
	// 	http.Error(w, "User is already a supporter", http.StatusBadRequest)
	// 	return
	// }

	// // Periksa duplikasi invoice
	// _, err = atdb.GetOneDoc[model.Invoice](config.Mongoconn, "invoices", bson.M{
	// 	"name":   paymentData.Name,
	// 	"amount": paymentData.Amount,
	// 	"status": "Paid",
	// })
	// if err == nil {
	// 	http.Error(w, "Invoice already exists for this payment", http.StatusBadRequest)
	// 	return
	// }

	// Gunakan pipeline untuk memperbarui pengguna
	pipeline := []bson.M{
		{"$set": bson.M{
			"isSupport": true,
			"updatedAt": time.Now(),
		}},
	}

	// Perbarui pengguna
	_, err = atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, pipeline)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat invoice baru
	invoice := model.Invoice{
		ID:            primitive.NewObjectID(),
		Name:          user.Name,
		Email:         user.Email,
		Amount:        paymentData.Amount,
		Status:        "Paid",
		Details:       "Support Payment",
		PaymentMethod: "QRIS",
		CreatedAt:     time.Now(),
	}

	// DEBUG: Log invoice yang akan disimpan
	log.Printf("[ConfirmPaymentHandler] Creating invoice: Name='%s', Email='%s', Amount=%d", invoice.Name, invoice.Email, invoice.Amount)

	// Simpan invoice ke koleksi `invoices`
	insertedID, err := atdb.InsertOneDoc(config.Mongoconn, "invoices", invoice)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
		http.Error(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// DEBUG: Log invoice berhasil disimpan
	log.Printf("[ConfirmPaymentHandler] Invoice created successfully with ID: %s", insertedID.Hex())

	// Kirim respon sukses
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

	// Validasi token dari header Authorization
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

	// Validasi token di database
	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Log email dari token untuk debugging
	log.Printf("[GetInvoicesHandler] Token valid, email: %s", tokenData.Email)

	// Ambil data user berdasarkan email dari token
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		log.Printf("[GetInvoicesHandler] User not found for email: %s, error: %v", tokenData.Email, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Log user ditemukan untuk debugging
	log.Printf("[GetInvoicesHandler] User found: name=%s, email=%s", user.Name, user.Email)

	// Filter invoice berdasarkan NAME user (lebih reliable karena semua invoice pasti punya name)
	// Untuk keamanan tambahan, jika user punya email, cek juga invoice dengan email yang cocok
	var filter bson.M
	if user.Email != "" {
		// User punya email: cari invoice dengan email ATAU name yang cocok
		filter = bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"name": user.Name},
			},
		}
	} else {
		// User tidak punya email: cari hanya berdasarkan name
		filter = bson.M{"name": user.Name}
	}
	log.Printf("[GetInvoicesHandler] Fetching invoices with filter: %+v", filter)

	invoices, err := atdb.GetAllDoc[[]model.Invoice](config.Mongoconn, "invoices", filter)
	if err != nil {
		log.Printf("[GetInvoicesHandler] Error fetching invoices: %v", err)
		http.Error(w, "Oops! We couldn't fetch the invoices. Please contact support if the issue persists.", http.StatusInternalServerError)
		return
	}

	// Log jumlah invoice yang ditemukan
	log.Printf("[GetInvoicesHandler] Found %d invoices for user: %s", len(invoices), user.Name)

	// Kirim data invoice dalam format JSON
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

	// Validasi token dari header Authorization
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

	// Validasi token di database
	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var requestBody struct {
		ProfilePhoto string `json:"profilePhoto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi foto tidak kosong
	if requestBody.ProfilePhoto == "" {
		http.Error(w, "Profile photo is required", http.StatusBadRequest)
		return
	}

	// Update user dengan foto profil baru
	filter := bson.M{"email": tokenData.Email}
	pipeline := []bson.M{
		{"$set": bson.M{
			"profilePhoto": requestBody.ProfilePhoto,
			"updatedAt":    time.Now(),
		}},
	}

	result, err := atdb.UpdateWithPipeline(config.Mongoconn, "users", filter, pipeline)
	if err != nil {
		log.Printf("[UploadProfilePhotoHandler] Error updating profile photo: %v", err)
		http.Error(w, "Failed to update profile photo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.Printf("[UploadProfilePhotoHandler] Profile photo updated for user: %s", tokenData.Email)

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

	// Validasi token dari header Authorization
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

	// Validasi token di database
	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Ambil data user
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
