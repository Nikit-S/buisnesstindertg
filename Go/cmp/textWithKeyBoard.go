package cmp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TextWithKeyboard struct {
	Msg      string
	Keyboard tgbotapi.ReplyKeyboardMarkup
	ChatId   int64
	SharedConf
}

func (t *TextWithKeyboard) Execute(b *Bot) {
	b.Msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(t.HideKeyboard)
	b.Msg.Text = t.Msg
	b.Msg.ReplyMarkup = t.Keyboard
	b.BotApi.Send(b.Msg)
}
