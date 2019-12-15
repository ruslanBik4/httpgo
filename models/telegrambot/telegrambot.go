package telegrambot

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"

	"github.com/ruslanBik4/httpgo/logs"
)

// TelegramBot struct with tiken and one chatid
type TelegramBot struct {
	Token          string `yaml:"BotToken"`
	ChatID         string `yaml:"ChatID"`
	RequestURL     string
	Request        *fasthttp.Request
	Response       *fasthttp.Response
	FastHTTPClient *fasthttp.Client
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

	tb.RequestURL = baseURL
	tb.Request = &fasthttp.Request{}
	tb.Request.Header.SetMethod(fasthttp.MethodGet)
	tb.FastHTTPClient = &fasthttp.Client{}

	return
}

// SetRequestURL makes url for request
func (tbot *TelegramBot) SetRequestURL(action string, otherRequest string, markdown bool) {
	tbot.RequestURL = (baseURL + tbot.Token + "/" + action + "?" + otherRequest)
	if markdown {
		tbot.RequestURL += "&parse_mode=Markdown"
	}
	tbot.Request.SetRequestURI(tbot.RequestURL)

}

// SendMessage is used for sending messages
func (tbot *TelegramBot) SendMessage(message string, markdown bool) error {
	tbot.SetRequestURL(cmdSendMes, ("chat_id=" + tbot.ChatID + "&text=" + message), markdown)

	//err := tbot.MakeRequest()
	err := tbot.FastRequest()
	if err != nil {
		return err
	}
	return nil
}

// TelegramBotHandler reads bot params from configPath and accepts some log struct to find if its needed to print some mess to telegram bot
func (tbot *TelegramBot) Write(message []byte) error {
	err := tbot.SendMessage(string(message), false)
	if err != nil {
		return err
	}
	return nil
}

// FastRequest make fasthttp
func (tbot *TelegramBot) FastRequest() error {
	for {

		err := tbot.FastHTTPClient.DoTimeout(tbot.Request, tbot.Response, time.Minute)
		switch err {
		case fasthttp.ErrTimeout, fasthttp.ErrDialTimeout:
			<-time.After(time.Minute * 2)
			continue
		case fasthttp.ErrNoFreeConns:
			<-time.After(time.Minute * 2)
			continue
		case nil:
			return nil
		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				<-time.After(time.Minute * 2)
				continue
			} else {
				return err
			}
		}
	}
}
