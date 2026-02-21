package controller

import (
	"github.com/gocroot/config"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/model"
	"github.com/gofiber/fiber/v2"
)

func GetCountDocUser(c *fiber.Ctx) error {
	var resp model.Response
	rkp, err := lms.GetRekapPendaftaranUsers(config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		return c.Status(fiber.StatusConflict).JSON(resp)
	}
	return c.Status(fiber.StatusOK).JSON(rkp)
}

func RefreshLMSCookie(c *fiber.Ctx) error {
	var resp model.Response
	err := lms.RefreshCookie(config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}
	resp.Info = "ok"
	return c.Status(fiber.StatusOK).JSON(resp)
}
