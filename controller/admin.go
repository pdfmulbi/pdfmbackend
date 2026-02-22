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

// AdminHandler mengelompokkan fitur admin (Donasi & Activity Logs)
type AdminHandler struct{}

// ==========================================
// 1. MANAGE DONASI (INVOICES)
// ==========================================

// GetAllInvoicesAdmin – Admin melihat semua donasi/invoice
func (h *AdminHandler) GetAllInvoicesAdmin(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{Message: "Forbidden: Admin access required"})
	}

	invoices, err := atdb.GetAllDoc[[]model.Invoice](config.Mongoconn, "invoices", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal mengambil data donasi: " + err.Error()})
	}

	return c.JSON(invoices)
}

// DeleteInvoiceAdmin – Admin menghapus satu donasi/invoice berdasarkan ID
func (h *AdminHandler) DeleteInvoiceAdmin(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{Message: "Forbidden: Admin access required"})
	}

	var req model.DeleteInvoiceInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Data tidak valid"})
	}

	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ResponseMessage{Message: "Format ID tidak valid"})
	}

	_, err = atdb.DeleteOneDoc(config.Mongoconn, "invoices", bson.M{"_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal menghapus donasi: " + err.Error()})
	}

	return c.JSON(model.ResponseMessage{Message: "Donasi berhasil dihapus"})
}

// ==========================================
// 2. ACTIVITY LOGS
// ==========================================

// GetActivityLogs – Admin melihat semua log aktivitas user
// Mendukung filter via query param: ?activity=login|logout|register|merge|compress|convert|summary
func (h *AdminHandler) GetActivityLogs(c *fiber.Ctx) error {
	user, err := h.GetUserFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ResponseMessage{Message: "Unauthorized: " + err.Error()})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ResponseMessage{Message: "Forbidden: Admin access required"})
	}

	filter := bson.M{}
	activityFilter := c.Query("activity")
	if activityFilter != "" && activityFilter != "all" {
		filter["activity"] = activityFilter
	}

	logs, err := atdb.GetAllDoc[[]model.ActivityLog](config.Mongoconn, "activity_logs", filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ResponseMessage{Message: "Gagal mengambil data activity logs: " + err.Error()})
	}

	return c.JSON(logs)
}

// ==========================================
// 3. HELPER
// ==========================================

// GetUserFromToken – Helper validasi token untuk AdminHandler
func (h *AdminHandler) GetUserFromToken(c *fiber.Ctx) (model.PdfmUsers, error) {
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
