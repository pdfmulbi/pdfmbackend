package main

import (
	"net/http"

	"github.com/gocroot/route"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/gocroot/docs"
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
	http.HandleFunc("/", route.URL)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	http.ListenAndServe(":8080", nil)
}
