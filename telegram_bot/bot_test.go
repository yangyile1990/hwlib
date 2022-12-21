package telegrambot

import (
	"fmt"
)

func BotTest() {
	bot, err := NewBot("5904746042:AAGjBMN_ahQ0uavSCakrEFUN7RV2Q8oDY4I")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = bot.SendMsg(832328501, "test")
	if err != nil {
		fmt.Println(err)
	}
}
