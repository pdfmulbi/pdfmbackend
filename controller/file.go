package controller

import (
	"encoding/base64"
	"io"
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/dokped"
	"github.com/gocroot/helper/ghupload"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AksesFileRepoDraft(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFileBase64 := c.Params("id") // Assuming the param name is "id", you might need to adjust this based on route.go
	if pathFileBase64 == "" {
		// Fallback if at.GetParam was doing something else, though c.Params should cover it
		pathFileBase64 = c.Query("id") // Or however 'id' is passed
	}

	// Decode string dari Base64
	decoded, err := base64.StdEncoding.DecodeString(pathFileBase64)
	if err != nil {
		respn.Status = "Error : decoding base64"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFile := string(decoded)
	pathslice := strings.Split(pathFile, "/")
	namaprj := pathslice[0]
	//cek apakah user memiliki akses ke project
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": namaprj})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if (prj.Owner.PhoneNumber != docuser.PhoneNumber) && (prj.Editor.PhoneNumber != docuser.PhoneNumber) {
		if !docuser.IsManager { //kalo bukan manager maka ga punya akses dong
			respn.Status = "Error : User bukan owner project tidak berhak"
			respn.Response = "User bukan owner dari project ini"
			return c.Status(fiber.StatusNotImplemented).JSON(respn)
		}
	}

	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	filecontent, err := ghupload.GithubGetFile(config.GHAccessToken, githubOrg, githubRepo, pathFile)
	if err != nil {
		respn.Status = "Error : Data tidak bisa diambil dari github"
		respn.Info = githubOrg + "/" + githubRepo
		respn.Location = pathFile
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	c.Set("Content-Disposition", "attachment; filename=\"file.ext\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.Status(fiber.StatusOK).Send(filecontent)
}

func GetFileDraftSPK(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFileBase64 := c.Params("id")
	if pathFileBase64 == "" {
		pathFileBase64 = c.Query("id")
	}
	// Decode string dari Base64
	decoded, err := base64.StdEncoding.DecodeString(pathFileBase64)
	if err != nil {
		respn.Status = "Error : decoding base64"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	namaprj := string(decoded)
	//cek apakah user memiliki akses ke project
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": namaprj})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	filecontent, err := dokped.GenerateSPK(prj, config.AESKey)
	if err != nil {
		respn.Status = "Error : Dokumen gagal di generate"
		respn.Info = prj.Name
		respn.Location = prj.ID.Hex()
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	c.Set("Content-Disposition", "attachment; filename=\"file.ext\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.Status(fiber.StatusOK).Send(filecontent)
}

func GetFileDraftSPKT(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFileBase64 := c.Params("id")
	if pathFileBase64 == "" {
		pathFileBase64 = c.Query("id")
	}
	// Decode string dari Base64
	decoded, err := base64.StdEncoding.DecodeString(pathFileBase64)
	if err != nil {
		respn.Status = "Error : decoding base64"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	namaprj := string(decoded)
	//cek apakah user memiliki akses ke project
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": namaprj})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	filecontent, err := dokped.GenerateSPKT(prj, config.AESKey)
	if err != nil {
		respn.Status = "Error : Dokumen gagal di generate"
		respn.Info = prj.Name
		respn.Location = prj.ID.Hex()
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	c.Set("Content-Disposition", "attachment; filename=\"file.ext\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.Status(fiber.StatusOK).Send(filecontent)
}

func GetFileDraftSPI(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFileBase64 := c.Params("id")
	if pathFileBase64 == "" {
		pathFileBase64 = c.Query("id")
	}
	// Decode string dari Base64
	decoded, err := base64.StdEncoding.DecodeString(pathFileBase64)
	if err != nil {
		respn.Status = "Error : decoding base64"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	pathFile := string(decoded)
	pathslice := strings.Split(pathFile, "/")
	namaprj := pathslice[0]
	//cek apakah user memiliki akses ke project
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": namaprj})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//ambil surat pengantar
	filecontentpengantar, err := dokped.GenerateSPI(prj, config.AESKey)
	if err != nil {
		respn.Status = "Error : Dokumen gagal di generate"
		respn.Info = prj.Name
		respn.Location = prj.ID.Hex()
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	c.Set("Content-Disposition", "attachment; filename=\"file.ext\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.Status(fiber.StatusOK).Send(filecontentpengantar)
}

func UploadProfilePictureHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	userdoc, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	fileHeader, err := c.FormFile("profpic")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "profile"
	pathFile := "picture/" + userdoc.ID.Hex() + "/" + userdoc.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	userdoc.ProfilePicture = "https://raw.githubusercontent.com/" + githubOrg + "/" + githubRepo + "/main/" + *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "user", bson.M{"_id": userdoc.ID}, userdoc)

	// Respond with success message
	respn.Info = userdoc.ID.Hex()
	respn.Location = userdoc.ProfilePicture
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func FileUploadWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("profpic")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "img"
	pathFile := prj.Name + "/menu/" + hashedFileName + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Respond with success message
	respn.Info = hashedFileName
	respn.Location = "/" + githubRepo + "/" + *content.Content.Path
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadCoverBukuWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("coverbuku")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "katalog"
	pathFile := prj.Name + "/cover/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	prj.CoverBuku = "https://raw.githubusercontent.com/" + githubOrg + "/" + githubRepo + "/main/" + *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.CoverBuku
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadDraftBukuWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("draftbuku")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	pathFile := prj.Name + "/draft/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	//prj.DraftBuku = "https://raw.githubusercontent.com/" + githubOrg + "/" + githubRepo + "/main/" + *content.Content.Path
	prj.DraftBuku = *content.Content.Path //karena repo private
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.DraftBuku
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadDraftBukuPDFWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("draftpdfbuku")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	pathFile := prj.Name + "/pdf/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	prj.DraftPDFBuku = *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.DraftPDFBuku
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadSPKPDFWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("spk")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	pathFile := prj.Name + "/spk/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	prj.SPK = *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.SPK
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadSPIPDFWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("spi")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	pathFile := prj.Name + "/spi/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	prj.SPI = *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.SPI
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}

func UploadSampulBukuPDFWithParamFileHandler(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjid := c.Params("id")
	if prjid == "" {
		prjid = c.Query("id")
	}
	objectId, _ := primitive.ObjectIDFromHex(prjid)
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if prj.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	fileHeader, err := c.FormFile("sampulpdfbuku")
	if err != nil {
		respn.Status = "Error : File tidak ada"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	file, err := fileHeader.Open()
	if err != nil {
		respn.Status = "Error : Gagal membuka file"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer file.Close()
	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		respn.Status = "Error : File tidak bisa dibaca"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	// Calculate hash of the file content
	//hashedFileName := ghupload.CalculateHash(fileContent)

	// Get GitHub credentials and other details from the request or environment variables
	GitHubAccessToken := config.GHAccessToken
	GitHubAuthorName := "Rolly Maulana Awangga"
	GitHubAuthorEmail := "awangga@gmail.com"
	githubOrg := "penerbitbukupedia"
	githubRepo := "draft"
	pathFile := prj.Name + "/sampul/" + prj.ID.Hex() + fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):] // Append the original file extension
	replace := true

	// Use GithubUpload function to upload the file to GitHub
	content, _, err := ghupload.GithubUpload(GitHubAccessToken, GitHubAuthorName, GitHubAuthorEmail, fileContent, githubOrg, githubRepo, pathFile, replace)
	if err != nil {
		respn.Status = "Error : File tidak bisa diupload ke github"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	//update data profpic
	prj.SampulPDFBuku = *content.Content.Path
	atdb.ReplaceOneDoc(config.Mongoconn, "project", bson.M{"_id": prj.ID}, prj)

	// Respond with success message
	respn.Info = prj.ID.Hex()
	respn.Location = prj.SampulPDFBuku
	respn.Response = *content.Content.URL
	respn.Status = *content.Content.HTMLURL
	return c.Status(fiber.StatusOK).JSON(respn)
}
