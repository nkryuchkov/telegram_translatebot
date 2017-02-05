package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	// replace by your Telegram token
	telegramToken        = `your Telegram token`
	
	// replace by your Yandex.Translate token
	yandextranslateToken = `your Yandex.Translate token`
)

// Received struct contains unmarshalled responses from Yandex.Translate
type Received struct {
	Code int      `json:"code"`
	Lang string   `json:"lang"`
	Text []string `json:"text"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}

	// set it to false if you don't need debug output
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%d] %s %s wrote: %s", update.Message.Chat.ID,
			update.Message.From.FirstName, update.Message.From.LastName,
			update.Message.Text)

		var answMsg string

		// if the user sends less than 2 words print help message
		if len(strings.Split(update.Message.Text, " ")) < 2 {
			answMsg = "Usage: [language] [message]"
		} else {
			// use Yandex.Translate service
			URL, err := url.Parse("https://translate.yandex.net")
			if err != nil {
				panic("URL incorrect")
			}

			URL.Path += "/api/v1.5/tr.json/translate"
			parameters := url.Values{}
			parameters.Add("key", yandextranslateToken)
			parameters.Add("lang", strings.Split(update.Message.Text, " ")[0])
			parameters.Add("text",
				strings.Join(strings.Split(update.Message.Text, " ")[1:], " "))
			parameters.Add("format", "plain")
			URL.RawQuery = parameters.Encode()

			resp, err := http.Get(URL.String())
			if err != nil {
				log.Fatal(err)
			}
			answer, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			data := Received{}
			json.Unmarshal(answer, &data)

			answMsg = data.Text[0]
		}

		replymsg := tgbotapi.NewMessage(update.Message.Chat.ID, answMsg)
		replymsg.ReplyToMessageID = update.Message.MessageID

		log.Printf("Translation: %s", answMsg)

		bot.Send(replymsg)
	}
}
