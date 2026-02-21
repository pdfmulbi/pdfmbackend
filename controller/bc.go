package controller

import (
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDataSenders(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.SenderDasboard](config.Mongoconn, "sender", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = "Kakak belum input sender, silahkan input dulu ya"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

func GetDataSendersTerblokir(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	existingprjs, err := atdb.GetAllDoc[[]model.SenderDasboard](config.Mongoconn, "bin", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = "Kakak belum input sender, silahkan input dulu ya"
		return c.Status(fiber.StatusNotFound).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(existingprjs)
}

func GetRekapBlast(c *fiber.Ctx) error {
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
	docuser.Name = payload.Alias
	//melakukan pengambilan data belum terlayani
	// Menghitung jumlah dokumen dalam koleksi
	countqueue, err := atdb.GetCountDoc(config.Mongoconn, "peserta", bson.M{})
	if err != nil {
		respn.Status = "Error : penghitungan data queue"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}
	countsent, err := atdb.GetCountDoc(config.Mongoconn, "sent", bson.M{})
	if err != nil {
		respn.Status = "Error : penghitungan data sent"
		respn.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(respn)
	}

	rekap := model.HelpdeskRekap{
		ToDo: int(countqueue),
		Done: int(countsent),
		All:  int(countqueue + countsent),
	}
	return c.Status(fiber.StatusOK).JSON(rekap)
}

// melakukan pendaftaran nomor blast dengan pengecekan apakah suda link device
func PutNomorBlast(c *fiber.Ctx) error {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeaderFiber(c)
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var newbot model.SenderDasboard
	if err := c.BodyParser(&newbot); err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		return c.Status(fiber.StatusNotFound).JSON(docuser)
	}
	docuser.Name = payload.Alias
	//check apakah user sudah linked device atau belum
	if docuser.LinkedDevice == "" {
		var respn model.Response
		respn.Status = "Error : User belum melakukan linked device"
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	//check validasi nomor inputan dengan mengirimkan pesan
	newmsg := model.SendText{
		To:       newbot.Phonenumber,
		IsGroup:  false,
		Messages: "Hai hai... permisi... nomor ini saya daftarkan untuk broadcast ya mohon clear semua notifikasi, karena nanti ada notifikasi kode untuk linked device ke wa business, mohon notif wa business nya juga di allow di handphone",
	}
	httpstatuscode, _, err := atapi.PostStructWithToken[model.Response]("token", config.WAAPIToken, newmsg, config.WAAPIMessage)
	if httpstatuscode != 200 || err != nil {
		var respn model.Response
		respn.Status = "Error : Nomor yang diinputkan tidak valid"
		if err != nil {
			respn.Response = err.Error()
		}
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	//request linked device nomor yang didaftarkan
	tokenbotbaru, err := watoken.Encode(newbot.Phonenumber, config.PrivateKey)
	if err != nil {
		return c.Status(fiber.StatusMisdirectedRequest).JSON(docuser)
	}
	hcode, qrstat, err := atapi.Get[model.QRStatus](config.WAAPIGetDevice + tokenbotbaru)
	if err != nil {
		return c.Status(fiber.StatusMisdirectedRequest).JSON(docuser)
	}
	if hcode != fiber.StatusOK {
		return c.Status(fiber.StatusFailedDependency).JSON(docuser)
	}
	//insert ke coll sender dan profile
	contohsender, err := atdb.GetOneLatestDoc[itmodel.Profile](config.Mongoconn, "sender", bson.M{})
	if err != nil {
		return c.Status(fiber.StatusFailedDependency).JSON(docuser)
	}
	contohsender.Botname = docuser.Name
	contohsender.Phonenumber = newbot.Phonenumber
	contohsender.Triggerword = newbot.Triggerword
	// Temukan posisi terakhir dari '/'
	lastSlashIndex := strings.LastIndex(contohsender.URL, "/")
	// Potong URL hingga posisi terakhir '/'
	baseURL := contohsender.URL[:lastSlashIndex+1]
	// Gabungkan baseURL dengan phonenumber
	contohsender.URL = baseURL + newbot.Phonenumber
	contohsender.Token, err = watoken.EncodeforHours(newbot.Phonenumber, docuser.Name, config.PrivateKey, 43830)
	if err != nil {
		return c.Status(fiber.StatusFailedDependency).JSON(docuser)
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "sender", contohsender)
	if err != nil {
		return c.Status(fiber.StatusFailedDependency).JSON(docuser)
	}
	//daftarkan ke webhook agar bot aktif dan insert kan ke profile
	whdt := model.Webhook{
		URL:    contohsender.URL,
		Secret: contohsender.Secret,
	}
	httpstatuscode, _, err = atapi.PostStructWithToken[model.Response]("token", contohsender.Token, whdt, config.WAAPIGetToken)
	if httpstatuscode != 200 || err != nil {
		var respn model.Response
		respn.Status = "Error : Gagal mendaftarkan ke webhook"
		if err != nil {
			respn.Response = err.Error()
		}
		return c.Status(fiber.StatusExpectationFailed).JSON(respn)
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "profile", contohsender)
	if err != nil {
		return c.Status(fiber.StatusFailedDependency).JSON(docuser)
	}

	//jika belum linked device status = true maka kasih code
	//kirim kode ke wa dari user
	if qrstat.Status { //true jika belum linked device
		newmsg = model.SendText{
			To:       newbot.Phonenumber,
			IsGroup:  false,
			Messages: "Masukkan kode: *" + qrstat.Code + "*\n" + qrstat.Message + "\nUntuk nomor" + qrstat.PhoneNumber,
		}
		httpstatuscode, _, err = atapi.PostStructWithToken[model.Response]("token", config.WAAPIToken, newmsg, config.WAAPITextMessage)
		if httpstatuscode != 200 || err != nil {
			var respn model.Response
			respn.Status = "Error : Nomor yang diinputkan tidak valid"
			if err != nil {
				respn.Response = err.Error()
			}
			return c.Status(fiber.StatusExpectationFailed).JSON(respn)
		}
	}
	//kirim ke frontend
	return c.Status(fiber.StatusOK).JSON(qrstat)
}
