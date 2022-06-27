package cmp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TextWithButtons struct {
	Msg     string
	Buttons tgbotapi.InlineKeyboardMarkup
	ChatId  int64
	SharedConf
}

func (t TextWithButtons) Execute(b *Bot) {
	m := tgbotapi.NewMessage(t.ChatId, t.Msg)
	m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(t.HideKeyboard)
	m.ReplyMarkup = t.Buttons
	msg, _ := b.BotApi.Send(m)
	b.LogMsg(msg)
}
