package telegrambot

import (
	"fmt"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot interface {
	SendMsg(chat_id int64, msg string) error
	SendMsgByUser(user string, msg string) error
	Init(api_key string) error
}

type bot struct {
	sync.Mutex
	Bot
	c *tgbotapi.BotAPI
}

func (b *bot) SendMsg(chat_id int64, msg string) error {
	if b.c == nil {
		return fmt.Errorf("robot is nil")
	}
	msg_instance := tgbotapi.NewMessage(chat_id, msg)
	_, err := b.c.Send(msg_instance)
	return err
}

func (b *bot) SendMsgByUser(user string, msg string) error {
	if b.c == nil {
		return fmt.Errorf("robot is nil")
	}
	msg_instance := tgbotapi.NewMessageToChannel(user, msg)
	_, err := b.c.Send(msg_instance)
	return err
}

func (b *bot) Init(api_key string) error {
	c, err := tgbotapi.NewBotAPIWithClient(
		api_key,
		tgbotapi.APIEndpoint, &http.Client{},
	)
	if err != nil {
		return err
	}
	b.Lock()
	defer b.Unlock()
	b.c = c
	return nil
}

func NewBot(api_key string) (Bot, error) {
	tmp := &bot{}
	return tmp, tmp.Init(api_key)
}
