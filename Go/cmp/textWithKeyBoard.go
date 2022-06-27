package cmp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TextWithKeyboard struct {
	Msg      string
	Keyboard tgbotapi.ReplyKeyboardMarkup
	ChatId   int64
	SharedConf
}

func (t *TextWithKeyboard) Execute(b *Bot) {
	m := tgbotapi.NewMessage(t.ChatId, t.Msg)
	m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(t.HideKeyboard)
	m.ReplyMarkup = t.Keyboard
	b.BotApi.Send(m)
}
