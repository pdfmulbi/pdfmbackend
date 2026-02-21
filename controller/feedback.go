package controller

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedbackHandler adalah Class (Struct) untuk mengelompokkan fitur Feedback
type FeedbackHandler struct{}

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
func (h *FeedbackHandler) InsertFeedback(c *fiber.Ctx) error {
	// 2. Cek siapa yang login
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	// 3. Siapkan wadah data
	var data model.Feedback

	// 4. Decode data
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	if data.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Pesan tidak boleh kosong"})
	}

	// 5. Lengkapi data
	data.ID = primitive.NewObjectID()
	data.CreatedAt = time.Now()
	data.UserID = user.ID.Hex()

	if data.Name == "" {
		data.Name = user.Name
	}
	if data.Email == "" {
		data.Email = user.Email
	}

	// 7. Simpan ke database
	_, err = atdb.InsertOneDoc(config.Mongoconn, "feedback", data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan feedback: " + err.Error())
	}

	// 8. Respon sukses
	return c.JSON(model.FeedbackResponse{
		Message: "Terima kasih atas masukan Anda!",
		ID:      data.ID,
	})
}

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
func (h *FeedbackHandler) GetAllFeedback(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{Message: "Forbidden: Admin access required"})
	}

	var feedbacks []model.Feedback
	feedbacks, err = atdb.GetAllDoc[[]model.Feedback](config.Mongoconn, "feedback", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal mengambil data feedback: " + err.Error())
	}

	return c.JSON(feedbacks)
}

// Helper: GetUserFromToken untuk FeedbackHandler
func (h *FeedbackHandler) GetUserFromToken(c *fiber.Ctx) (model.PdfmUsers, error) {
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
