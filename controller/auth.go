package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/auth"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func RegisterGmailAuth(c *fiber.Ctx) error {
	logintoken, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeaderFiber(c)
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var request struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"message": "Invalid request"})
	}

	// Ambil kredensial dari database
	creds, err := atdb.GetOneDoc[auth.GoogleCredential](config.Mongoconn, "credentials", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(map[string]string{"message": "Database Connection Problem: Unable to fetch credentials"})
	}

	// Verifikasi ID token menggunakan client_id
	payload, err := auth.VerifyIDToken(request.Token, creds.ClientID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]string{"message": "Invalid token: Token verification failed"})
	}

	userInfo := model.Userdomyikado{
		Name:                 payload.Claims["name"].(string),
		PhoneNumber:          logintoken.Id,
		Email:                payload.Claims["email"].(string),
		GoogleProfilePicture: payload.Claims["picture"].(string),
	}

	// Simpan atau perbarui informasi pengguna di database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("user")
	filter := bson.M{"phonenumber": logintoken.Id}

	var existingUser model.Userdomyikado
	err = collection.FindOne(ctx, filter).Decode(&existingUser)
	if err != nil || existingUser.PhoneNumber == "" {
		// User does not exist or exists but has no phone number, insert into db
		id, err := atdb.InsertOneDoc(config.Mongoconn, "user", userInfo)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(map[string]string{"message": "Database Connection Problem: Unable to fetch credentials"})
		}
		response := map[string]interface{}{
			"message": "User Berhasil Terdaftar",
			"user":    userInfo,
			"id":      id.Hex(),
		}
		return c.Status(fiber.StatusOK).JSON(response)
	} else if existingUser.PhoneNumber != "" {
		existingUser.Email = userInfo.Email
		existingUser.GoogleProfilePicture = userInfo.GoogleProfilePicture
		_, err := atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"_id": existingUser.ID}, existingUser)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(map[string]string{"message": "Database Connection Problem: Unable to update user"})
		}
		response := map[string]interface{}{
			"message": "Authenticated successfully",
			"user":    existingUser,
			"id":      existingUser.ID,
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	update := bson.M{
		"$set": userInfo,
	}
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"message": "Failed to save user info: Database update failed"})
	}

	response := map[string]interface{}{
		"user": userInfo,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func Auth(c *fiber.Ctx) error {
	var request struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"message": "Invalid request"})
	}

	// Ambil kredensial dari database
	creds, err := atdb.GetOneDoc[auth.GoogleCredential](config.Mongoconn, "credentials", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(map[string]string{"message": "Database Connection Problem: Unable to fetch credentials"})
	}

	// Verifikasi ID token menggunakan client_id
	payload, err := auth.VerifyIDToken(request.Token, creds.ClientID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]string{"message": "Invalid token: Token verification failed"})
	}

	userInfo := model.Userdomyikado{
		Name:                 payload.Claims["name"].(string),
		Email:                payload.Claims["email"].(string),
		GoogleProfilePicture: payload.Claims["picture"].(string),
	}

	// Simpan atau perbarui informasi pengguna di database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("user")
	filter := bson.M{"email": userInfo.Email}

	var existingUser model.Userdomyikado
	err = collection.FindOne(ctx, filter).Decode(&existingUser)
	if err != nil || existingUser.PhoneNumber == "" {
		// User does not exist or exists but has no phone number, request QR scan
		response := map[string]interface{}{
			"message": "Please scan the QR code to provide your phone number",
			"user":    userInfo,
			"token":   "",
		}
		return c.Status(fiber.StatusUnauthorized).JSON(response)
	} else if existingUser.PhoneNumber != "" {
		token, err := watoken.EncodeforHours(existingUser.PhoneNumber, existingUser.Name, config.PrivateKey, 18) // Generating a token for 18 hours
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"message": "Token generation failed"})
		}
		response := map[string]interface{}{
			"message": "Authenticated successfully",
			"user":    userInfo,
			"token":   token,
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	update := bson.M{
		"$set": userInfo,
	}
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"message": "Failed to save user info: Database update failed"})
	}

	response := map[string]interface{}{
		"user": userInfo,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func GeneratePasswordHandler(c *fiber.Ctx) error {
	var request struct {
		PhoneNumber string `json:"phonenumber"`
		Captcha     string `json:"captcha"`
	}
	if err := c.BodyParser(&request); err != nil {
		var respn model.Response
		respn.Status = "Invalid Request"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	// Validate CAPTCHA
	captchaResponse, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
		"secret":   {"0x4AAAAAAAfj2NjfaHRBhkd2VjcfmRe5gvI"},
		"response": {request.Captcha},
	})
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to verify captcha"
		respn.Response = err.Error()
		return c.Status(fiber.StatusServiceUnavailable).JSON(respn)
	}
	defer captchaResponse.Body.Close()

	var captchaResult struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(captchaResponse.Body).Decode(&captchaResult); err != nil {
		var respn model.Response
		respn.Status = "Failed to decode captcha response"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	if !captchaResult.Success {
		var respn model.Response
		respn.Status = "Unauthorized"
		respn.Response = "Invalid captcha"
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	// Validate phone number
	re := regexp.MustCompile(`^62\d{9,15}$`)
	if !re.MatchString(request.PhoneNumber) {
		var respn model.Response
		respn.Status = "Bad Request"
		respn.Response = "Invalid phone number format"
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Check if phone number exists in the 'user' collection
	userFilter := bson.M{"phonenumber": request.PhoneNumber}
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", userFilter)
	if err != nil {
		var respn model.Response
		respn.Status = "Unauthorized"
		respn.Response = "Phone number not registered"
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	// Generate random password
	randomPassword, err := auth.GenerateRandomPassword(12)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to generate password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(randomPassword)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to hash password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Update or insert the user in the database
	stpFilter := bson.M{"phonenumber": request.PhoneNumber}
	_, err = atdb.GetOneDoc[model.Stp](config.Mongoconn, "stp", stpFilter)
	var responseMessage string

	if err == mongo.ErrNoDocuments {
		// Document not found, insert new one
		newUser := model.Stp{
			PhoneNumber:  request.PhoneNumber,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
		}
		_, err = atdb.InsertOneDoc(config.Mongoconn, "stp", newUser)
		if err != nil {
			var respn model.Response
			respn.Status = "Failed to insert new user"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotModified).JSON(respn)
		}
		responseMessage = "New user created and password generated successfully"
	} else {
		// Document found, update the existing one
		stpUpdate := bson.M{
			"phonenumber": request.PhoneNumber,
			"password":    hashedPassword,
			"createdAt":   time.Now(),
		}
		_, err = atdb.UpdateOneDoc(config.Mongoconn, "stp", stpFilter, stpUpdate)
		if err != nil {
			var respn model.Response
			respn.Status = "Failed to update user"
			respn.Response = err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(respn)
		}
		responseMessage = "User info updated and password generated successfully"
	}

	// Send the random password via WhatsApp
	auth.SendWhatsAppPasswordFiber(c, request.PhoneNumber, randomPassword)

	// Respond with success and the generated password
	response := map[string]interface{}{
		"message":     responseMessage,
		"phonenumber": request.PhoneNumber,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

var (
	rl = auth.NewRateLimiter(1, 5) // 1 request per second, burst of 5
)

func VerifyPasswordHandler(c *fiber.Ctx) error {
	var request struct {
		PhoneNumber string `json:"phonenumber"`
		Password    string `json:"password"`
	}
	if err := c.BodyParser(&request); err != nil {
		var respn model.Response
		respn.Status = "Invalid Request"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Implementasi rate limiting
	limiter := rl.GetLimiter(request.PhoneNumber)
	if !limiter.Allow() {
		var respn model.Response
		respn.Status = "Too Many Requests"
		respn.Response = "Please try again later."
		return c.Status(fiber.StatusTooManyRequests).JSON(respn)
	}

	// Find user in the database
	userFilter := bson.M{"phonenumber": request.PhoneNumber}
	user, err := atdb.GetOneDoc[model.Stp](config.Mongoconn, "stp", userFilter)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to verify password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	// Verify password and expiry
	if time.Now().After(user.CreatedAt.Add(4 * time.Minute)) {
		var respn model.Response
		respn.Status = "Unauthorized"
		respn.Response = "Password Expired"
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to verify password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	// Find user in the 'user' collection
	myiUserFilter := bson.M{"phonenumber": request.PhoneNumber}
	existingUser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", myiUserFilter)
	if err != nil {
		var respn model.Response
		respn.Status = "Unauthorized"
		respn.Response = "Phone number not registered"
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}

	token, err := watoken.EncodeforHours(existingUser.PhoneNumber, existingUser.Name, config.PrivateKey, 18)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to give the token"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	response := map[string]interface{}{
		"message": "Authenticated successfully",
		"token":   token,
		"name":    existingUser.Name,
	}

	// Respond with success
	return c.Status(fiber.StatusOK).JSON(response)
}

func ResendPasswordHandler(c *fiber.Ctx) error {
	var request struct {
		PhoneNumber string `json:"phonenumber"`
	}
	if err := c.BodyParser(&request); err != nil {
		var respn model.Response
		respn.Status = "Invalid Request"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Generate random password
	randomPassword, err := auth.GenerateRandomPassword(12)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to generate password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(randomPassword)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to hash password"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Check if phone number exists in the 'stp' collection
	stpFilter := bson.M{"phonenumber": request.PhoneNumber}
	_, stpErr := atdb.GetOneDoc[model.Stp](config.Mongoconn, "stp", stpFilter)

	if stpErr == mongo.ErrNoDocuments {
		// Document not found, insert new one
		newUser := model.Stp{
			PhoneNumber:  request.PhoneNumber,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
		}
		_, err = atdb.InsertOneDoc(config.Mongoconn, "stp", newUser)
		if err != nil {
			var respn model.Response
			respn.Status = "Failed to insert new user"
			respn.Response = err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(respn)
		}
		responseMessage := "New user created and password generated successfully"

		// Send the random password via WhatsApp
		auth.SendWhatsAppPasswordFiber(c, request.PhoneNumber, randomPassword)

		// Respond with success and the generated password
		response := map[string]interface{}{
			"message":     responseMessage,
			"phonenumber": request.PhoneNumber,
		}
		return c.Status(fiber.StatusOK).JSON(response)
	} else if stpErr != nil {
		var respn model.Response
		respn.Status = "Failed to fetch user info"
		respn.Response = stpErr.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Document found, update the existing one
	stpUpdate := bson.M{
		"phonenumber": request.PhoneNumber,
		"password":    hashedPassword,
		"createdAt":   time.Now(),
	}
	_, err = atdb.UpdateOneDoc(config.Mongoconn, "stp", stpFilter, stpUpdate)
	if err != nil {
		var respn model.Response
		respn.Status = "Failed to update user"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	responseMessage := "User info updated and password generated successfully"

	// Send the random password via WhatsApp
	auth.SendWhatsAppPasswordFiber(c, request.PhoneNumber, randomPassword)

	// Respond with success and the generated password
	response := map[string]interface{}{
		"message":     responseMessage,
		"phonenumber": request.PhoneNumber,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
