package cmp

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SharedConf struct {
	HideKeyboard bool
	Exec         bool
}

type Component interface {
	Execute(b *Bot)
	AddToScreen(s *Screen)
}

type Screen struct {
	Block []Component
}

func (s *Screen) Exec(b *Bot) {
	for _, bl := range s.Block {
		bl.Execute(b)
	}
}

type Chat struct {
	Id      int64
	Updates chan tgbotapi.Update
}
