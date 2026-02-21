package main

import (
	"log"

	"github.com/gocroot/route"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// Logging Middleware
	app.Use(logger.New())

	// Register Routes
	route.RegisterRoutes(app)

	// Start Server
	log.Fatal(app.Listen(":3000"))
}
