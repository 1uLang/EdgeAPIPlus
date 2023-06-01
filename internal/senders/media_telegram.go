//go:build plus

package senders

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/iwind/TeaGo/types"
)

// TelegramMedia Telegram媒介
type TelegramMedia struct {
	Token string `yaml:"token" json:"token"`
}

// NewTelegramMedia 获取新对象
func NewTelegramMedia() *TelegramMedia {
	return &TelegramMedia{}
}

// Send 发送消息
func (this *TelegramMedia) Send(user string, subject string, body string) (respBytes []byte, err error) {
	bot, err := tgbotapi.NewBotAPI(this.Token)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(types.Int64(user), subject+"\n"+body)
	_, err = bot.Send(msg)
	return nil, err
}

// RequireUser 是否需要用户标识
func (this *TelegramMedia) RequireUser() bool {
	return true
}
