package main

import (
	"log"

	"github.com/gocroot/route"
	"github.com/gofiber/fiber/v2"

	_ "github.com/gocroot/docs"
	"github.com/gofiber/swagger"
)

// @title PDF Merger API
// @version 1.0
// @description API untuk mengelola PDF (Merge, Compress, dll)
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host asia-southeast2-personalsmz.cloudfunctions.net
// @BasePath /pdfmerger
// @schemes https

func main() {
	app := fiber.New()

	app.Get("/swagger/*", swagger.HandlerDefault)
	route.RegisterRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
