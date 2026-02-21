package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gcallapi"
	"github.com/gocroot/helper/normalize"
	"github.com/gocroot/helper/watoken"
)

func PostKatalogBuku(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var prj model.Project
	err = c.BodyParser(&prj)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	//mendapatkan user dari token
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : User tidak berhak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//cek apakah user memiliki akses ke project
	project, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prj.ID})
	if err != nil {
		respn.Status = "Error : Data lapak tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//check apakah dia owner
	if project.Owner.PhoneNumber != docuser.PhoneNumber {
		respn.Status = "Error : User bukan owner project tidak berhak"
		respn.Response = "User bukan owner dari project ini"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	//check cover buku apakah kosong  atau engga
	if project.CoverBuku == "" {
		var respn model.Response
		respn.Status = "Belum ada Cover Buku"
		respn.Response = "Mohon upload dahulu cover buku anda pada form yang disediakan"
		return c.Status(fiber.StatusConflict).JSON(respn)
	}

	//publish ke blog katalog
	postingan := strings.ReplaceAll(config.KatalogPost, "##URLCOVERBUKU##", project.CoverBuku)
	postingan = strings.ReplaceAll(postingan, "##SINOPSISBUKU##", project.Description)
	postingan = strings.ReplaceAll(postingan, "##HURUFPERTAMASINOPSIS##", string(project.Description[0]))
	postingan = strings.ReplaceAll(postingan, "##KALIMATPROMOSIBUKU##", project.KalimatPromosi)
	postingan = strings.ReplaceAll(postingan, "##EDITOR##", project.Editor.Name)
	postingan = strings.ReplaceAll(postingan, "##ISBN##", project.ISBN)
	postingan = strings.ReplaceAll(postingan, "##TERBIT##", project.Terbit)
	postingan = strings.ReplaceAll(postingan, "##UKURAN##", project.Ukuran)
	postingan = strings.ReplaceAll(postingan, "##JUMLAHHALAMAN##", project.JumlahHalaman)
	postingan = strings.ReplaceAll(postingan, "##TEBAL##", project.Tebal)
	var daftarpenulisdengantagli string
	for _, penulis := range project.Members {
		daftarpenulisdengantagli += "<li>" + penulis.Name + "</li>"
	}
	postingan = strings.ReplaceAll(postingan, "##DAFTARPENULISDENGANTAGLI##", daftarpenulisdengantagli)
	postingan = strings.ReplaceAll(postingan, "##LINKGRAMED##", project.LinkGramed)
	postingan = strings.ReplaceAll(postingan, "##LINKPLAYBOOK##", project.LinkPlayBook)
	postingan = strings.ReplaceAll(postingan, "##LINKKUBUKU##", project.LinkKubuku)
	postingan = strings.ReplaceAll(postingan, "##LINKMYEDISI##", project.LinkMyedisi)

	bpost, err := gcallapi.PostToBlogger(config.Mongoconn, project.URLKatalog, "3471446342567707906", project.Title, postingan)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal post ke blogger"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	//update data content
	project.URLKatalog = bpost.Id
	project.PATHKatalog = bpost.Url
	//update project data
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": project.ID}, project)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal replaceonedoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(project)
}

func PostDataProject(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var prj model.Project
	err = c.BodyParser(&prj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prj.Owner = docuser
	prj.Secret = watoken.RandomString(48)
	prj.Name = normalize.SetIntoID(prj.Name)
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": prj.Name})
	if err != nil {
		idprj, err := atdb.InsertOneDoc(config.Mongoconn, "project", prj)
		if err != nil {
			var respn model.Response
			respn.Status = "Gagal Insert Database"
			respn.Response = err.Error()
			return c.Status(fiber.StatusNotModified).JSON(respn)
		}
		prj.ID = idprj
		return c.Status(fiber.StatusOK).JSON(prj)
	} else {
		var respn model.Response
		respn.Status = "Error : Nama Project sudah ada"
		respn.Response = existingprj.Name
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
}

func GetDataProject(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"owner._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = "Kakak belum input proyek, silahkan input dulu ya"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

// untuk manager
func GetEditorApprovedProject(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//akses khusus manager
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id, "ismanager": true})
	if err != nil {
		var respn model.Response
		respn.Status = "Akses dibatasi"
		respn.Response = "Anda bukan manager bukupedia"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"isapproved": true})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project yang di approve tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data project yang di approve tidak di temukan"
		respn.Response = "Kakak belum input proyek, silahkan input dulu ya"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

func PutMetaDataProject(c *fiber.Ctx) error {
	// Decode token from header
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	// Decode the project data from the request body
	var prj model.Project
	err = c.BodyParser(&prj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Get user data from the database
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Data user tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	// Check if the project exists and belongs to the user
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prj.ID, "owner._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Project tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// modif hanya meta data saja
	existingprj.Ukuran = prj.Ukuran
	existingprj.JumlahHalaman = prj.JumlahHalaman
	existingprj.Tebal = prj.Tebal

	// Save the updated project back to the database using ReplaceOneDoc
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID}, existingprj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Gagal memperbarui database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Return the updated project
	return c.Status(fiber.StatusOK).JSON(prj)
}

func PutPublishProject(c *fiber.Ctx) error {
	// Decode token from header
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	// Decode the project data from the request body
	var prj model.Project
	err = c.BodyParser(&prj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Get user data from the database and make sure its manager
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id, "ismanager": true})
	if err != nil {
		var respn model.Response
		respn.Status = "Tidak ada akses"
		respn.Response = "Kakak bukan manager"
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	// ambil data project
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prj.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Project tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// modif hanya publisher data saja
	existingprj.ISBN = prj.ISBN
	existingprj.Terbit = prj.Terbit
	existingprj.LinkPlayBook = prj.LinkPlayBook
	existingprj.LinkGramed = prj.LinkGramed
	existingprj.LinkKubuku = prj.LinkMyedisi
	existingprj.LinkMyedisi = prj.LinkMyedisi
	existingprj.LinkDepositPerpusnas = prj.LinkDepositPerpusnas
	existingprj.LinkDepositPerpusda = prj.LinkDepositPerpusda
	existingprj.Manager = docuser

	// Save the updated project back to the database using ReplaceOneDoc
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID}, existingprj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Gagal memperbarui database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Return the updated project
	return c.Status(fiber.StatusOK).JSON(prj)
}

func PutDataProject(c *fiber.Ctx) error {
	// Decode token from header
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	// Decode the project data from the request body
	var prj model.Project
	err = c.BodyParser(&prj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Get user data from the database
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Data user tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	// Check if the project exists and belongs to the user
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prj.ID, "owner._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Project tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// Preserve unmodifiable fields
	prj.ID = existingprj.ID
	prj.Name = existingprj.Name
	prj.Secret = existingprj.Secret
	prj.Owner = existingprj.Owner
	prj.Members = existingprj.Members
	prj.CoverBuku = existingprj.CoverBuku
	prj.DraftBuku = existingprj.DraftBuku
	prj.DraftPDFBuku = existingprj.DraftPDFBuku
	prj.SampulPDFBuku = existingprj.SampulPDFBuku

	// Save the updated project back to the database using ReplaceOneDoc
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID}, prj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Gagal memperbarui database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Return the updated project
	return c.Status(fiber.StatusOK).JSON(prj)
}

func DeleteDataProject(c *fiber.Ctx) error {
	// Dekode token dari header permintaan
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	// Dekode nama proyek dari body permintaan
	var reqBody struct {
		ProjectName string `json:"project_name"`
	}
	err = c.BodyParser(&reqBody)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Dapatkan data pengguna berdasarkan ID dari payload token
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	// Cek apakah proyek dengan nama yang diberikan ada dan dimiliki oleh pengguna
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": reqBody.ProjectName, "owner._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = "Proyek dengan nama tersebut tidak ditemukan atau bukan milik Anda"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// Hapus proyek dari koleksi "project" di MongoDB
	_, err = atdb.DeleteOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Gagal menghapus project"
		respn.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}

	// Berhasil menghapus proyek
	return c.Status(fiber.StatusOK).JSON(map[string]string{"status": "Project berhasil dihapus"})
}

func GetDataMemberProject(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = "Kakak belum menjadi anggota proyek manapun"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

func GetDataEditorProject(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"editor._id": docuser.ID})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = "Kakak belum menjadi anggota proyek manapun"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

func PostDataMemberProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var idprjuser model.Userdomyikado
	err = c.BodyParser(&idprjuser)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuserowner, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": idprjuser.ID, "owner._id": docuserowner.ID})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	docusermember, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": idprjuser.PhoneNumber})
	if err != nil {
		respn.Status = "Error : Data member tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	docusermember.Poin = 0 //set user poin per project, jika baru dimasukkan maka set0 karena belum ada kontribusi di project ini
	rest, err := atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", idprjuser.ID, "members", docusermember)
	if err != nil {
		respn.Status = "Error : Gagal menambahkan member ke project"
		respn.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	if rest.ModifiedCount == 0 {
		respn.Status = "Error : Gagal menambahkan member ke project"
		respn.Response = "Tidak ada perubahan pada dokumen proyek"
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprj)
}

func PostDataEditorProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var idprjuser model.Project
	err = c.BodyParser(&idprjuser)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuserowner, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": idprjuser.ID, "owner._id": docuserowner.ID})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	docusermember, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"_id": idprjuser.Editor.ID})
	if err != nil {
		respn.Status = "Error : Data editor tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	docusermember.Poin = 0 //set user poin per project, jika baru dimasukkan maka set0 karena belum ada kontribusi di project ini
	existingprj.Editor = docusermember
	existingprj.IsApproved = false
	//update project
	// Save the updated project back to the database using ReplaceOneDoc
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID}, existingprj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Gagal memperbarui database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprj)
}

func PUtApprovedEditorProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var idprjuser model.Project
	err = c.BodyParser(&idprjuser)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docusereditor, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": idprjuser.ID, "editor.phonenumber": docusereditor.PhoneNumber})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	existingprj.IsApproved = true
	//update project
	// Save the updated project back to the database using ReplaceOneDoc
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "project", primitive.M{"_id": existingprj.ID}, existingprj)
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Gagal memperbarui database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprj)
}

func PostDataMenuProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var idprjuser model.MenuItem
	err = c.BodyParser(&idprjuser)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuserowner, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": idprjuser.IDDatabase, "owner._id": docuserowner.ID})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	//bikin insert menu
	//idprjuser.IDDatabase = primitive.NilObjectID
	rest, err := atdb.AddDocToArray[model.MenuItem](config.Mongoconn, "project", idprjuser.IDDatabase, "menu", idprjuser)
	if err != nil {
		respn.Status = "Error : Gagal menambahkan menu ke lapak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	if rest.ModifiedCount == 0 {
		respn.Status = "Error : Gagal menambahkan member ke project"
		respn.Response = "Tidak ada perubahan pada dokumen proyek"
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprj)
}

func DeleteDataMenuProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	var requestPayload struct {
		ProjectName string `json:"project_name"`
		MenuID      string `json:"menu_id"`
	}

	err = c.BodyParser(&requestPayload)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	docuserowner, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": requestPayload.ProjectName, "owner._id": docuserowner.ID})
	if err != nil {
		respn.Status = "Error : Data project tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// Menghapus member dari project
	menuToDelete := model.MenuItem{ID: requestPayload.MenuID}
	rest, err := atdb.DeleteDocFromArray[model.MenuItem](config.Mongoconn, "project", existingprj.ID, "menu", menuToDelete)
	if err != nil {
		respn.Status = "Error : Gagal menghapus menu dari lapak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	if rest.ModifiedCount == 0 {
		respn.Status = "Error : Gagal menghapus menu dari lapak"
		respn.Response = "Tidak ada perubahan pada dokumen proyek:" + menuToDelete.ID
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}

	return c.Status(fiber.StatusOK).JSON(existingprj)
}

func DeleteDataMemberProject(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}

	var requestPayload struct {
		ProjectName string `json:"project_name"`
		PhoneNumber string `json:"phone_number"`
	}

	err = c.BodyParser(&requestPayload)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	docuserowner, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data owner tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	existingprj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": requestPayload.ProjectName, "owner._id": docuserowner.ID})
	if err != nil {
		respn.Status = "Error : Data project tidak ditemukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}

	// Menghapus member dari project
	memberToDelete := model.Userdomyikado{PhoneNumber: requestPayload.PhoneNumber}
	rest, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", existingprj.ID, "members", memberToDelete)
	if err != nil {
		respn.Status = "Error : Gagal menghapus member dari project"
		respn.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	if rest.ModifiedCount == 0 {
		respn.Status = "Error : Gagal menghapus member dari project"
		respn.Response = "Tidak ada perubahan pada dokumen proyek"
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}

	return c.Status(fiber.StatusOK).JSON(existingprj)
}
