package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gcallapi"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
)

func GetDataUserFromApi(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeaderFiber(c)
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	userdt := lms.GetDataFromAPI(payload.Id)
	if userdt.Data.Fullname == "" {
		return c.Status(fiber.StatusNotFound).JSON(userdt)
	}
	return c.Status(fiber.StatusOK).JSON(userdt)
}

func GetDataUser(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeaderFiber(c)
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	return c.Status(fiber.StatusOK).JSON(docuser)
}

// melakukan pengecekan apakah suda link device klo ada generate token 5tahun
func PutTokenDataUser(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeaderFiber(c)
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetDevice + at.GetLoginFromHeaderFiber(c))
	if err != nil {
		return c.Status(fiber.StatusMisdirectedRequest).JSON(docuser)
	}
	if hcode == http.StatusOK && !qrstat.Status {
		docuser.LinkedDevice, err = watoken.EncodeforHours(docuser.PhoneNumber, docuser.Name, config.PrivateKey, 43830)
		if err != nil {
			return c.Status(fiber.StatusFailedDependency).JSON(docuser)
		}
	} else {
		docuser.LinkedDevice = ""
	}
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		return c.Status(fiber.StatusExpectationFailed).JSON(docuser)
	}
	return c.Status(fiber.StatusOK).JSON(docuser)
}

func PostDataUser(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var usr model.Userdomyikado
	err = c.BodyParser(&usr)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	//pengecekan isian usr
	if usr.NIK == "" || usr.Pekerjaan == "" || usr.AlamatRumah == "" || usr.AlamatKantor == "" {
		var respn model.Response
		respn.Status = "Isian tidak lengkap"
		respn.Response = "Mohon isi lengkap NIK, Pekerjaan, dan kedua alamat"
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		usr.PhoneNumber = payload.Id
		usr.Name = payload.Alias
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
		if err != nil {
			var respn model.Response
			respn.Status = "Gagal Insert Database"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotModified).JSON(respn)
		}
		usr.ID = idusr
		return c.Status(fiber.StatusOK).JSON(usr)
	}
	//jika email belum gsign maka gsign dulu
	if docuser.Email == "" {
		var respn model.Response
		respn.Status = "Email belum terdaftar"
		respn.Response = "Mohon lakukan google sign in dahulu agar email bisa terdaftar"
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuser.NIK = usr.NIK
	docuser.Pekerjaan = usr.Pekerjaan
	docuser.AlamatRumah = usr.AlamatRumah
	docuser.AlamatKantor = usr.AlamatKantor
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal replaceonedoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	//melakukan update di seluruh member project
	//ambil project yang member sebagai anggota
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil { //kalo belum jadi anggota project manapun aman langsung ok
		return c.Status(fiber.StatusOK).JSON(docuser)
	}
	if len(existingprjs) == 0 { //kalo belum jadi anggota project manapun aman langsung ok
		return c.Status(fiber.StatusOK).JSON(docuser)
	}
	//loop keanggotaan setiap project dan menggantinya dengan doc yang terupdate
	for _, prj := range existingprjs {
		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Data project tidak di temukan"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotFound).JSON(respn)
		}
		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Gagal menambahkan member ke project"
			respn.Response = err.Error()
			return c.Status(fiber.StatusExpectationFailed).JSON(respn)
		}

	}

	return c.Status(fiber.StatusOK).JSON(docuser)
}

func PostDataBioUser(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var usr model.Userdomyikado
	err = c.BodyParser(&usr)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		usr.PhoneNumber = payload.Id
		usr.Name = payload.Alias
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
		if err != nil {
			var respn model.Response
			respn.Status = "Gagal Insert Database"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotModified).JSON(respn)
		}
		usr.ID = idusr
		return c.Status(fiber.StatusOK).JSON(usr)
	}
	//check profpic apakah kosong  atau engga
	if docuser.ProfilePicture == "" {
		var respn model.Response
		respn.Status = "Belum ada Profile Picture"
		respn.Response = "Mohon upload dahulu profile picture anda pada form yang disediakan"
		return c.Status(fiber.StatusConflict).JSON(respn)
	}

	//publish ke blog
	postingan := strings.ReplaceAll(config.ProfPost, "##PROFPIC##", docuser.ProfilePicture)
	postingan = strings.ReplaceAll(postingan, "##BIO##", usr.Bio)
	bpost, err := gcallapi.PostToBlogger(config.Mongoconn, docuser.URLBio, "2587271685863777988", docuser.Name, postingan)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal post ke blogger"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	//update data content
	docuser.URLBio = bpost.Id
	docuser.PATHBio = bpost.Url
	docuser.Bio = usr.Bio
	//update user data
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal replaceonedoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	//melakukan update di seluruh member project
	//ambil project yang member sebagai anggota
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil { //kalo belum jadi anggota project manapun aman langsung ok
		return c.Status(fiber.StatusOK).JSON(docuser)
	}
	if len(existingprjs) == 0 { //kalo belum jadi anggota project manapun aman langsung ok
		return c.Status(fiber.StatusOK).JSON(docuser)
	}
	//loop keanggotaan setiap project dan menggantinya dengan doc yang terupdate
	for _, prj := range existingprjs {
		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Data project tidak di temukan"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotFound).JSON(respn)
		}
		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Gagal menambahkan member ke project"
			respn.Response = err.Error()
			return c.Status(fiber.StatusExpectationFailed).JSON(respn)
		}

	}

	return c.Status(fiber.StatusOK).JSON(docuser)
}

func PostDataUserFromWA(c *fiber.Ctx) error {
	var resp itmodel.Response
	prof, err := whatsauth.GetAppProfile(c.Query("url"), config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	if at.GetSecretFromHeaderFiber(c) != prof.Secret {
		resp.Response = "Salah secret: " + at.GetSecretFromHeaderFiber(c)
		return c.Status(fiber.StatusUnauthorized).JSON(resp)
	}
	var usr model.Userdomyikado
	err = c.BodyParser(&usr)
	if err != nil {
		resp.Response = "Error : Body tidak valid"
		resp.Info = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": usr.PhoneNumber})
	if err != nil {
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
		if err != nil {
			resp.Response = "Gagal Insert Database"
			resp.Info = err.Error()
			return c.Status(fiber.StatusNotModified).JSON(resp)
		}
		resp.Info = idusr.Hex()
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	docuser.Name = usr.Name
	docuser.Email = usr.Email
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": usr.PhoneNumber}, docuser)
	if err != nil {
		resp.Response = "Gagal replaceonedoc"
		resp.Info = err.Error()
		return c.Status(fiber.StatusConflict).JSON(resp)
	}
	//melakukan update di seluruh member project
	//ambil project yang member sebagai anggota
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil { //kalo belum jadi anggota project manapun aman langsung ok
		resp.Response = "belum terdaftar di project manapun"
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	if len(existingprjs) == 0 { //kalo belum jadi anggota project manapun aman langsung ok
		resp.Response = "belum terdaftar di project manapun"
		return c.Status(fiber.StatusOK).JSON(resp)
	}
	//loop keanggotaan setiap project dan menggantinya dengan doc yang terupdate
	for _, prj := range existingprjs {
		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
		if err != nil {
			resp.Response = "Error : Data project tidak di temukan"
			resp.Info = err.Error()
			return c.Status(fiber.StatusNotFound).JSON(resp)
		}
		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
		if err != nil {
			resp.Response = "Error : Gagal menambahkan member ke project"
			resp.Info = err.Error()
			return c.Status(fiber.StatusExpectationFailed).JSON(resp)
		}

	}
	resp.Info = docuser.ID.Hex()
	resp.Response = docuser.Email
	return c.Status(fiber.StatusOK).JSON(resp)
}

func ApproveBimbinganbyPoin(c *fiber.Ctx) error {
	noHp := c.Get("nohp")
	if noHp == "" {
		return c.Status(fiber.StatusForbidden).SendString("No valid phone number found")
	}

	var requestData struct {
		NIM   string `json:"nim"`
		Topik string `json:"topik"`
	}
	err := c.BodyParser(&requestData)
	if err != nil || requestData.NIM == "" || requestData.Topik == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body or NIM/Topik not provided")
	}

	// Get the API URL from the database
	var conf model.Config
	err = config.Mongoconn.Collection("config").FindOne(context.TODO(), bson.M{"phonenumber": "62895601060000"}).Decode(&conf)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Mohon maaf ada kesalahan dalam pengambilan config di database: " + err.Error())
	}

	// Prepare the request body
	requestBody, err := json.Marshal(map[string]string{
		"nim":   requestData.NIM,
		"topik": requestData.Topik,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membuat request body: " + err.Error())
	}

	// Create and send the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", conf.ApproveBimbinganURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal membuat request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("nohp", noHp)

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal mengirim request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return c.Status(fiber.StatusNotFound).SendString("Token tidak ditemukan! Silahkan Login Kembali")
		case http.StatusForbidden:
			return c.Status(fiber.StatusForbidden).SendString("Gagal, Bimbingan telah disetujui!")
		default:
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Gagal approve bimbingan, status code: %d", resp.StatusCode))
		}
	}

	var responseMap map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal memproses response: " + err.Error())
	}

	// Kurangi poin berdasarkan nomor telepon yang ada di response
	phonenumber := responseMap["no_hp"]
	_, err = report.KurangPoinUserbyPhoneNumber(config.Mongoconn, phonenumber, 13.0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal mengurangi poin: " + err.Error())
	}

	// Get updated user data to return the current points
	usr, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal mengambil data pengguna: " + err.Error())
	}

	// Add the current points to the response
	responseMap["message"] = "Bimbingan berhasil di approve!"
	responseMap["status"] = "success"
	responseMap["poin_mahasiswa"] = fmt.Sprintf("Poin mahasiswa telah berkurang menjadi: %f", usr.Poin)

	return c.Status(fiber.StatusOK).JSON(responseMap)
}
