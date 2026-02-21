package controller

import (
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
)

func GetHome(c *fiber.Ctx) error {
	var resp model.Response
	resp.Response = at.GetIPaddress()
	return c.Status(fiber.StatusOK).JSON(resp)
}

func PostInboxNomor(c *fiber.Ctx) error {
	var resp itmodel.Response
	var msg itmodel.IteungMessage
	httpstatus := fiber.StatusUnauthorized
	resp.Response = "Wrong Secret"
	waphonenumber := c.Params("phonenumber") // Adjust param name if different in route.go
	if waphonenumber == "" {                 // fallback to getparam equivalent if param not named
		waphonenumber = c.Params("*")
	}
	prof, err := whatsauth.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		httpstatus = fiber.StatusServiceUnavailable
	}
	if at.GetSecretFromHeaderFiber(c) == prof.Secret {
		err := c.BodyParser(&msg)
		if err != nil {
			resp.Response = err.Error()
		} else {
			resp, err = whatsauth.WebHook(prof, msg, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
			}
		}
	}
	return c.Status(httpstatus).JSON(resp)
}

// jalan setiap jam 3 pagi
// jalan setiap jam 3 pagi
func GetNewToken(c *fiber.Ctx) error {
	var resp model.Response
	httpstatus := fiber.StatusServiceUnavailable

	var wg sync.WaitGroup
	wg.Add(3)

	var mu sync.Mutex
	var lastErr error

	go func() {
		defer wg.Done()
		profs, err := atdb.GetAllDoc[[]model.Profile](config.Mongoconn, "profile", bson.M{})
		if err != nil {
			mu.Lock()
			lastErr = err
			resp.Response = err.Error()
			mu.Unlock()
			return
		}
		for _, prof := range profs {
			dt := &itmodel.WebHook{
				URL:    prof.URL,
				Secret: prof.Secret,
			}
			res, err := whatsauth.RefreshToken(dt, prof.Phonenumber, config.WAAPIGetToken, config.Mongoconn)
			if err != nil {
				mu.Lock()
				lastErr = err
				resp.Response = err.Error()
				httpstatus = fiber.StatusInternalServerError
				mu.Unlock()
				continue
			}
			mu.Lock()
			resp.Response = at.Jsonstr(res.ModifiedCount)
			httpstatus = fiber.StatusOK
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if err := report.RekapMeetingKemarin(config.Mongoconn); err != nil {
			mu.Lock()
			lastErr = err
			resp.Response = err.Error()
			httpstatus = fiber.StatusInternalServerError
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if err := report.RekapPagiHari(config.Mongoconn); err != nil {
			mu.Lock()
			lastErr = err
			resp.Response = err.Error()
			httpstatus = fiber.StatusInternalServerError
			mu.Unlock()
		}
	}()

	wg.Wait()

	if lastErr != nil {
		return c.Status(httpstatus).JSON(resp)
	} else {
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func NotFound(c *fiber.Ctx) error {
	var resp model.Response
	resp.Response = "Not Found"
	return c.Status(fiber.StatusNotFound).JSON(resp)
}
