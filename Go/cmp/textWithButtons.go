package cmp

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TextWithButtons struct {
	Msg     string
	Buttons tgbotapi.InlineKeyboardMarkup
	ChatId  int64
	SharedConf
}

func (t *TextWithButtons) Execute(b *Bot) {
	b.Msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(t.HideKeyboard)
	b.Msg.Text = t.Msg
	b.Msg.ReplyMarkup = t.Buttons
	b.BotApi.Send(b.Msg)
}
