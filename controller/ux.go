package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gcallapi"
	"github.com/gocroot/helper/kimseok"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/helper/phone"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/tiket"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/mod/helpdesk"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostTaskList(c *fiber.Ctx) error {
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
	var tasklists []report.TaskList
	err = c.BodyParser(&tasklists)
	if err != nil {
		resp.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	docusr, err := atdb.GetOneLatestDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": tasklists[0].PhoneNumber})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		return c.Status(fiber.StatusForbidden).JSON(resp)
	}
	lapuser, err := atdb.GetOneLatestDoc[model.Laporan](config.Mongoconn, "uxlaporan", primitive.M{"_id": tasklists[0].LaporanID})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		return c.Status(fiber.StatusForbidden).JSON(resp)
	}
	for _, task := range tasklists {
		task.ProjectID = lapuser.Project.ID
		task.ProjectName = lapuser.Project.Name
		task.Email = docusr.Email
		task.UserID = docusr.ID
		task.MeetID = lapuser.MeetID
		task.MeetGoal = lapuser.MeetEvent.Summary
		task.MeetDate = lapuser.MeetEvent.Date
		task.ProjectWAGroupID = lapuser.Project.WAGroupID
		_, err = atdb.InsertOneDoc(config.Mongoconn, "tasklist", task)
		if err != nil {
			resp.Info = "Kakak sudah melaporkan tasklist sebelumnya"
			resp.Response = "Error : tidak bisa insert ke database " + err.Error()
			return c.Status(fiber.StatusForbidden).JSON(resp)
		}
	}
	res, err := report.TambahPoinTasklistbyPhoneNumber(config.Mongoconn, docusr.PhoneNumber, lapuser.Project, float64(len(tasklists)), "tasklist")
	if err != nil {
		resp.Info = "Tambah Poin Tasklist gagal"
		resp.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(resp)
	}
	resp.Response = strconv.Itoa(int(res.ModifiedCount))
	resp.Info = docusr.Name
	return c.Status(fiber.StatusOK).JSON(resp)
}

func PostPresensi(c *fiber.Ctx) error {
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
	var presensi report.PresensiDomyikado
	err = c.BodyParser(&presensi)
	if err != nil {
		resp.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	docusr, err := atdb.GetOneLatestDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": presensi.PhoneNumber})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		return c.Status(fiber.StatusForbidden).JSON(resp)
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "presensi", presensi)
	if err != nil {
		resp.Info = "Kakak sudah melaporkan presensi sebelumnya"
		resp.Response = "Error : tidak bisa insert ke database " + err.Error()
		return c.Status(fiber.StatusForbidden).JSON(resp)
	}
	res, err := report.TambahPoinPresensibyPhoneNumber(config.Mongoconn, presensi.PhoneNumber, presensi.Lokasi, presensi.Skor, config.WAAPIToken, config.WAAPIMessage, "presensi")
	if err != nil {
		resp.Info = "Tambah Poin Presensi gagal"
		resp.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(resp)
	}
	resp.Response = strconv.Itoa(int(res.ModifiedCount))
	resp.Info = docusr.Name
	return c.Status(fiber.StatusOK).JSON(resp)
}

func PostTestimoni(c *fiber.Ctx) error {
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
	//pindah ke struck user
	var usersub model.Peserta
	usersub.Fullname = userdt.Data.Fullname
	usersub.Desa = userdt.Data.Village
	usersub.Kec = userdt.Data.District
	usersub.Kab = userdt.Data.Regency
	usersub.PhoneNumber = payload.Id
	usersub.Provinsi = userdt.Data.Province

	var rating report.Rating
	var respn model.Response
	err = c.BodyParser(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	usersub.Rating = rating.Rating
	usersub.Komentar = rating.Komentar
	res, err := atdb.InsertOneDoc(config.Mongoconn, "unsubs", usersub)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	respn.Response = res.Hex()
	respn.Info = usersub.Fullname
	return c.Status(fiber.StatusOK).JSON(respn)
}

func GetRandomTesti4(c *fiber.Ctx) error {
	var respn model.Response
	lstpeserta, err := atdb.GetRandomDoc[model.Peserta](config.Mongoconn, "unsubs", 4)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	var listtesti []model.Testi
	for _, testi := range lstpeserta {
		tst := model.Testi{
			Isi:    testi.Komentar,
			Nama:   testi.Fullname,
			Daerah: "Desa " + testi.Desa + " Kec. " + testi.Kec + " " + testi.Kab + " Prov. " + testi.Provinsi,
		}
		listtesti = append(listtesti, tst)
	}
	testidepan := model.Depan{
		List: listtesti,
	}
	return c.Status(fiber.StatusOK).JSON(testidepan)
}

func PostUnsubscribe(c *fiber.Ctx) error {
	var rating report.Rating
	var respn model.Response
	err := c.BodyParser(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	objectId, err := primitive.ObjectIDFromHex(rating.ID)
	if err != nil {
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	hasil, err := atdb.GetOneLatestDoc[model.Peserta](config.Mongoconn, "sent", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	hasil.Rating = rating.Rating
	hasil.Komentar = rating.Komentar
	res, err := atdb.InsertOneDoc(config.Mongoconn, "unsubs", hasil)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	respn.Response = res.Hex()
	respn.Info = hasil.Fullname
	return c.Status(fiber.StatusOK).JSON(respn)
}

func GetFAQ(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	hasil, err := atdb.GetOneLatestDoc[kimseok.Datasets](config.Mongoconn, "faq", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data profile user sent tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	return c.Status(fiber.StatusOK).JSON(hasil)
}

func GetSentItem(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	hasil, err := atdb.GetOneLatestDoc[model.Peserta](config.Mongoconn, "sent", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data profile user sent tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	hasil.PhoneNumber = ""
	return c.Status(fiber.StatusOK).JSON(hasil)
}

func GetClosedTicket(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	hasil, err := atdb.GetOneLatestDoc[tiket.Bantuan](config.Mongoconn, "tiket", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data tiket tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	return c.Status(fiber.StatusOK).JSON(hasil)
}

func GetBotList(c *fiber.Ctx) error {
	Profiles, err := atdb.GetAllDoc[[]itmodel.Profile](config.Mongoconn, "profile", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data tiket tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	var phonelist []string

	for _, profile := range Profiles {
		phonelist = append(phonelist, phone.MaskPhoneNumber(profile.Phonenumber))
	}
	hasil := model.PhoneList{PhoneList: phonelist}

	return c.Status(fiber.StatusOK).JSON(hasil)
}

func PostMasukanTiket(c *fiber.Ctx) error {
	var rating report.Rating
	var respn model.Response
	err := c.BodyParser(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	objectId, err := primitive.ObjectIDFromHex(rating.ID)
	if err != nil {
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	updatefields := primitive.M{
		"ratelayanan": rating.Rating,
		"masukan":     rating.Komentar,
	}
	res, err := atdb.UpdateOneDoc(config.Mongoconn, "tiket", primitive.M{"_id": objectId}, updatefields)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//mendapatkan document untuk informasi ke admin
	hasil, err := atdb.GetOneLatestDoc[tiket.Bantuan](config.Mongoconn, "tiket", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}

	nama := hasil.UserName
	if nama == "" {
		nama = phone.MaskPhoneNumber(hasil.UserPhone)
	}

	respn.Response = strconv.Itoa(int(res.ModifiedCount))
	respn.Info = nama
	//info ke admin
	message := helpdesk.GetPrefillMessage("adminnotiffeedback", config.Mongoconn)
	message = fmt.Sprintf(message, rating.Rating, nama, rating.Komentar)
	dt := &whatsauth.TextMessage{
		To:       hasil.AdminPhone,
		IsGroup:  false,
		Messages: message,
	}
	go atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)

	return c.Status(fiber.StatusOK).JSON(respn)
}

func PostMeeting(c *fiber.Ctx) error {
	var respn model.Response
	//otorisasi dan validasi inputan
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var event gcallapi.SimpleEvent
	err = c.BodyParser(&event)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	//check validasi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan: " + payload.Id
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": event.ProjectID})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan: " + event.ProjectID.Hex()
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//lojik inputan post
	var lap model.Laporan
	lap.User = docuser
	lap.Project = prjuser
	lap.Phone = prjuser.Owner.PhoneNumber
	lap.Nama = prjuser.Owner.Name
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber
	lap.Solusi = event.Description
	//mengambil daftar email dari project member
	var attendees []string
	for _, member := range prjuser.Menu {
		attendees = append(attendees, member.Name)
	}
	event.Attendees = attendees

	gevt, err := gcallapi.HandlerCalendar(config.Mongoconn, event)
	if err != nil {
		respn.Status = "Gagal Membuat Google Calendar"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "meetinglog", gevt)
	if err != nil {
		respn.Status = "Gagal Insert Database meetinglog"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	event.ID, err = atdb.InsertOneDoc(config.Mongoconn, "meeting", event)
	if err != nil {
		respn.Status = "Gagal Insert Database meeting"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	lap.MeetID = event.ID
	lap.MeetEvent = event
	lap.Kode = gevt.HtmlLink
	lap.ID, err = atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "meeting")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinLaporanbyPhoneNumber gagal"
		resp.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(resp)
	}

	message := "*" + strings.TrimSpace(event.Summary) + "*\n" + lap.Kode + "\nLokasi:\n" + event.Location + "\nAgenda:\n" + event.Description + "\nTanggal: " + event.Date + "\nJam: " + event.TimeStart + " - " + event.TimeEnd + "\nNotulen : " + docuser.Name + "\nURL Input Risalah Pertemuan:\n" + "https://www.do.my.id/resume/#" + lap.ID.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Project.WAGroupID,
		IsGroup:  true,
		Messages: message,
	}
	_, respAPI, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		respn.Info = "Tidak berhak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}
	_ = respAPI // Ignore respAPI since we don't return it
	return c.Status(fiber.StatusOK).JSON(lap)
}

func PostLaporan(c *fiber.Ctx) error {
	//otorisasi dan validasi inputan
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var lap model.Laporan
	err = c.BodyParser(&lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	if lap.Solusi == "" {
		var respn model.Response
		respn.Status = "Error : Telepon atau nama atau solusi tidak diisi"
		respn.Response = "Isi lebih lengkap dahulu"
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	//check validasi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan: " + payload.Id
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//ambil data project
	prjobjectId, err := primitive.ObjectIDFromHex(lap.Kode)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = lap.Kode
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prjobjectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan: " + lap.Kode
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//lojik inputan post
	lap.User = docuser
	lap.Project = prjuser
	lap.Phone = prjuser.Owner.PhoneNumber
	lap.Nama = prjuser.Owner.Name
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber

	idlap, err := atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "laporan")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinPushRepobyGithubUsername gagal"
		resp.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(resp)
	}
	message := "*Permintaan Feedback Pekerjaan*\n" + "Petugas : " + docuser.Name + "\nDeskripsi:" + lap.Solusi + "\n Beri Nilai: " + "https://www.do.my.id/rate/#" + idlap.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Phone,
		IsGroup:  false,
		Messages: message,
	}
	_, respAPI, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		var respn model.Response
		respn.Info = "Tidak berhak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}
	_ = respAPI
	return c.Status(fiber.StatusOK).JSON(lap)

}

func PostFeedback(c *fiber.Ctx) error {
	//otorisasi dan validasi inputan
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeaderFiber(c))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeaderFiber(c)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusForbidden).JSON(respn)
	}
	var lap model.Laporan
	err = c.BodyParser(&lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	if lap.Phone == "" || lap.Nama == "" || lap.Solusi == "" {
		var respn model.Response
		respn.Status = "Error : Telepon atau nama atau solusi tidak diisi"
		respn.Response = "Isi lebih lengkap dahulu"
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	//validasi eksistensi user di db
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//ambil data project
	prjobjectId, err := primitive.ObjectIDFromHex(lap.Kode)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = lap.Kode
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prjobjectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan: " + lap.Kode
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotImplemented).JSON(respn)
	}
	//lojik inputan post
	lap.Project = prjuser
	lap.User = docuser
	lap.Phone = ValidasiNoHP(lap.Phone)
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber

	idlap, err := atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		return c.Status(fiber.StatusNotModified).JSON(respn)
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "feedback")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinLaporanbyPhoneNumber gagal"
		resp.Response = err.Error()
		return c.Status(fiber.StatusExpectationFailed).JSON(resp)
	}
	message := "*Permintaan Feedback*\n" + "Petugas : " + docuser.Name + "\nDeskripsi:" + lap.Solusi + "\n Beri Nilai: " + "https://www.do.my.id/rate/#" + idlap.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Phone,
		IsGroup:  false,
		Messages: message,
	}
	_, respAPI, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		var respn model.Response
		respn.Info = "Tidak berhak"
		respn.Response = err.Error()
		return c.Status(fiber.StatusUnauthorized).JSON(respn)
	}
	_ = respAPI
	return c.Status(fiber.StatusOK).JSON(lap)

}

func ValidasiNoHP(nomor string) string {
	nomor = strings.ReplaceAll(nomor, " ", "")
	nomor = strings.ReplaceAll(nomor, "+", "")
	nomor = strings.ReplaceAll(nomor, "-", "")
	return nomor
}
