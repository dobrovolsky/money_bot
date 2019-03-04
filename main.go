package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"

	tb "gopkg.in/tucnak/telebot.v2"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// NOTIFICATIONTIMEOUT is used for removing system notifications
const NOTIFICATIONTIMEOUT = 4 * time.Second

var db = &gorm.DB{}

func forwardToOldBot(s string) {
	url := os.Getenv("OLD_BOT_URL")

	var jsonStr = []byte(fmt.Sprintf(`{"message": {"text": "%s"}}`, s))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	Check(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()
}

func main() {
	err := godotenv.Load()
	Check(err)

	token := os.Getenv("TELEGRAM_TOKEN")

	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 30 * time.Second},
	})
	Check(err)

	db, err = gorm.Open("sqlite3", "db.sqlite3")
	Check(err)
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&User{}, &LogItem{}, &Category{})

	b.Handle("/start", func(m *tb.Message) {
		handleStart(m, b)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		if m.Sender.ID == 154701187 {
			forwardToOldBot(m.Text)
		}
		handleNewMessage(m, b)
	})

	b.Handle(tb.OnEdited, func(m *tb.Message) {
		handleEdit(m, b)
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		_, err := b.Send(m.Sender, "Sorry i don't support images 😓")
		Check(err)
	})

	b.Handle("/income", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "In development 💪")
		Check(err)
	})

	b.Handle("/stats", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "In development 💪")
		Check(err)
	})

	b.Handle("/export", func(m *tb.Message) {
		handleExport(m, b)
	})

	b.Start()
}
