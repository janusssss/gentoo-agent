package api

import (
	"honi/internal/auth"

	"github.com/gofiber/fiber/v3"
)

func LoginWechat(c fiber.Ctx) error {
	bot := auth.GetWechatBot()
	if err := bot.Login(); err != nil {
		return err
	}
	if err := bot.Run(); err != nil {
		return err
	}
	return c.SendString("Hello, World 👋!")
}
