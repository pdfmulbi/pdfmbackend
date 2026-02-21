package controller

import (
	"context"
	"testing"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// setupMockToken menyiapkan data user dan token di MongoDB agar lolos validasi GetUserFromToken
func setupMockToken(t *testing.T) string {
	token := "valid_test_token_history_999"
	email := "tester_history@example.com"

	// 1. Membersihkan data lama di collection users dan tokens agar tidak terjadi conflict
	config.Mongoconn.Collection("users").DeleteMany(context.TODO(), bson.M{"email": email})
	config.Mongoconn.Collection("tokens").DeleteMany(context.TODO(), bson.M{"token": token})

	// 2. Menyiapkan data User baru untuk testing
	user := model.PdfmUsers{
		ID:        primitive.NewObjectID(),
		Name:      "User Tester History",
		Email:     email,
		IsAdmin:   true,
		CreatedAt: time.Now(),
	}
	_, err := atdb.InsertOneDoc(config.Mongoconn, "users", user)
	if err != nil {
		t.Fatalf("Gagal setup mock user: %v", err)
	}

	// 3. Menyiapkan data Token valid yang akan kadaluarsa dalam 1 jam
	tokenData := model.Token{
		Token:     token,
		Email:     email,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "tokens", tokenData)
	if err != nil {
		t.Fatalf("Gagal setup mock token: %v", err)
	}

	return token
}
