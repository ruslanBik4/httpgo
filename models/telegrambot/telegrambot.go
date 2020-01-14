package telegrambot

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
	"github.com/acarl005/stripansi"

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

	props map[string]interface{}
}

// NewTelegramBot reads a config file for bot token and chatID and creates new TelegramBot struct
func NewTelegramBot(confPath string) (tb *TelegramBot, err error) {

	yamlFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &tb)
	if err != nil {
		logs.DebugLog(" %+v", string(yamlFile))
		return nil, err
	}

	tb.RequestURL = baseURL
	tb.Request = &fasthttp.Request{}
	tb.Response = &fasthttp.Response{}
	tb.FastHTTPClient = &fasthttp.Client{}

	return
}

// NewTelegramBot is a constructor from ENV
func NewTelegramBotFromEnv() (tb *TelegramBot, err error) {
	if os.Getenv("TBTOKEN") == "" || os.Getenv("TBCHATID") == "" {
		return nil, errors.New("Empty environment variables (TBTOKEN or TBCHATID) for TelegramBot creation.")
	}

	return &TelegramBot{
		Token:          os.Getenv("TBTOKEN"),
		ChatID:         os.Getenv("TBCHATID"),
		Response:       &fasthttp.Response{},
		RequestURL:     baseURL,
		Request:        &fasthttp.Request{},
		FastHTTPClient: &fasthttp.Client{},
	}, nil

}

// SetRequestURL makes url for request
func (tbot *TelegramBot) SetRequestURL(action string, otherRequest string, markdown bool) {
	newUrl := (tbot.RequestURL + tbot.Token + "/" + action + "?" + otherRequest)
	if markdown {
		newUrl += "&parse_mode=Markdown"
	}

	tbot.Request.SetRequestURI(newUrl)

}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetUpdates() error {
	tbot.SetRequestURL(cmdgetUpdates, "", true)

	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

func (tbot *TelegramBot) readResponse(err error) error {
	d := tbot.Response.Body()

	enc := jsoniter.NewDecoder(bytes.NewReader(d))

	err = enc.Decode(&tbot.props)
	if err != nil {
		return err
	}

	return nil
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetChat(name string) error {
	tbot.SetRequestURL(cmdGetChat, "chat_id="+name, true)

	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetChatMemberCount(name string) error {
	tbot.SetRequestURL(cmdGetChMbrsCount, "chat_id="+name, true)

	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetChatMember(name string, user string) error {
	tbot.SetRequestURL(cmdGetChMbr, "chat_id="+name+"&user_id="+user, true)

	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) InviteUser(name string) error {
	tbot.SetRequestURL(cmdInlineMThd, "chat_id="+name, true)

	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return nil
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) SendMessage(message string, markdown bool) error {
	tbot.SetRequestURL(cmdSendMes, ("chat_id=" + tbot.ChatID + "&text=" + stripansi.Strip(strings.Replace(message, " ", "%20", -1))), markdown)
	
	err := tbot.FastRequest()
	if err != nil {
		return err
	}

	return nil
}

// TelegramBotHandler reads bot params from configPath and accepts some log struct to find if its needed to print some mess to telegram bot
func (tbot *TelegramBot) Write(message []byte) (int, error) {
	if tbot.Token == "" {
		return -1, errors.New("TelegramBot.Token empty")
	}
	if tbot.ChatID == "" {
		return -1, errors.New("TelegramBot.ChatID empty")
	}
	if tbot.FastHTTPClient == nil {
		return -1, errors.New("TelegramBot.FastHTTPClient == nil")
	}
	if tbot.Request == nil {
		return -1, errors.New("TelegramBot.Request == nil")
	}
	if tbot.Response == nil {
		return -1, errors.New("TelegramBot.Response == nil")
	}


	err := tbot.SendMessage(string(message), false)
	if err != nil {
		return -1, err
	}

	return 1, nil
}

// FastRequest make fasthttp request
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
			// todo: сделать анализ ответа
			//logs.DebugLog(" %+v", tbot.Response)
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

//GetResult
func (tbot *TelegramBot) GetResult() interface{} {
	return tbot.props["result"]
}
