package src

import (
	"fmt"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var commands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/register", Description: "Зарегистрироваться в Бизнес-Тиндере"},
)

func ExecCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
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
	case "register":
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Пришли, пожалуйста, свой номер телефона"))
		b.LogMsg(msg)
		upd = <-ch.Updates
		phone := upd.Message.Text
		txt := tgbotapi.NewMessage(ch.Id,
			`Расскажи о себе *(в одном сообщении)* — нам нужно знать, как презентовать тебя твоему партнеру по диджитал-знакомствам.`)
		//`Расскажи о себе *(в одном сообщении)* — нам нужно знать, как презентовать тебя`)
		txt.ParseMode = "Markdown"
		msg, _ = b.BotApi.Send(txt)
		b.LogMsg(msg)
		txt = tgbotapi.NewMessage(ch.Id,
			`Подумай, что ты сам хотел бы узнать о своём новом друге. Сфера, в которой он хорош? Опыт работы с разными нишами? Проекты, которыми он гордится? Может, селебрити, с которыми работал? Любимый цвет или канал в Телеграме?
			
			Вот все это и нужно написать о себе. Но помни: это не собеседование — вы уже в нашем клане самых крутых ребят. Пиши живо и в свободной форме — помни, что читать это будет какой-то классный человек, и ему должно быть интересно.
			`)
		//`Подумай, что ты сам хотел бы узнать.
		//`)
		txt.ParseMode = "Markdown"
		msg, _ = b.BotApi.Send(txt)
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
			msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Аж самим захотелось с тобой познакомиться! Начинаем искать!"))
			b.LogMsg(msg)
			go Participant(b, ch)
			return nil
		}
	}
	return fmt.Errorf("no such command")
}
func SimpleUser(b *cmp.Bot, ch *cmp.Chat) {
	commands.Scope.ChatID = ch.Id
	b.BotApi.Request(commands)
	txt := tgbotapi.NewMessage(ch.Id, "*Привет!* Если ты здесь, значит, принял правильное решение становиться сильным, ловким и умелым вместе с Up&Boost ❤️‍🔥")
	txt.ParseMode = "Markdown"
	msg, _ := b.BotApi.Send(txt)
	b.LogMsg(msg)
	txt = tgbotapi.NewMessage(ch.Id, "И один из инструментов твоей прокачки как крутого специалиста — диджитал-тиндер, где ты сможешь знакомиться каждую неделю с новым человеком из своей среды.")
	txt.ParseMode = "Markdown"
	msg, _ = b.BotApi.Send(txt)
	b.LogMsg(msg)
	txt = tgbotapi.NewMessage(ch.Id,
		`Вот несколько моментов, которые тебе нужно знать:
	
	Всё как в тиндере, «интересно» «неинтересно» — твои главные кнопки
	Вы коннектитесь и сами решаете, как часто и о чем будете болтать. Наш совет — используйте каждое знакомство на максимум и расширяйте свой пул полезных связей.
	Не только получайте, но и давайте. Больше рассказывайте о себе и говорите о том, чем вы можете быть полезны.
	Будьте вежливы, приятны и нетоксичны — все как в жизни, только теперь ещё и с пользой для работы. Ваша репутация формируется прямо сейчас.
	Мы будем помогать вам словить тот самый мэтч и дзынь. Ждите от нас подсказки, как построить крутую коммуникацию — мы рядом.
	Если вдруг что-то идёт не так, и вы хотели бы поговорить об этом, пишите нам вот сюда: @Avantiina_Team`)
	txt.ParseMode = ""
	msg, _ = b.BotApi.Send(txt)
	b.LogMsg(msg)
	b.Execute(cmp.WaitBlock{ChatId: ch.Id, Typing: true, Time: 10})
	txt = tgbotapi.NewMessage(ch.Id,
		`Если всё понятно жми команду /register
		`)
	txt.ParseMode = "Markdown"
	msg, _ = b.BotApi.Send(txt)
	b.LogMsg(msg)
	for update := range ch.Updates {
		if update.Message != nil && update.Message.IsCommand() {
			if ExecCommand(b, ch, update) == nil {
				return
			}
		}
	}
}
