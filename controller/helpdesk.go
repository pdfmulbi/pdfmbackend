package controller

import (
	"strconv"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// pindahkan task dari to do ke doing
func PutTaskUser(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	var task report.TaskList
	if err := c.BodyParser(&task); err != nil {
		respn.Status = "Error : Body Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Body Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	taskuser, err := atdb.GetOneDoc[report.TaskList](config.Mongoconn, "tasklist", bson.M{"_id": task.ID})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(taskuser)
	}
	insertid, err := atdb.InsertOneDoc(config.Mongoconn, "taskdoing", taskuser)
	if err != nil {
		respn.Status = "Error : Gagal insert ke doing"
		respn.Info = insertid.Hex()
		respn.Location = "InsertOneDoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	rest, err := atdb.DeleteOneDoc(config.Mongoconn, "tasklist", bson.M{"_id": task.ID})
	if err != nil {
		respn.Status = "Error : Gagal hapus di tasklist"
		respn.Info = strconv.FormatInt(rest.DeletedCount, 10)
		respn.Location = "DeleteOneDoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	respn.Info = strconv.FormatInt(rest.DeletedCount, 10)
	respn.Status = insertid.Hex()
	return c.Status(fiber.StatusOK).JSON(respn)
}

// pindahkan task dari doing ke done
func PostTaskUser(c *fiber.Ctx) error {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	var task report.TaskList
	if err := c.BodyParser(&task); err != nil {
		respn.Status = "Error : Body Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Body Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	taskuser, err := atdb.GetOneDoc[report.TaskList](config.Mongoconn, "taskdoing", bson.M{"_id": task.ID})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(taskuser)
	}
	insertid, err := atdb.InsertOneDoc(config.Mongoconn, "taskdone", taskuser)
	if err != nil {
		respn.Status = "Error : Gagal insert ke taskdone"
		respn.Info = insertid.Hex()
		respn.Location = "InsertOneDoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	rest, err := atdb.DeleteOneDoc(config.Mongoconn, "taskdoing", bson.M{"_id": task.ID})
	if err != nil {
		respn.Status = "Error : Gagal hapus di taskdoing"
		respn.Info = strconv.FormatInt(rest.DeletedCount, 10)
		respn.Location = "DeleteOneDoc"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	respn.Info = strconv.FormatInt(rest.DeletedCount, 10)
	respn.Status = insertid.Hex()
	return c.Status(fiber.StatusOK).JSON(respn)
}

func GetHelpdeskAll(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	//melakukan pengambilan data belum terlayani
	filterbelumterlayani := bson.M{
		"terlayani": bson.M{
			"$exists": false,
		},
		"user.phonenumber": docuser.PhoneNumber,
	}
	userbelumterlayani, _ := atdb.GetAllDoc[[]model.Laporan](config.Mongoconn, "helpdeskuser", filterbelumterlayani)
	//melakukan pengambilan data sudah terlayani
	filtersudahterlayani := bson.M{
		"terlayani": bson.M{
			"$exists": true,
		},
		"user.phonenumber": docuser.PhoneNumber,
	}
	usersudahterlayani, _ := atdb.GetAllDoc[[]model.Laporan](config.Mongoconn, "helpdeskuser", filtersudahterlayani)
	//melakukan pengambilan semu data user terlayani atau belum
	filtersemua := bson.M{
		"user.phonenumber": docuser.PhoneNumber,
	}
	usersemua, _ := atdb.GetAllDoc[[]model.Laporan](config.Mongoconn, "helpdeskuser", filtersemua)
	rekap := model.HelpdeskRekap{
		ToDo: len(userbelumterlayani),
		Done: len(usersudahterlayani),
		All:  len(usersemua),
	}
	return c.Status(fiber.StatusOK).JSON(rekap)
}

func GetLatestHelpdeskMasuk(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	//melakukan pengambilan data belum terlayani
	filterbelumterlayani := bson.M{
		"terlayani": bson.M{
			"$exists": false,
		},
		"user.phonenumber": docuser.PhoneNumber,
	}
	userbelumterlayani, err := atdb.GetOneLatestDoc[model.Laporan](config.Mongoconn, "helpdeskuser", filterbelumterlayani)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(userbelumterlayani)
	}
	return c.Status(fiber.StatusOK).JSON(userbelumterlayani)
}

func GetLatestHelpdeskSelesai(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	//melakukan pengambilan data sudah terlayani
	filtersudahterlayani := bson.M{
		"terlayani": bson.M{
			"$exists": true,
		},
		"user.phonenumber": docuser.PhoneNumber,
	}
	userbelumterlayani, err := atdb.GetOneLatestDoc[model.Laporan](config.Mongoconn, "helpdeskuser", filtersudahterlayani)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(userbelumterlayani)
	}
	return c.Status(fiber.StatusOK).JSON(userbelumterlayani)
}

func GetTaskDone(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	taskdoing, err := atdb.GetOneLatestDoc[report.TaskList](config.Mongoconn, "taskdone", bson.M{"phonenumber": docuser.PhoneNumber})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(taskdoing)
	}
	return c.Status(fiber.StatusOK).JSON(taskdoing)
}
