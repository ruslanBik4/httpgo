package telegrambot

import (
	"io/ioutil"
	"net/http"

	"github.com/ruslanBik4/httpgo/logs"
	"gopkg.in/yaml.v2"
)

type TelegramBot struct {
	Token  string `yaml:"BotToken"`
	ChatID string `yaml:"ChatID"`
}

// NewTelegramBot reads a config file for bot token and chatID and creates new TelegramBot struct
func NewTelegramBot(confPath string) (tb *TelegramBot, err error) {

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

// TelegramRequest makes base url for request
func (tbot TelegramBot) TelegramRequest(action string) string {
	return ("https://api.telegram.org/bot" + tbot.Token + "/" + action + "?")
}

// SendMessage is used for sending messages
func (tbot TelegramBot) SendMessage(message string, markdown bool) error {
	sendRequestURL := (tbot.TelegramRequest("sendMessage") + "chat_id=" + tbot.ChatID + "&text=" + message)

	// For using bold and italic font in message sent
	if markdown {
		sendRequestURL += "&parse_mode=Markdown"
	}

	err := MakeRequest(sendRequestURL)
	if err != nil {
		return err
	}
	return nil
}

// MakeRequest executes request and gets response converting it to string
func MakeRequest(url string) error {
	_, err := http.Get(url)

	if err != nil {
		return err
	}
	return nil

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	logs.ErrorLog(err, "TelegramBot")
	// }
	// return string(body)
}



// TelegramBotHandler reads bot params from configPath and accepts some log struct to find if its needed to print some mess to telegram bot
func (tbot TelegramBot) Write(message []byte) error {
	mess := string(message)
	err := tbot.SendMessage(mess, false)
	if err != nil {
		return err
	}
	return nil
}
