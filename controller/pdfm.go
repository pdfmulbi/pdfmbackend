package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/kimseokgis/backend-ai/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "golang.org/x/crypto/bcrypt"
)

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
	registrationData.IsSupport = false
	registrationData.LastMergeTime = time.Time{} // Nilai default
	registrationData.MergeCount = 0
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

// GetUser mengambil informasi user dari database berdasarkan email dan password.
func GetUser(respw http.ResponseWriter, req *http.Request) {
	var loginDetails model.PdfmUsers
	if err := json.NewDecoder(req.Body).Decode(&loginDetails); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	var user model.PdfmUsers
	filter := bson.M{"email": loginDetails.Email, "password": loginDetails.Password}
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		helper.WriteJSON(respw, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	helper.WriteJSON(respw, http.StatusOK, user)
}

// Get All Users
func GetUsers(respw http.ResponseWriter, req *http.Request) {
	users, err := atdb.GetAllDoc[[]model.PdfmUsers](config.Mongoconn, "users", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusOK, users)
}

// Get User By ID
func GetOneUser(respw http.ResponseWriter, req *http.Request) {
	// Ambil parameter "name" dari query string
	name := req.URL.Query().Get("name")
	if name == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Missing user name")
		return
	}

	// Buat filter untuk mencari berdasarkan nama
	filter := bson.M{"name": name}

	// Cari user berdasarkan filter
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		helper.WriteJSON(respw, http.StatusNotFound, "User not found")
		return
	}

	// Kirim data user sebagai respons
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
		IsSupport bool   `json:"isSupport"` // Status pembayaran
	}

	// Decode JSON body
	if err := json.NewDecoder(req.Body).Decode(&updateUser); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(updateUser.ID)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Check if isSupport (paid status) is true
	if !updateUser.IsSupport {
		helper.WriteJSON(respw, http.StatusForbidden, "User is not a supporter. Update not allowed.")
		return
	}

	// Create filter and update fields
	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"name":      updateUser.Name,
			"email":     updateUser.Email,
			"password":  updateUser.Password,
			"isSupport": updateUser.IsSupport,
			"updatedAt": time.Now(),
		},
	}

	// Perform update operation
	result, err := atdb.UpdateOneDoc(config.Mongoconn, "users", filter, update)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, "Failed to update user: "+err.Error())
		return
	}

	// Check if any document was updated
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

// MergePDFHandler checks user status and enforces limits for non-premium users
func MergePDFHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Convert userID to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Fetch user data from database
	filter := bson.M{"_id": objID}
	result, err := atdb.GetOneDocPdfm(config.Mongoconn, "users", filter)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user model.PdfmUsers
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, "Failed to decode user data", http.StatusInternalServerError)
		return
	}

	// Parse the number of PDFs from the request
	files := r.MultipartForm.File["pdfs"]
	if len(files) == 0 {
		http.Error(w, "No PDF files uploaded", http.StatusBadRequest)
		return
	}

	// Check user status and enforce PDF count limit
	if !user.IsSupport && len(files) > 3 {
		http.Error(w, "Non-premium users can merge up to 3 PDFs only", http.StatusForbidden)
		return
	}

	// Allow premium users to merge without restriction
	if user.IsSupport {
		// Premium user logic (e.g., log activity if needed)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Premium user: PDF merge allowed"))
		return
	}

	// Non-premium user limit checks
	now := time.Now()
	if user.LastMergeTime.Add(2 * time.Hour).After(now) {
		http.Error(w, "Non-premium users can only merge PDFs every 2 hours", http.StatusTooManyRequests)
		return
	}

	// Update user merge info for non-premium users
	user.LastMergeTime = now
	user.MergeCount++

	// Update user data in the database
	update := bson.M{
		"$set": bson.M{
			"lastMergeTime": user.LastMergeTime,
			"mergeCount":    user.MergeCount,
		},
	}
	_, err = atdb.UpdateOneDoc(config.Mongoconn, "users", filter, update)
	if err != nil {
		http.Error(w, "Failed to update user data", http.StatusInternalServerError)
		return
	}

	// Proceed with merging PDFs (implementation not shown)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PDF merge successful"))
}
