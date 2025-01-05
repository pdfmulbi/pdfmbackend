package controller

import (
	"encoding/json"
	// "io"
	// "log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	// "github.com/gocroot/helper/fpdf"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

// RegisterHandler menghandle permintaan registrasi admin.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var registrationData model.PdfmRegistration

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&registrationData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if registrationData.Password != registrationData.ConfirmPassword {
		http.Error(w, "Password tidak sesuai", http.StatusBadRequest)
		return
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "user", registrationData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Registrasi berhasil"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser mengambil informasi user dari database berdasarkan email dan password.
func GetUser(respw http.ResponseWriter, req *http.Request) {
	var loginDetails model.PdfmUser
	if err := json.NewDecoder(req.Body).Decode(&loginDetails); err != nil {
		at.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	var user model.PdfmUser
	filter := bson.M{"email": loginDetails.Email, "password": loginDetails.Password}
	user, err := atdb.GetOneDoc[model.PdfmUser](config.Mongoconn, "user", filter)
	if err != nil {
		at.WriteJSON(respw, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	at.WriteJSON(respw, http.StatusOK, user)
}

// MergePDFHandler handles the merging of multiple uploaded PDF files
// func MergePDFHandler(w http.ResponseWriter, r *http.Request) {
//     // Parse the multipart form
//     err := r.ParseMultipartForm(10 << 20) // 10 MB max upload
//     if err != nil {
//         http.Error(w, "Failed to parse form", http.StatusBadRequest)
//         return
//     }

//     // Retrieve uploaded files
//     files := r.MultipartForm.File["files"]
//     if len(files) < 2 {
//         http.Error(w, "Please upload at least 2 PDF files", http.StatusBadRequest)
//         return
//     }

//     // Collect file contents
//     var pdfBytes [][]byte
//     for _, fileHeader := range files {
//         file, err := fileHeader.Open()
//         if err != nil {
//             http.Error(w, "Failed to open file", http.StatusInternalServerError)
//             return
//         }
//         defer file.Close()

//         // Read the file content into memory
//         fileContent, err := io.ReadAll(file)
//         if err != nil {
//             http.Error(w, "Failed to read file", http.StatusInternalServerError)
//             return
//         }
//         pdfBytes = append(pdfBytes, fileContent)
//     }

//     // Merge PDFs
//     var mergedPDF []byte
//     for i := 0; i < len(pdfBytes)-1; i++ {
//         if i == 0 {
//             mergedPDF, err = fpdf.MergePDFBytes(pdfBytes[i], pdfBytes[i+1])
//         } else {
//             mergedPDF, err = fpdf.MergePDFBytes(mergedPDF, pdfBytes[i+1])
//         }
//         if err != nil {
//             log.Printf("Merge error: %v", err)
//             http.Error(w, "Failed to merge PDF files", http.StatusInternalServerError)
//             return
//         }
//     }

//     // Serve the merged PDF file
//     w.Header().Set("Content-Type", "application/pdf")
//     w.Header().Set("Content-Disposition", `attachment; filename="merged_output.pdf"`)
//     w.Write(mergedPDF)
// }
