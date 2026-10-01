package auth

import (
	"context"
	"honi/internal/agent"
	"log/slog"
	"os"
	"sync"

	"github.com/SpellingDragon/wechat-robot-go/wechat"
	"github.com/mdp/qrterminal/v4"
)

var GetWechatBot = sync.OnceValue(func() *Wechat {
	return &Wechat{
		bot: wechat.NewBot(),
		ctx: context.Background(),
	}
})

type Wechat struct {
	bot *wechat.Bot
	ctx context.Context
}

func (w *Wechat) HasValidateCredentials() bool {
	w.bot.ClearAllContextTokens()
	return w.bot.HasValidateCredentials()
}

func (w *Wechat) Run() error {
	return w.bot.Run(w.ctx)
}

func (w *Wechat) Login() error {
	ctx := context.Background()
	err := w.bot.Login(ctx, func(qrCodeUrl string) {
		config := qrterminal.Config{
			Writer:     os.Stdout,
			HalfBlocks: true,
			Level:      qrterminal.L,
			QuietZone:  2,
		}
		qrterminal.GenerateWithConfig(qrCodeUrl, config)
	})
	if err != nil {
		slog.Error("微信登录失败", "err", err)
		return err
	}

	// 注册消息处理器
	w.bot.OnMessage(func(ctx context.Context, msg *wechat.Message) error {
		text := msg.Text()
		if text == "" {
			return nil
		}
		return w.bot.Reply(ctx, msg, agent.GetAgent().Ask(text))
	})

	return nil
}
