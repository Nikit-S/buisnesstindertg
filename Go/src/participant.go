package src

import (
	"fmt"
	"strings"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type participantStruct struct {
	users  []cmp.User
	uIndex int
	pairs  []cmp.User
	pIndex int
}

var partCommands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/unregister", Description: "Отменить участие в бизнес-тиндере"},
	tgbotapi.BotCommand{Command: "/get_matches", Description: "Проверить свои Мэтчи"},
	tgbotapi.BotCommand{Command: "/get_pair", Description: "Подобрать новые пары"},
	tgbotapi.BotCommand{Command: "/one_more_time", Description: "Подобрать новые пары и пересмотреть старые отказы"},
)

func (p *participantStruct) ExecPartCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
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
		fmt.Println("CHAT ID:", ch.Id)
		p.users = b.GetParticipants(ch)
		fmt.Println("got users: ", p.users)
		//b.Execute(cmp.WaitBlock{ChatId: ch.Id, Typing: true, Time: 1})
		pair_index := 0
		if len(p.users) == 0 {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Мы пока не подобрали тебе пару, можешь попробовать ещё раз через команду /get_pair или пройтись по второму кругу /one_more_time"))
			b.LogMsg(msg)
		} else {
			str := fmt.Sprintf(`
		Имя: %s
		О себе: %s
		`, p.users[pair_index].FirstName, p.users[pair_index].Description)
			mt := m
			mt.ChatId = ch.Id
			fmt.Println("CHAT ID:", ch.Id)
			mt.Msg = str
			b.Execute(mt)
		}
		return fmt.Errorf("Not a root")
	case "one_more_time":
		p.users = b.GetParticipantsWithDecline(ch)
		fmt.Println("got users: ", p.users)
		b.Execute(cmp.WaitBlock{ChatId: ch.Id, Typing: true, Time: 1})
		pair_index := 0
		if len(p.users) == 0 {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Мы пока не подобрали тебе пару, можешь попробовать ещё раз через команду /get_pair или пройтись по второму кругу /one_more_time"))
			b.LogMsg(msg)
		} else {
			str := fmt.Sprintf(`
		Имя: %s
		О себе: %s
		`, p.users[pair_index].FirstName, p.users[pair_index].Description)
			mt := m
			mt.ChatId = ch.Id
			mt.Msg = str
			b.Execute(mt)
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
	case "get_matches":
		p.pairs = b.GetMatchesForId(ch)
		p.pIndex = 0
		if len(p.pairs) == 0 {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Пока что мэтчей нет, можешь попробовать расширить свою базу через команду /get_pair или пройтись по второму кругу /one_more_time\nНу или поробуй ввести команду /get_matches еще раз через какое-то время"))
			b.LogMsg(msg)
		} else {
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Ура! У тебя есть мэтчи! Давай посмотрим. "))
			b.LogMsg(msg)
			str := fmt.Sprintf(`
Имя: %s
Фамилия: %s
О себе: %s
Ник: %s
Телефон: %s
					`,
				p.pairs[p.pIndex].FirstName,
				p.pairs[p.pIndex].LastName,
				p.pairs[p.pIndex].Description,
				p.pairs[p.pIndex].Username,
				p.pairs[p.pIndex].Phone,
			)
			mt := m2
			mt.ChatId = ch.Id
			mt.Msg = str
			b.Execute(mt)
			p.pIndex++
			return fmt.Errorf("Same state")
		}
	}
	return fmt.Errorf("no such command")
}

var m cmp.TextWithButtons = cmp.TextWithButtons{
	Buttons: tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Интересно", "user_true"),
			tgbotapi.NewInlineKeyboardButtonData("Мимо", "user_false"),
		),
	),
}

var m2 cmp.TextWithButtons = cmp.TextWithButtons{
	Buttons: tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Дальше", "pair_next"),
		),
	),
}

func Participant(b *cmp.Bot, ch *cmp.Chat) {
	partCommands.Scope.ChatID = ch.Id
	b.BotApi.Request(partCommands)
	msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id,
		`Если захочешь отказаться от участия то нажимай /unregister
А как захочется узнать с кем у вас мэтч нажимай /get_matches`))
	b.LogMsg(msg)

	var part participantStruct
	msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Если будешь готов подобрать пару жми /get_pair или пройти по второму кругу /one_more_time"))
	b.LogMsg(msg)
	var err error
	for update := range ch.Updates {
		if update.Message != nil && update.Message.IsCommand() {
			if err = part.ExecPartCommand(b, ch, update); err == nil {
				return
			}
		}
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			b.BotApi.Request(callback)
			if strings.Contains(update.CallbackData(), "user") {
				switch update.CallbackData() {
				case "user_true":
					b.RegisterPair(ch, part.users[part.uIndex], true)
				case "user_false":
					b.RegisterPair(ch, part.users[part.uIndex], false)
				}
				part.uIndex++
				if part.uIndex < len(part.users) {
					str := fmt.Sprintf(`
						Имя: %s
						О себе: %s
						`, part.users[part.uIndex].FirstName, part.users[part.uIndex].Description)
					mt := m
					mt.ChatId = ch.Id
					mt.Msg = str
					b.Execute(mt)
				} else {
					msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Кажется пока что пары закончились, можешь попробовать ещё раз через команду /get_pair или пройтись по второму кругу /one_more_time"))
					b.LogMsg(msg)
					part.uIndex = 0
				}
			} else if strings.Contains(update.CallbackData(), "pair") {
				switch update.CallbackData() {
				case "pair_next":
					if part.pIndex < len(part.pairs) {
						str := fmt.Sprintf(`
Имя: %s
Фамилия: %s
О себе: %s
Ник: %s
Телефон: %s`,
							part.pairs[part.pIndex].FirstName,
							part.pairs[part.pIndex].LastName,
							part.pairs[part.pIndex].Description,
							part.pairs[part.pIndex].Username,
							part.pairs[part.pIndex].Phone,
						)
						mt := m2
						mt.ChatId = ch.Id
						mt.Msg = str
						b.Execute(mt)
						part.pIndex++
					} else {
						part.pIndex = 0
						msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Пока что мэтчей нет, можешь попробовать расширить свою базу через команду /get_pair или пройтись по второму кругу /one_more_time\nНу или поробуй ввести команду /get_matches еще раз через какое-то время"))
						b.LogMsg(msg)
					}
				}
			}
		}
	}
}
