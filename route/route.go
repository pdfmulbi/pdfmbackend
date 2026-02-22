package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"

	_ "github.com/gocroot/docs"
)

func RegisterRoutes(app *fiber.App) {
	app.Use(config.CorsMiddleware())

	app.Use(func(c *fiber.Ctx) error {
		config.SetEnv()
		return c.Next()
	})

	feedbackHandler := &controller.FeedbackHandler{}
	historyHandler := &controller.HistoryHandler{}
	notifHandler := &controller.NotificationHandler{}
	userHandler := &controller.UserHandler{}

	app.Get("/swagger/*", func(c *fiber.Ctx) error {
		// httpSwagger.WrapHandler is for net/http. Needs fiber-swagger.
		// We'll replace this later with fiber-swagger
		return c.SendString("Swagger UI")
	})

	app.Get("/", controller.GetHome)

	//chat bot inbox
	app.Post("/webhook/nomor/:nomorwa", controller.PostInboxNomor)

	//masking list nmor official
	app.Get("/data/phone/all", controller.GetBotList)

	//akses data helpdesk layanan user
	app.Get("/data/user/helpdesk/all", controller.GetHelpdeskAll)
	app.Get("/data/user/helpdesk/masuk", controller.GetLatestHelpdeskMasuk)
	app.Get("/data/user/helpdesk/selesai", controller.GetLatestHelpdeskSelesai)

	//pamong desa data from api
	app.Get("/data/lms/user", controller.GetDataUserFromApi)

	//simpan testimoni dari pamong desa lms api
	app.Post("/data/lms/testi", controller.PostTestimoni)

	//get random 4 testi
	app.Get("/data/lms/random/testi", controller.GetRandomTesti4)

	//mendapatkan data sent item
	app.Get("/data/peserta/sent/:id", controller.GetSentItem)

	//simpan feedback unsubs user
	app.Post("/data/peserta/unsubscribe", controller.PostUnsubscribe)

	//generate token linked device
	app.Put("/data/user", controller.PutTokenDataUser)

	//Menambhahkan data nomor sender untuk broadcast
	app.Put("/data/sender", controller.PutNomorBlast)

	//mendapatkan data list nomor sender untuk broadcast
	app.Get("/data/sender", controller.GetDataSenders)

	//mendapatkan data list nomor sender yang kena blokir dari broadcast
	app.Get("/data/blokir", controller.GetDataSendersTerblokir)

	//mendapatkan data rekap pengiriman wa blast
	app.Get("/data/rekap", controller.GetRekapBlast)

	//mendapatkan data faq
	app.Get("/data/faq/:id", controller.GetFAQ)

	//legacy
	app.Put("/data/user/task/doing", controller.PutTaskUser)
	app.Get("/data/user/task/done", controller.GetTaskDone)
	app.Post("/data/user/task/done", controller.PostTaskUser)
	app.Get("/data/pushrepo/kemarin", controller.GetYesterdayDistincWAGroup)

	//Helpdesk
	//mendapatkan data tiket
	app.Get("/data/tiket/closed/:id", controller.GetClosedTicket)

	//simpan feedback tiket user
	app.Post("/data/tiket/rate", controller.PostMasukanTiket)
	// order
	app.Post("/data/order/:namalapak", controller.HandleOrder)

	//user data
	app.Get("/data/user", controller.GetDataUser)

	//user pendaftaran
	app.Post("/auth/register/users", controller.RegisterGmailAuth)
	app.Post("/data/user", controller.PostDataUser)
	app.Post("/upload/profpic", controller.UploadProfilePictureHandler)
	app.Post("/data/user/bio", controller.PostDataBioUser)

	//data proyek
	app.Get("/data/proyek", controller.GetDataProject)
	app.Get("/data/proyek/approved", controller.GetEditorApprovedProject)
	app.Post("/data/proyek", controller.PostDataProject)
	app.Put("/data/metadatabuku", controller.PutMetaDataProject)
	app.Put("/data/proyek/publishbuku", controller.PutPublishProject)
	app.Put("/data/proyek", controller.PutDataProject)
	app.Delete("/data/proyek", controller.DeleteDataProject)
	app.Get("/data/proyek/anggota", controller.GetDataMemberProject)
	app.Get("/data/proyek/editor", controller.GetDataEditorProject)
	app.Delete("/data/proyek/anggota", controller.DeleteDataMemberProject)
	app.Post("/data/proyek/anggota", controller.PostDataMemberProject)
	app.Post("/data/proyek/editor", controller.PostDataEditorProject)
	app.Put("/data/proyek/editor", controller.PUtApprovedEditorProject)

	//upload cover,draft,pdf,sampul buku project
	app.Post("/upload/coverbuku/:projectid", controller.UploadCoverBukuWithParamFileHandler)
	app.Post("/upload/draftbuku/:projectid", controller.UploadDraftBukuWithParamFileHandler)
	app.Post("/upload/draftpdfbuku/:projectid", controller.UploadDraftBukuPDFWithParamFileHandler)
	app.Post("/upload/sampulpdfbuku/:projectid", controller.UploadSampulBukuPDFWithParamFileHandler)
	app.Post("/upload/spk/:projectid", controller.UploadSPKPDFWithParamFileHandler)
	app.Post("/upload/spi/:projectid", controller.UploadSPIPDFWithParamFileHandler)
	app.Get("/download/draft/:path", controller.AksesFileRepoDraft)
	app.Post("/data/proyek/katalog", controller.PostKatalogBuku)
	app.Get("/download/dokped/spk/:namaproject", controller.GetFileDraftSPK)
	app.Get("/download/dokped/spkt/:namaproject", controller.GetFileDraftSPKT)
	app.Get("/download/dokped/spi/:path", controller.GetFileDraftSPI)

	app.Post("/data/proyek/menu", controller.PostDataMenuProject)
	app.Post("/approvebimbingan", controller.ApproveBimbinganbyPoin)
	app.Delete("/data/proyek/menu", controller.DeleteDataMenuProject)
	app.Post("/notif/ux/postlaporan", controller.PostLaporan)
	app.Post("/notif/ux/postfeedback", controller.PostFeedback)

	app.Post("/notif/ux/postmeeting", controller.PostMeeting)
	app.Post("/notif/ux/postpresensi/:id", controller.PostPresensi)
	app.Post("/notif/ux/posttasklists/:id", controller.PostTaskList)
	// Webhook nomor handler - duplicated due to two definitions in old router
	// app.Post("/webhook/nomor/:nomorwa", controller.PostInboxNomor)

	// LMS
	app.Get("/lms/refresh/cookie", controller.RefreshLMSCookie)
	app.Get("/lms/count/user", controller.GetCountDocUser)

	//PDFM
	//Profile Photo
	app.Post("/pdfm/profile/photo", userHandler.UploadProfilePhotoHandler)
	app.Get("/pdfm/profile/photo", userHandler.GetProfilePhotoHandler)

	//Register
	app.Post("/pdfm/register", userHandler.RegisterHandler)
	//Login
	app.Post("/pdfm/login", userHandler.GetUser)
	//Logout
	app.Post("/pdfm/logout", userHandler.LogoutHandler)

	//PaymentHandler
	app.Post("/pdfm/payment", userHandler.ConfirmPaymentHandler)

	//Get InvoiceHandler
	app.Get("/pdfm/invoices", userHandler.GetInvoicesHandler)

	//CRUD
	app.Get("/pdfm/get/users", userHandler.GetUsers)
	app.Post("/pdfm/create/users", userHandler.CreateUser)
	app.Get("/pdfm/getone/users", userHandler.GetOneUser)
	app.Get("/pdfm/getoneadmin/users", userHandler.GetOneUserAdmin)
	app.Put("/pdfm/update/users", userHandler.UpdateUser)
	app.Delete("/pdfm/delete/users", userHandler.DeleteUser)

	//Notifications
	app.Get("/pdfm/notifications", notifHandler.GetNotifications)
	app.Post("/pdfm/notifications", notifHandler.AddNotification)
	app.Put("/pdfm/notifications/read", notifHandler.MarkAllAsRead)
	app.Delete("/pdfm/notifications", notifHandler.ClearNotifications)

	// 1. Merge Logs
	app.Post("/pdfm/log/merge", historyHandler.CreateMergeHistory)
	app.Get("/pdfm/log/merge", historyHandler.GetMergeHistory)

	// 2. Compress Logs
	app.Post("/pdfm/log/compress", historyHandler.CreateCompressHistory)
	app.Get("/pdfm/log/compress", historyHandler.GetCompressHistory)

	// 3. Convert Logs
	app.Post("/pdfm/log/convert", historyHandler.CreateConvertHistory)
	app.Get("/pdfm/log/convert", historyHandler.GetConvertHistory)

	// 4. Summary Logs
	app.Post("/pdfm/log/summary", historyHandler.CreateSummaryHistory)
	app.Get("/pdfm/log/summary", historyHandler.GetSummaryHistory)
	app.Post("/pdfm/ai/summary", historyHandler.SummarizePDF)

	// 5. All History (Combined)
	app.Get("/pdfm/history/all", historyHandler.GetAllHistory)
	app.Delete("/pdfm/history/delete", historyHandler.DeleteHistory)

	// Feedback (Kotak Saran / Contact Us)
	app.Post("/pdfm/feedback", feedbackHandler.InsertFeedback)
	app.Get("/pdfm/feedback", feedbackHandler.GetAllFeedback)

	// Google Auth
	app.Post("/auth/users", controller.Auth)
	app.Post("/auth/login", controller.GeneratePasswordHandler)
	app.Post("/auth/verify", controller.VerifyPasswordHandler)
	app.Post("/auth/resend", controller.ResendPasswordHandler)

	// 404 setup equivalent
	app.Use(controller.NotFound)
}
