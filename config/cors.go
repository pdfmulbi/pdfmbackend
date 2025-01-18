package config

import (
	"log"
	"net/http"
	"strings"
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

// Fungsi untuk mengatur header CORS
func SetAccessControlHeaders(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	normalizedOrigin := normalizeOrigin(origin)

	// Log origin untuk debugging
	log.Printf("Incoming request from Origin: %s", origin)

	if isAllowedOrigin(normalizedOrigin) {
		// Tambahkan header Vary untuk cache
		w.Header().Set("Vary", "Origin")

		// Tangani preflight request (OPTIONS)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusNoContent)
			return true
		}

		// Header untuk permintaan utama
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login")
		return false
	}

	return false
}
