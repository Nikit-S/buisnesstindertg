package src

import (
	"fmt"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var partCommands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/unregister", Description: "Отменить участие в бизнес-тиндере"},
)

func ExecPartCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
	var msg tgbotapi.Message
	switch upd.Message.Command() {
	case "start":
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Вы уже начали чат"))
		b.LogMsg(msg)
	case "superuser":
		if b.CheckForRoot(upd) {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Теперь ты супер-пользователь"))
			b.LogMsg(msg)
			go Root(b, ch)
			return nil
		}
		return fmt.Errorf("Not a root")
	case "unregister":
		_, err := b.Db.Exec(`UPDATE public.chats
			SET part=false
			WHERE id=$1`,
			upd.Message.From.ID)
		if err == nil {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Ты больше не участвуешь в Бизнес-Тиндере"))
			b.LogMsg(msg)
			go SimpleUser(b, ch)
			return nil
		}
	}
	return fmt.Errorf("no such command")
}

func Participant(b *cmp.Bot, ch *cmp.Chat) {
	partCommands.Scope.ChatID = ch.Id
	b.BotApi.Request(partCommands)
	msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Если захочешь отказаться от участия напиши /unregister"))
	b.LogMsg(msg)
	for update := range ch.Updates {
		if update.Message.IsCommand() {
			if ExecPartCommand(b, ch, update) == nil {
				return
			}
		}
	}
}
