package config

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Daftar origins yang diizinkan
var Origins = []string{
	"https://www.bukupedia.co.id",
	"https://naskah.bukupedia.co.id",
	"https://bukupedia.co.id",
	"https://pdfmulbi.github.io", // Domain aplikasi Anda
	"http://127.0.0.1:5500",      // Untuk pengetesan lokal
	"http://localhost:5500",
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(origin, "/")
}

func isAllowedOrigin(origin string) bool {
	normalizedOrigin := normalizeOrigin(origin)
	for _, o := range Origins {
		if o == normalizedOrigin {
			return true
		}
	}
	return false
}

// CorsMiddleware utama yang sudah diperbaiki
func CorsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		normalizedOrigin := normalizeOrigin(origin)

		if isAllowedOrigin(normalizedOrigin) {
			c.Set("Access-Control-Allow-Origin", origin) // Set dinamis
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login")
			c.Set("Vary", "Origin")

			if c.Method() == fiber.MethodOptions {
				return c.SendStatus(fiber.StatusNoContent)
			}
		}
		return c.Next()
	}
}