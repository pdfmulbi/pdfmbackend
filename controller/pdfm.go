package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "golang.org/x/crypto/bcrypt"
)

type UserHandler struct{}

// RegisterHandler menghandle permintaan registrasi.
// @Summary Pendaftaran Akun Baru
// @Description User mendaftarkan diri dengan Nama, Email, dan Password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterInput true "Payload Register"
// @Success 200 {object} model.ResponseMessage
// @Failure 400 {object} model.ResponseMessage
// @Router /pdfm/register [post]
func (h *UserHandler) RegisterHandler(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.RegisterInput sesuai Swagger
	var req model.RegisterInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Data tidak valid: " + err.Error())
	}

	// Validasi field wajib
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name, Email, dan Password wajib diisi")
	}

	// Mapping ke struct database
	registrationData := model.PdfmUsers{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Email:        req.Email,
		Password:     req.Password, // TODO: Hash password ini untuk keamanan!
		SummaryQuota: 10,
		IsAdmin:      false,
		IsSupport:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Simpan data ke database
	_, err := atdb.InsertOneDoc(config.Mongoconn, "users", registrationData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan data: " + err.Error())
	}

	// Log Activity
	go func() {
		activityLog := model.ActivityLog{
			ID:        primitive.NewObjectID(),
			UserID:    registrationData.ID,
			Name:      registrationData.Name,
			Email:     registrationData.Email,
			Activity:  "register",
			Details:   "User mendaftar akun baru",
			IPAddress: c.IP(),
			CreatedAt: time.Now(),
		}
		atdb.InsertOneDoc(config.Mongoconn, "activity_logs", activityLog)
	}()

	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "Registrasi berhasil"})
}

// GetUser menangani login dan menghasilkan token sederhana
// @Summary Login Pengguna
// @Description Masuk ke sistem untuk mendapatkan Token Akses
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginInput true "Payload login"
// @Success 200 {object} model.LoginResponse
// @Failure 401 {object} model.ResponseMessage
// @Router /pdfm/login [post]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.LoginInput sesuai Swagger
	var req model.LoginInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Data tidak valid: " + err.Error())
	}

	// Cari pengguna di database
	filter := bson.M{"email": req.Email, "password": req.Password}
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Email atau password salah")
	}

	now := time.Now()
	if user.UpdatedAt.Year() != now.Year() || user.UpdatedAt.Month() != now.Month() || user.UpdatedAt.Day() != now.Day() {
		// Jika login terakhir adalah kemarin atau lebih lama, reset kuota ke 10
		user.SummaryQuota = 10
		update := bson.M{
			"$set": bson.M{
				"summary_quota": 10,
				"updatedAt":     now,
			},
		}
		atdb.UpdateOneDoc(config.Mongoconn, "users", bson.M{"_id": user.ID}, update)
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
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan token")
	}

	// Autologing background
	go func(ip string, ua string) {
		loginLog := model.LoginLog{
			ID:        primitive.NewObjectID(),
			UserID:    user.ID,
			Name:      user.Name,
			Email:     user.Email,
			IPAddress: ip,
			UserAgent: ua,
			LoginAt:   time.Now(),
		}
		atdb.InsertOneDoc(config.Mongoconn, "login_logs", loginLog)

		// Activity Log
		activityLog := model.ActivityLog{
			ID:        primitive.NewObjectID(),
			UserID:    user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Activity:  "login",
			Details:   "User login ke sistem",
			IPAddress: ip,
			CreatedAt: time.Now(),
		}
		atdb.InsertOneDoc(config.Mongoconn, "activity_logs", activityLog)
	}(c.IP(), string(c.Request().Header.UserAgent()))

	response := model.LoginResponse{
		Token:    token,
		UserName: user.Name,
		IsAdmin:  user.IsAdmin,
		Message:  "Login berhasil",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// LogoutHandler godoc
// @Summary Keluar Aplikasi (Logout)
// @Description Menghapus token akses dari database
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.ResponseMessage
// @Router /pdfm/logout [post]
// @Security BearerAuth
func (h *UserHandler) LogoutHandler(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Token tidak ditemukan")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusUnauthorized).SendString("Format token tidak valid")
	}
	token := authHeader[len(bearerPrefix):]

	// Ambil info user sebelum hapus token
	tokenData, _ := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	logUser, _ := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})

	_, err := atdb.DeleteOneDoc(config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal logout")
	}

	// Activity Log
	go func() {
		activityLog := model.ActivityLog{
			ID:        primitive.NewObjectID(),
			UserID:    logUser.ID,
			Name:      logUser.Name,
			Email:     logUser.Email,
			Activity:  "logout",
			Details:   "User logout dari sistem",
			IPAddress: c.IP(),
			CreatedAt: time.Now(),
		}
		atdb.InsertOneDoc(config.Mongoconn, "activity_logs", activityLog)
	}()

	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "Logout berhasil"})
}

// GetUsers godoc
// @Summary Ambil Semua Data User (Admin)
// @Description Mengambil list semua pengguna yang terdaftar
// @Tags User Management
// @Accept json
// @Produce json
// @Success 200 {array} model.PdfmUsers
// @Router /pdfm/get/users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	users, err := atdb.GetAllDoc[[]model.PdfmUsers](config.Mongoconn, "users", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(users)
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
func (h *UserHandler) GetOneUserAdmin(c *fiber.Ctx) error {
	id := c.Query("id")
	var filter bson.M

	if id != "" {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid user ID format")
		}
		filter = bson.M{"_id": objectID}
	} else {
		name := c.Query("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Missing user identifier")
		}
		filter = bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	}

	fmt.Printf("Filter: %+v\n", filter)

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(user)
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
func (h *UserHandler) GetOneUser(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing token")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token format")
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired token")
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(user)
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
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.RegisterInput
	var req model.RegisterInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Mapping ke struct DB
	newUser := model.PdfmUsers{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Email:        req.Email,
		Password:     req.Password,
		SummaryQuota: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	count, err := atdb.GetCountDoc(config.Mongoconn, "users", bson.M{"email": newUser.Email})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	if count > 0 {
		return c.Status(fiber.StatusConflict).SendString("Email already exists")
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "users", newUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(newUser)
}

// UpdateUser godoc
// @Summary Update Data User
// @Description Memperbarui data user (nama, password, dll)
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.UpdateUserInput true "Update Payload"
// @Success 200 {object} model.ResponseMessage
// @Router /pdfm/update/users [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.UpdateUserInput
	var req model.UpdateUserInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	if req.ID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("User ID is required")
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format")
	}

	if req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name and Email cannot be empty")
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
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update user: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "User updated successfully"})
}

// DeleteUser godoc
// @Summary Hapus User
// @Description Menghapus user berdasarkan ID
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.DeleteUserInput true "Payload Hapus"
// @Success 200 {object} model.ResponseMessage
// @Router /pdfm/delete/users [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.DeleteUserInput
	var req model.DeleteUserInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid user ID format")
	}

	if _, err := atdb.DeleteOneDoc(config.Mongoconn, "users", bson.M{"_id": objectID}); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(model.ResponseMessage{Message: "User deleted successfully"})
}

// ConfirmPaymentHandler godoc
// @Summary Konfirmasi Pembayaran
// @Description Mengubah status user menjadi Supporter setelah bayar
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body model.PaymentInput true "Payload Payment"
// @Success 200 {object} model.PaymentResponse
// @Router /pdfm/payment [post]
func (h *UserHandler) ConfirmPaymentHandler(c *fiber.Ctx) error {
	// PERBAIKAN: Gunakan model.PaymentInput
	var req model.PaymentInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input: " + err.Error())
	}

	if req.Amount < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("Minimal donasi adalah Rp1")
	}

	filter := bson.M{"name": req.Name}
	var user model.PdfmUsers
	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", filter)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return c.Status(fiber.StatusNotFound).SendString("User not found: " + err.Error())
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
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update user: " + err.Error())
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
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create invoice: " + err.Error())
	}

	log.Printf("[ConfirmPaymentHandler] Invoice created successfully with ID: %s", insertedID.Hex())

	return c.Status(fiber.StatusOK).JSON(model.PaymentResponse{
		Message:     "Pembayaran telah dilakukan, terima kasih!",
		InvoiceId:   invoice.ID,
		InvoiceDate: invoice.CreatedAt,
		AmountPaid:  invoice.Amount,
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
func (h *UserHandler) GetInvoicesHandler(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing token")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token format")
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired token")
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
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
		return c.Status(fiber.StatusInternalServerError).SendString("Oops! We couldn't fetch the invoices.")
	}

	return c.Status(fiber.StatusOK).JSON(invoices)
}

// UploadProfilePhotoHandler handles uploading profile photo (Base64)
// UploadProfilePhotoHandler godoc
// @Summary Upload Foto Profil
// @Description Mengganti foto profil (Format Base64)
// @Tags User Profile
// @Accept json
// @Produce json
// @Param request body model.UploadProfilePhotoInput true "Payload Foto Base64"
// @Success 200 {object} model.ProfilePhotoResponse
// @Router /pdfm/profile/photo [post]
// @Security BearerAuth
func (h *UserHandler) UploadProfilePhotoHandler(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing token")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token format")
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired token")
	}

	// PERBAIKAN: Gunakan model.UploadProfilePhotoInput
	var req model.UploadProfilePhotoInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body: " + err.Error())
	}

	if req.ProfilePhoto == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Profile photo is required")
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
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update profile photo: " + err.Error())
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(model.ProfilePhotoResponse{
		Message: "Profile photo updated successfully",
	})
}

// GetProfilePhotoHandler returns the profile photo for authenticated user
// GetProfilePhotoHandler godoc
// @Summary Lihat Foto Profil
// @Description Mengambil string Base64 foto profil user
// @Tags User Profile
// @Accept json
// @Produce json
// @Success 200 {object} model.ProfilePhotoResponse
// @Router /pdfm/profile/photo [get]
// @Security BearerAuth
func (h *UserHandler) GetProfilePhotoHandler(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing token")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token format")
	}
	token := authHeader[len(bearerPrefix):]

	tokenData, err := atdb.GetOneDoc[model.Token](config.Mongoconn, "tokens", bson.M{"token": token})
	if err != nil || tokenData.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired token")
	}

	user, err := atdb.GetOneDoc[model.PdfmUsers](config.Mongoconn, "users", bson.M{"email": tokenData.Email})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.Status(fiber.StatusOK).JSON(model.ProfilePhotoResponse{
		ProfilePhoto: user.ProfilePhoto,
	})
}
