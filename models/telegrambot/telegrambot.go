package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"

	"github.com/ruslanBik4/httpgo/logs"
)

var BadTelegramBot = errors.New("Bad TelegramBot parameters")

// TelegramBot struct with token and one chatid
type TelegramBot struct {
	Token          string `yaml:"BotToken"`
	ChatID         string `yaml:"ChatID"`
	RequestURL     string
	Request        *fasthttp.Request
	Response       *fasthttp.Response
	FastHTTPClient *fasthttp.Client

	props map[string]interface{}

	messagesStack struct {
		messageText string
		messageTime time.Time
	}
}

//struct for telegram reply_markup keyboard
type TelegramKeyboard struct {
	Keyboard        [][]string `json:"keyboard"`
	OneTimeKeyboard bool       `json:"one_time_keyboard"`
	ResizeKeyboard  bool       `json:"resize_keyboard"`
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
	errmess := ""
	tbtoken := os.Getenv("TBTOKEN")
	tbchatid := os.Getenv("TBCHATID")
	if tbtoken == "" {
		errmess += "TBTOKEN "
	}

	if tbchatid == "" {
		errmess += "TBCHATID "
	}

	if errmess != "" {
		return nil, errors.New("Empty environment variables: " + errmess + "for TelegramBot creation.")
	}

	tb = &TelegramBot{
		Token:          tbtoken,
		ChatID:         tbchatid,
		Response:       &fasthttp.Response{},
		RequestURL:     baseURL,
		Request:        &fasthttp.Request{},
		FastHTTPClient: &fasthttp.Client{},
	}
	tb.Request.Header.SetMethod(fasthttp.MethodPost)

	err, resp := tb.SendMessage("Telegram Bot ready for "+filepath.Base(os.Args[0]), false)
	if err == BadTelegramBot {
		return nil, errors.Wrapf(
			BadTelegramBot,
			"StatusCode: %d Description: %s",
			resp.ErrorCode,
			resp.Description)
	} else if err != nil {
		return nil, err
	}

	return tb, nil

}

// setRequestURL makes url for request
func (tbot *TelegramBot) setRequestURL(action string) {
	newUrl := (tbot.RequestURL + tbot.Token + "/" + action + "?")
	tbot.Request.SetRequestURI(newUrl)
}

// Set multipart data for request
func (tbot *TelegramBot) setMultipartData(params map[string]string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for name, val := range params {
		err := w.WriteField(name, val)
		if err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	tbot.Request.Header.Set("Content-Type", w.FormDataContentType())
	tbot.Request.SetBody(b.Bytes())
	return nil
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetUpdates() error {
	err, _ := tbot.FastRequest(cmdgetUpdates, map[string]string{})
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
	err, _ := tbot.FastRequest(cmdGetChat,
		map[string]string{
			"chat_id": name,
		})
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetChatMemberCount(name string) error {
	err, _ := tbot.FastRequest(cmdGetChMbrsCount,
		map[string]string{
			"chat_id": name,
		})
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) GetChatMember(name string, user string) error {
	err, _ := tbot.FastRequest(cmdGetChMbr,
		map[string]string{
			"chat_id": name,
			"user_id": user,
		})
	if err != nil {
		return err
	}

	return tbot.readResponse(err)
}

// SendMessage is used for sending messages
func (tbot *TelegramBot) InviteUser(name string) error {
	err, _ := tbot.FastRequest(cmdInlineMThd,
		map[string]string{
			"chat_id": name,
		})
	if err != nil {
		return err
	}

	return nil
}

// SendMessage is used for sending messages. Arguments keys must contain TelegramKeyboard{} to add keys to your message
func (tbot *TelegramBot) SendMessage(message string, markdown bool, keys ...interface{}) (error, *TbResponseMessageStruct) {
	if string(tbot.Request.Header.Method()) == "GET" {
		strings.Replace(message, " ", "%20", -1)
	}

	requestParams := map[string]string{
		"chat_id": tbot.ChatID,
		"text":    message,
	}
	if markdown {
		requestParams["parse_mode"] = "Markdown"
	}

	if keys != nil {
		replyMarkup, checkType := keys[0].(TelegramKeyboard)

		if checkType == true {
			keysJsonString, err := json.Marshal(replyMarkup)
			if err != nil {
				fmt.Println(err)
			} else {
				requestParams["reply_markup"] = string(keysJsonString)
			}
		}
	}

	err, response := tbot.FastRequest(cmdSendMes, requestParams)

	if err != nil {
		return err, response
	}

	return nil, response
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

	err, _ := tbot.SendMessage(string(message), false)
	if err != nil {
		if err == BadTelegramBot {
			return len(message), logs.BadWriter
		}
		return -1, err
	}

	return len(message), nil
}

// json struct to parse response
type TbResponseMessageStruct struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	Result      struct {
		MessageId int `json:"message_id"`
		Chat      struct {
			Id       int64  `json:"id"`
			Title    string `json:"title"`
			Username string `json:"username"`
			Type     string `json:"type"`
		} `json:"chat"`
		Date int64  `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}

func (tbResp *TbResponseMessageStruct) String() string {
	if tbResp.Ok == true {
		return fmt.Sprintf("message request is Ok, message_id:%v; chat:%v; date:%v; text:%v",
			tbResp.Result.MessageId, tbResp.Result.Chat, tbResp.Result.Date, tbResp.Result.Text)
	}
	return fmt.Sprintf("message request is not Ok, ErrorCode:%v, %v", tbResp.ErrorCode, tbResp.Description)
}

// FastRequest make fasthttp request
func (tbot *TelegramBot) FastRequest(action string, params map[string]string) (error, *TbResponseMessageStruct) {
	tbot.setRequestURL(action)
	err := tbot.setMultipartData(params)
	if err != nil {
		return err, nil
	}

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
			var resp = &TbResponseMessageStruct{}
			err = json.Unmarshal(tbot.Response.Body(), resp)

			switch tbot.Response.StatusCode() {
			case 400:
				return BadTelegramBot, resp
			case 404:
				return BadTelegramBot, resp
			default:
				if !resp.Ok {
					// todo: add parsing error response
					logs.DebugLog(resp)
				}

				return nil, resp
			}

		default:
			if strings.Contains(err.Error(), "connection reset by peer") {
				<-time.After(time.Minute * 2)
				continue
			} else {
				return err, nil
			}
		}
	}
}

//GetResult
func (tbot *TelegramBot) GetResult() interface{} {
	return tbot.props["result"]
}
