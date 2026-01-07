package route

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper/at"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return // If it's a preflight request, return early.
	}
	config.SetEnv()

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "GET" && path == "/":
		controller.GetHome(w, r)

	//chat bot inbox
	case method == "POST" && at.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)

	//masking list nmor official
	case method == "GET" && path == "/data/phone/all":
		controller.GetBotList(w, r)

	//akses data helpdesk layanan user
	case method == "GET" && path == "/data/user/helpdesk/all":
		controller.GetHelpdeskAll(w, r)
	case method == "GET" && path == "/data/user/helpdesk/masuk":
		controller.GetLatestHelpdeskMasuk(w, r)
	case method == "GET" && path == "/data/user/helpdesk/selesai":
		controller.GetLatestHelpdeskSelesai(w, r)

	//pamong desa data from api
	case method == "GET" && path == "/data/lms/user":
		controller.GetDataUserFromApi(w, r)

	//simpan testimoni dari pamong desa lms api
	case method == "POST" && path == "/data/lms/testi":
		controller.PostTestimoni(w, r)

		//get random 4 testi
	case method == "GET" && path == "/data/lms/random/testi":
		controller.GetRandomTesti4(w, r)

	//mendapatkan data sent item
	case method == "GET" && at.URLParam(path, "/data/peserta/sent/:id"):
		controller.GetSentItem(w, r)

	//simpan feedback unsubs user
	case method == "POST" && path == "/data/peserta/unsubscribe":
		controller.PostUnsubscribe(w, r)

	//generate token linked device
	case method == "PUT" && path == "/data/user":
		controller.PutTokenDataUser(w, r)

	//Menambhahkan data nomor sender untuk broadcast
	case method == "PUT" && path == "/data/sender":
		controller.PutNomorBlast(w, r)

	//mendapatkan data list nomor sender untuk broadcast
	case method == "GET" && path == "/data/sender":
		controller.GetDataSenders(w, r)

	//mendapatkan data list nomor sender yang kena blokir dari broadcast
	case method == "GET" && path == "/data/blokir":
		controller.GetDataSendersTerblokir(w, r)

	//mendapatkan data rekap pengiriman wa blast
	case method == "GET" && path == "/data/rekap":
		controller.GetRekapBlast(w, r)

	//mendapatkan data faq
	case method == "GET" && at.URLParam(path, "/data/faq/:id"):
		controller.GetFAQ(w, r)

	//legacy
	case method == "PUT" && path == "/data/user/task/doing":
		controller.PutTaskUser(w, r)
	case method == "GET" && path == "/data/user/task/done":
		controller.GetTaskDone(w, r)
	case method == "POST" && path == "/data/user/task/done":
		controller.PostTaskUser(w, r)
	case method == "GET" && path == "/data/pushrepo/kemarin":
		controller.GetYesterdayDistincWAGroup(w, r)

	//Helpdesk
	//mendapatkan data tiket
	case method == "GET" && at.URLParam(path, "/data/tiket/closed/:id"):
		controller.GetClosedTicket(w, r)

	//simpan feedback tiket user
	case method == "POST" && path == "/data/tiket/rate":
		controller.PostMasukanTiket(w, r)
		// order
	case method == "POST" && at.URLParam(path, "/data/order/:namalapak"):
		controller.HandleOrder(w, r)

	//user data
	case method == "GET" && path == "/data/user":
		controller.GetDataUser(w, r)

	//user pendaftaran
	case method == "POST" && path == "/auth/register/users": //mendapatkan email gmail
		controller.RegisterGmailAuth(w, r)
	case method == "POST" && path == "/data/user":
		controller.PostDataUser(w, r)
	case method == "POST" && path == "/upload/profpic": //upload gambar profile
		controller.UploadProfilePictureHandler(w, r)
	case method == "POST" && path == "/data/user/bio":
		controller.PostDataBioUser(w, r)
		/* 	case method == "POST" && at.URLParam(path, "/data/user/wa/:nomorwa"):
		controller.PostDataUserFromWA(w, r) */

	//data proyek
	case method == "GET" && path == "/data/proyek":
		controller.GetDataProject(w, r)
	case method == "GET" && path == "/data/proyek/approved": //akses untuk manager
		controller.GetEditorApprovedProject(w, r)
	case method == "POST" && path == "/data/proyek":
		controller.PostDataProject(w, r)
	case method == "PUT" && path == "/data/metadatabuku":
		controller.PutMetaDataProject(w, r)
	case method == "PUT" && path == "/data/proyek/publishbuku": //publish buku isbn by manager
		controller.PutPublishProject(w, r)
	case method == "PUT" && path == "/data/proyek":
		controller.PutDataProject(w, r)
	case method == "DELETE" && path == "/data/proyek":
		controller.DeleteDataProject(w, r)
	case method == "GET" && path == "/data/proyek/anggota":
		controller.GetDataMemberProject(w, r)
	case method == "GET" && path == "/data/proyek/editor":
		controller.GetDataEditorProject(w, r)
	case method == "DELETE" && path == "/data/proyek/anggota":
		controller.DeleteDataMemberProject(w, r)
	case method == "POST" && path == "/data/proyek/anggota":
		controller.PostDataMemberProject(w, r)
	case method == "POST" && path == "/data/proyek/editor": //set editor oleh owner
		controller.PostDataEditorProject(w, r)
	case method == "PUT" && path == "/data/proyek/editor": //set approved oleh editor
		controller.PUtApprovedEditorProject(w, r)

	//upload cover,draft,pdf,sampul buku project
	case method == "POST" && at.URLParam(path, "/upload/coverbuku/:projectid"):
		controller.UploadCoverBukuWithParamFileHandler(w, r)
	case method == "POST" && at.URLParam(path, "/upload/draftbuku/:projectid"):
		controller.UploadDraftBukuWithParamFileHandler(w, r)
	case method == "POST" && at.URLParam(path, "/upload/draftpdfbuku/:projectid"):
		controller.UploadDraftBukuPDFWithParamFileHandler(w, r)
	case method == "POST" && at.URLParam(path, "/upload/sampulpdfbuku/:projectid"):
		controller.UploadSampulBukuPDFWithParamFileHandler(w, r)
	case method == "POST" && at.URLParam(path, "/upload/spk/:projectid"):
		controller.UploadSPKPDFWithParamFileHandler(w, r)
	case method == "POST" && at.URLParam(path, "/upload/spi/:projectid"):
		controller.UploadSPIPDFWithParamFileHandler(w, r)
	case method == "GET" && at.URLParam(path, "/download/draft/:path"): //downoad file draft
		controller.AksesFileRepoDraft(w, r)
	case method == "POST" && path == "/data/proyek/katalog": //post blog katalog
		controller.PostKatalogBuku(w, r)
	case method == "GET" && at.URLParam(path, "/download/dokped/spk/:namaproject"): //base64 namaproject
		controller.GetFileDraftSPK(w, r)
	case method == "GET" && at.URLParam(path, "/download/dokped/spkt/:namaproject"): //base64 namaproject
		controller.GetFileDraftSPKT(w, r)
	case method == "GET" && at.URLParam(path, "/download/dokped/spi/:path"): //base64 path sampul
		controller.GetFileDraftSPI(w, r)

	case method == "POST" && path == "/data/proyek/menu":
		controller.PostDataMenuProject(w, r)
	case method == "POST" && path == "/approvebimbingan":
		controller.ApproveBimbinganbyPoin(w, r)
	case method == "DELETE" && path == "/data/proyek/menu":
		controller.DeleteDataMenuProject(w, r)
	case method == "POST" && path == "/notif/ux/postlaporan":
		controller.PostLaporan(w, r)
	case method == "POST" && path == "/notif/ux/postfeedback":
		controller.PostFeedback(w, r)

	case method == "POST" && path == "/notif/ux/postmeeting":
		controller.PostMeeting(w, r)
	case method == "POST" && at.URLParam(path, "/notif/ux/postpresensi/:id"):
		controller.PostPresensi(w, r)
	case method == "POST" && at.URLParam(path, "/notif/ux/posttasklists/:id"):
		controller.PostTaskList(w, r)
	case method == "POST" && at.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)

	// LMS
	case method == "GET" && path == "/lms/refresh/cookie":
		controller.RefreshLMSCookie(w, r)
	case method == "GET" && path == "/lms/count/user":
		controller.GetCountDocUser(w, r)

	//PDFM
	//Profile Photo
	case method == "POST" && path == "/pdfm/profile/photo":
		controller.UploadProfilePhotoHandler(w, r)
	case method == "GET" && path == "/pdfm/profile/photo":
		controller.GetProfilePhotoHandler(w, r)

	//Register
	case method == "POST" && path == "/pdfm/register":
		controller.RegisterHandler(w, r)
	//Login
	case method == "POST" && path == "/pdfm/login":
		controller.GetUser(w, r)
	//Logout
	case method == "POST" && path == "/pdfm/logout":
		controller.LogoutHandler(w, r)

	//PaymentHandler
	case method == "POST" && path == "/pdfm/payment":
		controller.ConfirmPaymentHandler(w, r)

	//Get InvoiceHandler
	case method == "GET" && path == "/pdfm/invoices":
		controller.GetInvoicesHandler(w, r)

	//CRUD
	case method == "GET" && path == "/pdfm/get/users":
		controller.GetUsers(w, r)
	case method == "POST" && path == "/pdfm/create/users":
		controller.CreateUser(w, r)
	case method == "GET" && path == "/pdfm/getone/users":
		controller.GetOneUser(w, r)
	case method == "GET" && path == "/pdfm/getoneadmin/users":
		controller.GetOneUserAdmin(w, r)
	case method == "PUT" && path == "/pdfm/update/users":
		controller.UpdateUser(w, r)
	case method == "DELETE" && path == "/pdfm/delete/users":
		controller.DeleteUser(w, r)

	//Notifications
	case method == "GET" && path == "/pdfm/notifications":
		controller.GetNotifications(w, r)
	case method == "POST" && path == "/pdfm/notifications":
		controller.AddNotification(w, r)
	case method == "PUT" && path == "/pdfm/notifications/read":
		controller.MarkAllAsRead(w, r)
	case method == "DELETE" && path == "/pdfm/notifications":
		controller.ClearNotifications(w, r)

		// 1. Merge Logs
	case method == "POST" && path == "/pdfm/log/merge":
		controller.CreateMergeHistory(w, r)
	case method == "GET" && path == "/pdfm/log/merge":
		controller.GetMergeHistory(w, r)

	// 2. Compress Logs
	case method == "POST" && path == "/pdfm/log/compress":
		controller.CreateCompressHistory(w, r)
	case method == "GET" && path == "/pdfm/log/compress":
		controller.GetCompressHistory(w, r)

	// 3. Convert Logs
	case method == "POST" && path == "/pdfm/log/convert":
		controller.CreateConvertHistory(w, r)
	case method == "GET" && path == "/pdfm/log/convert":
		controller.GetConvertHistory(w, r)

	// 4. Summary Logs
	case method == "POST" && path == "/pdfm/log/summary":
		controller.CreateSummaryHistory(w, r)
	case method == "GET" && path == "/pdfm/log/summary":
		controller.GetSummaryHistory(w, r)

	// 5. All History (Combined)
	case method == "GET" && path == "/pdfm/history/all":
		controller.GetAllHistory(w, r)
	case method == "DELETE" && path == "/pdfm/history/delete":
		controller.DeleteHistory(w, r)
	
	// Feedback (Kotak Saran / Contact Us)
    case method == "POST" && path == "/pdfm/feedback":
        controller.InsertFeedback(w, r)

	// Google Auth
	case method == "POST" && path == "/auth/users":
		controller.Auth(w, r)
	case method == "POST" && path == "/auth/login":
		controller.GeneratePasswordHandler(w, r)
	case method == "POST" && path == "/auth/verify":
		controller.VerifyPasswordHandler(w, r)
	case method == "POST" && path == "/auth/resend":
		controller.ResendPasswordHandler(w, r)

	// Google Auth
	default:
		controller.NotFound(w, r)
	}
}
