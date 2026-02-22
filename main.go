package gocroot

import (
	"net/http"

	"github.com/gocroot/route"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor" 
)

// WebHook adalah fungsi entri (Entry Point) yang dipanggil oleh Google Cloud Functions.
func WebHook(w http.ResponseWriter, r *http.Request) {
	// 1. Inisialisasi Fiber App
	app := fiber.New()

	// 2. Daftarkan semua route yang sudah Anda buat (termasuk /pdfm/register)
	route.RegisterRoutes(app)

	// 3. Gunakan adaptor untuk mengubah Fiber menjadi standar http.Handler
	// Ini memungkinkan Fiber berjalan di lingkungan serverless GCF
	handler := adaptor.FiberApp(app)
	handler(w, r)
}