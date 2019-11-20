package telegrambot

import (
	"io/ioutil"
	"log"
	"net/http"
	"gopkg.in/yaml.v2"
	"github.com/ruslanBik4/httpgo/logs"
)

type TelegramBot struct {
	Token string `yaml:"BotToken"`
	ChatID string `yaml:"ChatID"`
}

// TelegramRequest makes base url for request
func (tbot TelegramBot) TelegramRequest(action string) string {
	return ("https://api.telegram.org/bot" + tbot.Token + "/" + action + "?")
}

// SendMessage is used for sending messages
func (tbot TelegramBot) SendMessage(message string, markdown bool) {
	sendRequestURL := (tbot.TelegramRequest("sendMessage") + "chat_id=" + tbot.ChatID + "&text=" + message)

	// For using bold and italic font in message sent
	if markdown {
		sendRequestURL += "&parse_mode=Markdown"
	}

	log.Println(MakeRequest(sendRequestURL))
}

// MakeRequest executes request and gets response? converting it to string
func MakeRequest(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		logs.ErrorLog(err, "TelegramBot")
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.ErrorLog(err, "TelegramBot")
	}

	return string(body)
}

// GetNewTelegramBot reads a config file for bot token and chatID and creates new TelegramBot struct
func GetNewTelegramBot(confPath string) (tb *TelegramBot, err error) {

	yamlFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		logs.ErrorLog(err)
	}
	err = yaml.Unmarshal(yamlFile, &tb)
	if err != nil {
		logs.ErrorLog(err)
	}

	return
}