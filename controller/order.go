package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/jualin"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
)

// Fungsi untuk menangani request order
func HandleOrder(c *fiber.Ctx) error {
	namalapak := c.Params("namalapak")
	var orderRequest jualin.PaymentRequest

	// Decode JSON request ke struct
	if err := c.BodyParser(&orderRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Bad request")
	}

	_, err := atdb.InsertOneDoc(config.Mongoconn, "order", orderRequest)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Insert Database Gagal")
	}

	//kirim pesan ke tenant
	message := "*Pesanan Masuk " + namalapak + "*\n" + orderRequest.User.Name + "\n" + orderRequest.User.Whatsapp + "\n" + orderRequest.User.Address + "\n" + createOrderMessage(orderRequest.Orders) + "\nTotal: " + strconv.Itoa(orderRequest.Total) + "\nPembayaran: " + orderRequest.PaymentMethod
	newmsg := model.SendText{
		To:       "628111269691",
		IsGroup:  false,
		Messages: message,
	}
	_, _, err = atapi.PostStructWithToken[model.Response]("token", config.WAAPIToken, newmsg, config.WAAPIMessage)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Gagal Mengirim pesan")
	}
	// Cetak data order ke terminal (bisa diganti dengan logic lain, misal menyimpan ke database)
	fmt.Printf("Received Order: %+v\n", orderRequest)

	// Kirim response kembali ke client
	response := map[string]string{"status": "success", "message": "Order received"}
	return c.JSON(response)
}

// Fungsi untuk membuat pesan dari orders
func createOrderMessage(orders []jualin.Order) string {
	var orderStrings []string

	for _, order := range orders {
		orderString := fmt.Sprintf("%s x%d - Rp %d", order.Name, order.Quantity, order.Price)
		orderStrings = append(orderStrings, orderString)
	}

	// Gabungkan semua orders menjadi satu string dengan new line sebagai separator
	return strings.Join(orderStrings, "\n")
}
