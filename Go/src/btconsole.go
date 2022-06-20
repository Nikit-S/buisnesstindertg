package src

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"testbot/cmp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var btCommands = tgbotapi.NewSetMyCommandsWithScope(
	tgbotapi.BotCommandScope{Type: "chat", ChatID: -1},
	tgbotapi.BotCommand{Command: "/msg_to_pairs", Description: "Отправить текст парам"},
	tgbotapi.BotCommand{Command: "/next_pair", Description: "Выслать ссылку на новую пару"},
	tgbotapi.BotCommand{Command: "/stop_bt", Description: "Закончить бизнес-тиндер"},
)

func ExecBtCommand(b *cmp.Bot, ch *cmp.Chat, upd tgbotapi.Update) error {
	//if !b.CheckForRoot(upd) {
	//	return fmt.Errorf("You are not root")
	//}
	switch upd.Message.Command() {
	case "msg_to_pairs":
		msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Введи текст сообщения, которое увидят все кто участвует"))
		b.LogMsg(msg)
		upd := <-ch.Updates
		msg_txt := upd.Message
		b.Notify("Select id from public.pairs", msg_txt)
	case "next_pair":
		NextPair()
		b.BotApi.Send(tgbotapi.NewMessage(ch.Id, fmt.Sprintf("First Pairs are ready: %v\n", pairsList)))
		b.SendPairs(pairsList)
	case "stop_bt":
		pairsList = []cmp.Pair{}
		go Root(b, ch)
		return nil
	}

	return fmt.Errorf("no such command")
}

var pairsList []cmp.Pair

func StartPairs(b *cmp.Bot, ch *cmp.Chat) error {
	msg, _ := b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Прикрепи файл с парами в формате csv"))
	b.LogMsg(msg)
	upd := <-ch.Updates
	if upd.Message.Document == nil {
		msg, _ = b.BotApi.Send(tgbotapi.NewMessage(ch.Id, "Не файл! Прерывание"))
		b.LogMsg(msg)
		return fmt.Errorf("Not correct file")
	}
	doc, _ := b.BotApi.GetFile(tgbotapi.FileConfig{FileID: upd.Message.Document.FileID})
	log.Println("https://api.telegram.org/file/bot" + b.BotApi.Token + "/" + doc.FilePath)
	file, err := http.Get("https://api.telegram.org/file/bot" + b.BotApi.Token + "/" + doc.FilePath)
	if err != nil {
		log.Println("reader", err)
		return err
	}
	csvReader := csv.NewReader(file.Body)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Println("reader", err)
		return err
	}
	pairsList = createPairsList(data)
	return nil
}

func NextPair() {
	temp := pairsList[0].Id_b
	for i := range pairsList {
		if i == 0 {
			continue
		}
		pairsList[i-1].Id_b = pairsList[i].Id_b
	}
	pairsList[len(pairsList)-1].Id_b = temp

}

func BtConsole(b *cmp.Bot, ch *cmp.Chat) {

	btCommands.Scope.ChatID = ch.Id
	b.BotApi.Request(btCommands)
	if StartPairs(b, ch) != nil {
		go Root(b, ch)
		return
	}
	b.BotApi.Send(tgbotapi.NewMessage(ch.Id, fmt.Sprintf("First Pairs are ready: %v\n", pairsList)))
	b.SendPairs(pairsList)
	for update := range ch.Updates {
		if update.Message.IsCommand() {
			if ExecBtCommand(b, ch, update) == nil {
				return
			}
		}
	}
}
