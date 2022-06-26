package src

import (
	"fmt"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var partCommands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/unregister", Description: "Отменить участие в бизнес-тиндере"},
	tgbotapi.BotCommand{Command: "/get_pair", Description: "Подобрать новые пары"},
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
	case "get_pair":
		var users []cmp.User
		users = b.GetParticipants(ch, false)
		pair_index := 0
		if len(users) == 0 {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Мы пока не подобрали тебе пару, можешь попробовать езе раз через команду /get_pair"))
			b.LogMsg(msg)
		} else {
			str := fmt.Sprintf(`
		Имя: %s
		О себе: %s
		`, users[pair_index].FirstName, users[pair_index].LastName)

			m.ChatId = ch.Id
			m.Msg = str
			b.Execute(m)
		}
		return fmt.Errorf("still this state with new pairs")
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

var m cmp.TextWithButtons = cmp.TextWithButtons{
	Buttons: tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Интересно", "true"),
			tgbotapi.NewInlineKeyboardButtonData("Мимо", "false"),
		),
	),
}

func Participant(b *cmp.Bot, ch *cmp.Chat) {
	partCommands.Scope.ChatID = ch.Id
	b.BotApi.Request(partCommands)
	msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Если захочешь отказаться от участия напиши /unregister"))
	b.LogMsg(msg)
	b.Execute(cmp.WaitBlock{ChatId: ch.Id, Typing: true, Time: 3})
	msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Скоро мы предложим тебе несколько человек для общения"))
	b.LogMsg(msg)
	var users []cmp.User
	users = b.GetParticipants(ch, false)
	fmt.Println("got users: ", users)
	b.Execute(cmp.WaitBlock{ChatId: ch.Id, Typing: true, Time: 10})
	pair_index := 0
	if len(users) == 0 {
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Мы пока не подобрали тебе пару, можешь попробовать езе раз через команду /get_pair или пройтись по второму кругу /one_more_time"))
		b.LogMsg(msg)
	} else {
		str := fmt.Sprintf(`
		Имя: %s
		О себе: %s
		`, users[pair_index].FirstName, users[pair_index].LastName)

		m.ChatId = ch.Id
		m.Msg = str
		b.Execute(m)

	}
	for update := range ch.Updates {
		if update.Message != nil && update.Message.IsCommand() {
			if ExecPartCommand(b, ch, update) == nil {
				return
			}
		}
		if update.CallbackQuery != nil {
			switch update.CallbackData() {
			case "true":
				b.RegisterPair(ch, users[pair_index], true)
			case "false":
				b.RegisterPair(ch, users[pair_index], false)
			}
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			b.BotApi.Request(callback)
			pair_index++
			if pair_index < len(users) {

				str := fmt.Sprintf(`
					Имя: %s
					О себе: %s
					`, users[pair_index].FirstName, users[pair_index].LastName)

				m.ChatId = ch.Id
				m.Msg = str
				b.Execute(m)
			} else {
				msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Все пары для тебя закончились, попробуй через какое то время найти новую пару через команду /get_pair"))
				b.LogMsg(msg)
			}
		}
	}
}
