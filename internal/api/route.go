package api

import (
	"honi/common"
	"honi/internal/auth"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func Run() {
	app := fiber.New()
	app.Get(common.RunBot, LoginWechat)
	app.Post("/v1/chat/completions", wechatBot)

	app.Hooks().OnListen(func(data fiber.ListenData) error {
		bot := auth.GetWechatBot()
		if !bot.HasValidateCredentials() {
			slog.Warn("登录凭证无效或缺失")
			return nil
		}
		if err := bot.Login(); err != nil {
			return err
		}
		go bot.Run()
		return nil
	})

	slog.Error("启动监听失败", "err", app.Listen(":12358"))
}
