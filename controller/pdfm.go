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
func GetUsers(respw http.ResponseWriter, req *http.Request) {
	users, err := atdb.GetAllDoc[[]model.PdfmUsers](config.Mongoconn, "users", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusOK, users)
}

// Get User By ID or Name for  Admin Dahsboard
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
		Amount:        paymentData.Amount,
		Status:        "Paid",
		Details:       "Support Payment",
		PaymentMethod: "QRIS",
		CreatedAt:     time.Now(),
	}

	// Simpan invoice ke koleksi `invoices`
	_, err = atdb.InsertOneDoc(config.Mongoconn, "invoices", invoice)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
		http.Error(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim respon sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Pembayaran telah dilakukan, terima kasih!",
		"invoiceId":   invoice.ID,
		"invoiceDate": invoice.CreatedAt,
		"amountPaid":  invoice.Amount,
	})
}

func GetInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ambil semua invoice dari koleksi `invoices`
	invoices, err := atdb.GetAllDoc[[]model.Invoice](config.Mongoconn, "invoices", bson.M{})
	if err != nil {
		http.Error(w, "Oops! We couldn't fetch the invoices. Please contact support if the issue persists.", http.StatusInternalServerError)
		return
	}

	// Kirim data invoice dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}
