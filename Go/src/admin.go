package src

import (
	"fmt"
	"strconv"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var rootCommands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/show_all_users", Description: "Показывает всю таблицу"},
	tgbotapi.BotCommand{Command: "/make_pairs", Description: "Выгрузить всех участников"},
	tgbotapi.BotCommand{Command: "/register", Description: "Зарегистрироваться"},
	tgbotapi.BotCommand{Command: "/unregister", Description: "Отменить регистрацию"},
	//tgbotapi.BotCommand{Command: "/make_root", Description: "Выгрузить всех участников"},
	tgbotapi.BotCommand{Command: "/start_bt", Description: "Начать бизнес-тиндер"},
	tgbotapi.BotCommand{Command: "/notify_pairs", Description: "Написать всем участникам"},
	tgbotapi.BotCommand{Command: "/test_message", Description: "Посмотреть как сверстается сообщение"},
)

func ExecRootCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
	//if !b.CheckForRoot(upd) {
	//	return fmt.Errorf("You are not root")
	//}
	switch upd.Message.Command() {
	case "test_message":
		msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введи текст сообщения, которое ты хочешь протестировать"))
		b.LogMsg(msg)
		upd := <-ch.Updates
		msg_txt := upd.Message
		b.TestMessage(msg_txt)
	case "show_all_users":
		b.ShowAllUsersTo(ch.Id)
		b.SendAllUsersTo(ch.Id)
	case "make_pairs":
		b.MakePairs(ch.Id)
	//case "make_root":
	//	msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введи айди пользователя, статус которого нужно изменить"))
	//	b.LogMsg(msg)
	//	upd := <-ch.Updates
	//	id := upd.Message.Text
	//	b.ChangeRoot(id)
	case "notify_pairs":
		b.MakePairs(ch.Id)
		msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введи текст сообщения, которое увидят все кто участвует"))
		b.LogMsg(msg)
		upd := <-ch.Updates
		msg_txt := upd.Message
		b.Notify("Select id from public.pairs", msg_txt)
	case "start_bt":
		go BtConsole(b, ch)
	case "unregister":
		_, err := b.Db.Exec(`UPDATE public.chats
			SET part=false
			WHERE id=$1`,
			upd.Message.From.ID)
		if err == nil {
			msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Ты больше не участвуешь в Бизнес-Тиндере"))
			b.LogMsg(msg)
		}
	case "register":
		msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введите свой номер телефона"))
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
		}
	}

	return fmt.Errorf("exited")
}

func createPairsList(data [][]string) []cmp.Pair {
	var pairs []cmp.Pair
	for i, line := range data {
		if i > 0 { // omit header line
			var rec cmp.Pair
			for j, field := range line {
				if j == 0 {
					rec.Id_a, _ = strconv.Atoi(field)
				} else if j == 1 {
					rec.Id_b, _ = strconv.Atoi(field)
				}
			}
			pairs = append(pairs, rec)
		}
	}
	return pairs
}

func Root(b *cmp.Bot, ch *cmp.Chat) {

	rootCommands.Scope.ChatID = ch.Id
	b.BotApi.Request(rootCommands)
	for update := range ch.Updates {
		if update.Message.IsCommand() {
			if ExecRootCommand(b, ch, update) == nil {
				return
			}
		}
	}
}
