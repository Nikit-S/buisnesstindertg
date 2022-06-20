package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testbot/cmp"
	"testbot/src"

	_ "github.com/lib/pq"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Chats map[int64]*cmp.Chat = make(map[int64]*cmp.Chat)

func main() {

	DBconnStr := "host=" + "nbshtech.ru" + " user=" + "wb_order" + " dbname=" + "postgres" + " password=" + "mypass" + " sslmode=disable"

	Bot := cmp.Bot{}
	var err error
	Bot.Db, err = sql.Open("postgres", DBconnStr)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhookWithCert("https://nbshtech.ru:8443/"+bot.Token, tgbotapi.FilePath("cert.pem"))

	if err != nil {
		log.Println("newwebhook", err)
	}

	bot.Request(wh)

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	Bot.BotApi = bot
	Bot.Msg = &tgbotapi.MessageConfig{}

	for update := range updates {
		log.Println("got update")
		//if update.CallbackQuery != nil {
		//	Bot.BotApi.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Отлично!"))
		//}
		if update.Message != nil {
			go Bot.LogMsg(*update.Message)
			if v, ok := Chats[update.Message.Chat.ID]; ok {
				v.Updates <- update
			} else {
				fmt.Println("starting new chat")
				v := &cmp.Chat{
					Updates: make(chan tgbotapi.Update),
					Id:      update.Message.Chat.ID}
				Chats[v.Id] = v
				go Bot.Db.Exec(`INSERT INTO public.chats
					(id, first_name, last_name, username)
					VALUES
					($1, $2, $3, $4)`,
					update.Message.From.ID,
					update.Message.From.FirstName,
					update.Message.From.LastName,
					update.Message.From.UserName,
				)
				go src.StartChatWithUser(&Bot, v)
				v.Updates <- update
			}
		}
	}
}
