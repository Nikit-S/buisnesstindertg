package src

import (
	"fmt"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var commands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/register", Description: "Зарегистрироваться в Бизнес-Тиндере"},
)

func ExecCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
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
	case "register":
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введите свой номер телефона"))
		b.LogMsg(msg)
		upd = <-ch.Updates
		phone := upd.Message.Text
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Расскажите о себе в одном коротком сообщении"))
		b.LogMsg(msg)
		upd = <-ch.Updates
		description := upd.Message.Text
		_, err := b.Db.Exec(`UPDATE public.chats
			SET phone=$1, description=$2, part=true
			WHERE id=$3`,
			phone,
			description,
			upd.Message.From.ID)
		if err == nil {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Данные записаны! Скоро тебе придут данные твоего первого собеседника."))
			b.LogMsg(msg)
			go Participant(b, ch)
			return nil
		}
	}
	return fmt.Errorf("no such command")
}
func SimpleUser(b *cmp.Bot, ch *cmp.Chat) {
	commands.Scope.ChatID = ch.Id
	b.BotApi.Request(commands)
	txt := tgbotapi.NewMessage(ch.Id, "*Привет*, зарегистрируйся с _помощью команды_ /register")
	txt.ParseMode = "Markdown"
	msg, _ := b.BotApi.Send(txt)
	b.LogMsg(msg)
	for update := range ch.Updates {
		if update.Message.IsCommand() {
			if ExecCommand(b, ch, update) == nil {
				return
			}
		}
	}
}
