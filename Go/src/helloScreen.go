package src

import (
	"testbot/cmp"
)

func HelloScreen(chatId int64) *cmp.Screen {
	scr := &cmp.Screen{}
	txt := cmp.Text{
		Msg:    `Привет`,
		ChatId: chatId,
	}
	txt.AddToScreen(scr)

	wait := cmp.WaitBlock{
		Typing: true,
		ChatId: chatId,
		Time:   1,
	}
	wait.AddToScreen(scr)

	txt.Msg = "Это бот для прохождения стажировки в WB"
	wait.AddToScreen(scr)

	wait = cmp.WaitBlock{
		Typing: true,
		ChatId: chatId,
		Time:   1,
	}
	txt.AddToScreen(scr)

	txt.Msg = "Давай знакомиться, пришли мне свои ФИО разделенные ОДНИМ пробелом\n\nНапример:\nИванов Иван Иванович"
	txt.AddToScreen(scr)
	return scr
}
