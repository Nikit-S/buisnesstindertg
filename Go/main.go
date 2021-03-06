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

var Chats map[int64]cmp.Chat = make(map[int64]cmp.Chat)

func main() {

	DBconnStr := "host=" + "nbshtech.ru" + " user=" + os.Getenv("POSTGRES_USER") + " dbname=" + os.Getenv("POSTGRES_USER") + " password=" + os.Getenv("POSTGRES_PASSWORD") + " sslmode=disable"

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
	Bot.TinderStart = make(chan struct{})

	//rows, err := Bot.Db.Query(`SELECT id, part FROM public.chats`)
	//if err == nil {
	//	var id int64
	//	var part bool
	//	for rows.Next() {
	//		rows.Scan(&id, &part)
	//		Chats[id] = &cmp.Chat{}
	//		Chats[id].Updates = make(chan tgbotapi.Update)
	//		Chats[id].Id = id
	//		Chats[id].Part = part
	//	}
	//}

	for update := range updates {
		log.Println("got update")
		if update.Message != nil || update.CallbackQuery != nil {
			go func() {
				if update.Message != nil {
					Bot.LogMsg(*update.Message)
				}
			}()
			if _, ok := Chats[update.FromChat().ChatConfig().ChatID]; !ok {
				fmt.Println("starting new chat")
				v := cmp.Chat{
					Updates: make(chan tgbotapi.Update),
					Id:      update.FromChat().ChatConfig().ChatID}
				Chats[v.Id] = v
				go Bot.Db.Exec(`INSERT INTO public.chats
					(id, first_name, last_name, username)
					VALUES
					($1, $2, $3, $4)`,
					update.FromChat().ChatConfig().ChatID,
					update.FromChat().FirstName,
					update.FromChat().LastName,
					update.FromChat().UserName,
				)
				go src.StartChatWithUser(&Bot, &v)
			}
			if update.Message != nil {

				fmt.Println("BALANCER: TO_ID:", update.FromChat().ChatConfig().ChatID, "FROM_ID:", update.Message.From.ID)
			} else if update.CallbackQuery != nil {

				fmt.Println("BALANCER: TO_ID:", update.FromChat().ChatConfig().ChatID, "FROM_ID:", update.CallbackQuery.From.ID)
			}
			fmt.Println(Chats[update.FromChat().ChatConfig().ChatID].Id)
			Chats[update.FromChat().ChatConfig().ChatID].Updates <- update
		}
	}
}
