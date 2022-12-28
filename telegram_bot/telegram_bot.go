package telegrambot

import (
	"fmt"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var apiMap sync.Map

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

func NewBot(api_key string, newInstance ...bool) (Bot, error) {
	if len(newInstance) > 0 && newInstance[0] { //填了数据返回全新的api实例
		tmp := &bot{}
		apiMap.Store(api_key, tmp)
		return tmp, tmp.Init(api_key)
	}
	single, ok := apiMap.Load(api_key)
	if ok {
		return single.(*bot), nil
	}
	tmp := &bot{}
	err := tmp.Init(api_key)
	if err == nil {
		apiMap.Store(api_key, tmp)
		return tmp, nil
	}
	return nil, err
}
