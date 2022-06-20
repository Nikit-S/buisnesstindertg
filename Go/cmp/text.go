package cmp

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Text struct {
	Msg    string
	ChatId int64
	SharedConf
}

func (t Text) Execute(b *Bot) {
	b.Msg.Text = t.Msg
	b.Msg.ChatID = t.ChatId
	b.Msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(t.HideKeyboard)
	b.BotApi.Send(b.Msg)
}

func (t Text) AddToScreen(s *Screen) {
	s.Block = append(s.Block, t)
}
