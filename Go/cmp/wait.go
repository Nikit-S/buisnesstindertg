package cmp

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WaitBlock struct {
	ChatId int64
	Typing bool
	Time   int64
	SharedConf
}

func (w WaitBlock) Execute(b *Bot) {
	b.Msg.ChatID = w.ChatId
	if w.Typing {
		fmt.Println("Sleeping and typing")
		b.BotApi.Send(tgbotapi.NewChatAction(w.ChatId, tgbotapi.ChatTyping))
	}
	time.Sleep(time.Duration(w.Time) * time.Second)
}

func (w WaitBlock) AddToScreen(s *Screen) {
	s.Block = append(s.Block, w)
}
