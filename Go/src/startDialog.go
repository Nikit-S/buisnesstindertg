package src

import (
	"fmt"
	"testbot/cmp"
)

func StartChatWithUser(b *cmp.Bot, ch *cmp.Chat) {
	update := <-ch.Updates

	if update.Message.Command() == "start" {

		fmt.Println("START")
		//msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Отправь свой телефон чтобы мы могли с тобой связаться"))
		//b.LogMsg(msg)
		if b.CheckForRoot(update) {
			go Root(b, ch)
			return
		}
	}
	go SimpleUser(b, ch)
}
