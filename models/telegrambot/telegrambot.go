package telegrambot

import (
	"io/ioutil"
	"log"
	"net/http"
)

type TelegramBot struct {
	Token string
}

func (tbot TelegramBot) TelegramRequest(action string) string {
	return ("https://api.telegram.org/bot" + tbot.Token + "/" + action + "?")
}

func (tbot TelegramBot) SendError(message string, chatID string, markdown bool) {
	sendRequestURL := (tbot.TelegramRequest("sendMessage") + "?chat_id=" + chatID + "&text=" + message)
	if markdown {
		sendRequestURL += "&parse_mode=Markdown"
	}

	log.Println(MakeRequest(sendRequestURL))
}

func MakeRequest(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}