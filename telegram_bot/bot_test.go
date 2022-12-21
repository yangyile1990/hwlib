package telegrambot_test

import (
	"fmt"
	"testing"

	"github.com/suiguo/telegrambot"
)

func TestBot(t *testing.T) {
	bot, err := telegrambot.NewBot("5904746042:AAGjBMN_ahQ0uavSCakrEFUN7RV2Q8oDY4I")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = telegrambot.SendMsg(832328501, "test")
	if err != nil {
		fmt.Println(err)
	}
}
