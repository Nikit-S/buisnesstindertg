package cmp

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	BotApi      *tgbotapi.BotAPI
	Msg         *tgbotapi.MessageConfig
	Upd         *tgbotapi.Update
	Db          *sql.DB
	TinderStart chan struct{}
}

func (s *Bot) Execute(c Component) {
	c.Execute(s)
}
func (b *Bot) CheckForRoot(update tgbotapi.Update) bool {
	var perm string
	row := b.Db.QueryRow(`SELECT perm from public.chats where id = $1`, update.Message.From.ID)

	row.Scan(&perm)
	if perm == "root" {
		return true
	}

	return false
}

func (b *Bot) CheckForUser(update tgbotapi.Update) bool {
	var name string
	row, _ := b.Db.Query(`SELECT u_name from users where user_t = 'trainee'`)
	for row.Next() {
		row.Scan(&name)
		if name == update.Message.Text {
			return true
		}
	}
	return false
}

func (b *Bot) ShowAllUsersTo(id int64) {
	var name string
	var phone string
	row, _ := b.Db.Query(`SELECT username, phone from public.chats`)
	for row.Next() {
		row.Scan(&name, &phone)
		b.BotApi.Send(tgbotapi.NewMessage(id, name+"\nphone: "+phone))
	}
}

func (b *Bot) SendAllUsersTo(r_id int64) {
	var id int
	var first_name, last_name, username, phone, description, perm string
	var part bool
	row, _ := b.Db.Query(`SELECT * from public.chats`)
	csvFile, err := os.Create("employee.csv")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	_ = csvwriter.Write([]string{"id", "first_name", "last_name", "username", "phone", "description", "perm", "part"})

	for row.Next() {
		row.Scan(&id, &first_name, &last_name, &username, &phone, &description, &perm, &part)
		_ = csvwriter.Write([]string{strconv.Itoa(id), first_name, last_name, username, phone, description, perm, strconv.FormatBool(part)})
	}
	csvwriter.Flush()
	csvFile.Close()
	//msg := tgbotapi.NewMessage(r_id, "Вот список людей")
	fp := tgbotapi.FilePath("employee.csv")
	b.BotApi.Send(tgbotapi.NewDocument(r_id, fp))

}

func (b *Bot) MakePairs(r_id int64) {
	var id int
	var first_name, last_name, username, phone, description string
	b.Db.Exec("TRUNCATE TABLE public.pairs")
	b.Db.Exec(`INSERT INTO public.pairs (id, first_name, last_name, username, phone, description)
	(SELECT id, first_name, last_name, username, phone, description 
	FROM public.chats
	WHERE part=true
	ORDER BY RANDOM ())`)
	csvFile, err := os.Create("pairs.csv")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	_ = csvwriter.Write([]string{"id", "first_name", "last_name", "username", "phone", "description"})

	row, _ := b.Db.Query("SELECT * FROM public.pairs")
	for row.Next() {
		row.Scan(&id, &first_name, &last_name, &username, &phone, &description)
		_ = csvwriter.Write([]string{strconv.Itoa(id), first_name, last_name, username, phone, description})
	}
	csvwriter.Flush()
	csvFile.Close()
	//msg := tgbotapi.NewMessage(r_id, "Вот список людей")
	fp := tgbotapi.FilePath("pairs.csv")
	b.BotApi.Send(tgbotapi.NewDocument(r_id, fp))

}

type Pair struct {
	Id_a int
	Id_b int
}

func (b *Bot) SendPairs(pairs []Pair) {
	var id int
	var first_name, last_name, username, phone, description string
	row, _ := b.Db.Query("SELECT * FROM public.pairs")
	for row.Next() {
		row.Scan(&id, &first_name, &last_name, &username, &phone, &description)
		var msg tgbotapi.MessageConfig
		str := fmt.Sprintf("Привет!\nВот твой собеседник:\n\n%v\n%v\n%v\n%v\n%v\n", first_name, last_name, "@"+username, phone, description)
		send := false
		for _, p := range pairs {
			if p.Id_a == id {
				msg = tgbotapi.NewMessage(int64(p.Id_b), str)
				send = true
				break
			} else if p.Id_b == id {
				send = true
				msg = tgbotapi.NewMessage(int64(p.Id_a), str)
				break
			}
		}
		if send {
			if msg.ChatID == -1 {
				t, _ := b.BotApi.Send(tgbotapi.NewMessage(int64(id), "к сожалению все чаты заняты, подожди когда освободиться собеседник"))
				b.LogMsg(t)
			} else {
				t, _ := b.BotApi.Send(msg)
				b.LogMsg(t)
			}
		}

	}
}

func (b *Bot) SetIds(update tgbotapi.Update) {
	fmt.Println("Setting id for root: ", update.Message.Text)
	res, err := b.Db.Exec(`UPDATE public.users
	SET chat_id=$1,user_id=$2, username=$3
	WHERE u_name=$4;`,
		update.Message.Chat.ID,
		update.Message.From.ID,
		update.Message.Chat.UserName,
		update.Message.Text)
	fmt.Println(res, err)
}

func (b *Bot) Notify(query string, msg *tgbotapi.Message) {
	row, _ := b.Db.Query(query)
	var chatId int64
	for row.Next() {
		row.Scan(&chatId)
		if msg.Photo != nil {
			file, _ := b.BotApi.GetFile(tgbotapi.FileConfig{FileID: msg.Photo[len(msg.Photo)-1].FileID})
			log.Println("https://api.telegram.org/file/bot" + b.BotApi.Token + "/" + file.FilePath)
			resp, err := http.Get(file.Link(b.BotApi.Token))
			by, _ := ioutil.ReadAll(resp.Body)
			if err == nil {
				ph := tgbotapi.NewPhoto(chatId, tgbotapi.FileBytes{Bytes: by, Name: "photo"})
				ph.Caption = msg.Caption
				ph.ParseMode = "Markdown"
				b.BotApi.Send(ph)
				return
			}
		}
		txt := tgbotapi.NewMessage(chatId, msg.Text)
		txt.ParseMode = "Markdown"
		m, err := b.BotApi.Send(txt)
		if err == nil {
			b.LogMsg(m)
		}
	}
}

func (b *Bot) TestMessage(msg *tgbotapi.Message) {
	if msg.Photo != nil {
		file, _ := b.BotApi.GetFile(tgbotapi.FileConfig{FileID: msg.Photo[len(msg.Photo)-1].FileID})
		log.Println("https://api.telegram.org/file/bot" + b.BotApi.Token + "/" + file.FilePath)
		resp, err := http.Get(file.Link(b.BotApi.Token))
		by, _ := ioutil.ReadAll(resp.Body)
		if err == nil {
			ph := tgbotapi.NewPhoto(msg.From.ID, tgbotapi.FileBytes{Bytes: by, Name: "photo"})
			ph.Caption = msg.Caption
			ph.ParseMode = "Markdown"
			b.BotApi.Send(ph)
			return
		}
	}
	txt := tgbotapi.NewMessage(msg.From.ID, msg.Text)
	txt.ParseMode = "Markdown"
	m, err := b.BotApi.Send(txt)
	if err == nil {
		b.LogMsg(m)
	}
}

func (b *Bot) LogMsg(msg tgbotapi.Message) {
	b.Db.Exec(`INSERT INTO public.messages
			(chat_id, m_id, is_bot, from_id, m_time, txt)
			VALUES
			($1, $2, $3, $4, $5, $6)`,
		msg.Chat.ID,
		msg.MessageID,
		msg.From.IsBot,
		msg.From.ID,
		time.Unix(int64(msg.Date), 0),
		msg.Text,
	)
}

type User struct {
	Id          int
	FirstName   string
	LastName    string
	Username    string
	Phone       string
	Description string
}

func (b *Bot) GetParticipants(chat *Chat) []User {
	user := User{}
	users := []User{}
	row, err := b.Db.Query(`SELECT id, first_name, last_name, username, phone, description FROM public.chats 
	WHERE chats.id != $1 AND chats.part = true
	AND chats.id NOT IN 
		(SELECT pair_id FROM public.user_pair_history
			WHERE id = $1)
		`, chat.Id)
	if err != nil {
		log.Println("sql error in GetParticipants")
		return nil
	}
	for row.Next() {
		if err = row.Scan(&user.Id, &user.FirstName, &user.LastName,
			&user.Username, &user.Phone, &user.Description); err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Println(user)
			users = append(users, user)
		}
	}

	fmt.Println("total users got from sql:", users)
	return users
}

func (b *Bot) GetMatchesForId(chat *Chat) []User {
	user := User{}
	users := []User{}
	row, err := b.Db.Query(`SELECT id, first_name, last_name, username, phone, description FROM public.chats 
	WHERE id IN 
		(SELECT id FROM public.user_pair_history
			WHERE pair_id = $1 AND want = true )
		AND
		id IN 
		(SELECT pair_id FROM public.user_pair_history
			WHERE id = $1 AND want = true )
		AND part = true
		`, chat.Id)
	if err != nil {
		log.Println("sql error in GetParticipants")
		return nil
	}
	for row.Next() {
		if err = row.Scan(&user.Id, &user.FirstName, &user.LastName,
			&user.Username, &user.Phone, &user.Description); err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Println(user)
			users = append(users, user)
		}
	}

	fmt.Println("total users got from sql:", users)
	return users
}

func (b *Bot) GetParticipantsWithDecline(chat *Chat) []User {
	user := User{}
	users := []User{}
	row, err := b.Db.Query(`SELECT id, first_name, last_name, username, phone, description FROM public.chats 
	WHERE chats.id != $1 AND chats.part = true
	AND chats.id NOT IN 
		(SELECT pair_id FROM public.user_pair_history
			WHERE id = $1 AND want != false)
		`, chat.Id)
	if err != nil {
		log.Println("sql error in GetParticipants")
		return nil
	}
	for row.Next() {
		if err = row.Scan(&user.Id, &user.FirstName, &user.LastName,
			&user.Username, &user.Phone, &user.Description); err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Println(user)
			users = append(users, user)
		}
	}

	fmt.Println("total users got from sql:", users)
	return users
}

func (b *Bot) RegisterPair(ch *Chat, user User, sign bool) {
	b.Db.Exec(`INSERT INTO public.user_pair_history
			(id, pair_id, want)
			VALUES
			($1, $2, $3)
			ON CONFLICT (id, pair_id) DO
			UPDATE
			SET
			want=$3`,
		ch.Id,
		user.Id,
		sign,
	)
}
