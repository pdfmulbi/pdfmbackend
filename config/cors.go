package config

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Daftar origins yang diizinkan
var Origins = []string{
	"https://www.bukupedia.co.id",
	"https://naskah.bukupedia.co.id",
	"https://bukupedia.co.id",
	"https://pdfmulbi.github.io",
	"http://127.0.0.1:5500",
	"http://localhost:5500",
}

// Fungsi untuk menormalisasi origin (menghapus trailing slash jika ada)
func normalizeOrigin(origin string) string {
	return strings.TrimRight(origin, "/")
}

// Fungsi untuk memeriksa apakah origin diizinkan
func isAllowedOrigin(origin string) bool {
	normalizedOrigin := normalizeOrigin(origin)
	for _, o := range Origins {
		if o == normalizedOrigin {
			return true
		}
	}
	return false
}

// Fungsi untuk mengatur header CORS sebagai Fiber Middleware
func CorsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		normalizedOrigin := normalizeOrigin(origin)

		// Log origin untuk debugging
		log.Printf("Incoming request from Origin: %s", origin)

		if isAllowedOrigin(normalizedOrigin) {
			// Tambahkan header Vary untuk cache
			c.Set("Vary", "Origin")

			// Tangani preflight request (OPTIONS)
			if c.Method() == fiber.MethodOptions {
				c.Set("Access-Control-Allow-Credentials", "true")
				c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login")
				c.Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PUT, OPTIONS")
				c.Set("Access-Control-Allow-Origin", "*")
				c.Set("Access-Control-Max-Age", "3600")
				return c.SendStatus(fiber.StatusNoContent)
			}

			// Header untuk permintaan utama
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login")

			return c.Next()
		}

		// Jika origin tidak diizinkan, tetap lanjutkan secara normal tapi tanpa header CORS
		return c.Next()
	}
}
