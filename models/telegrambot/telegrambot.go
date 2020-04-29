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
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/httpgo/logs"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

// TelegramBot struct with token and one chatid
type TelegramBot struct {
	Token          string `yaml:"BotToken"`
	ChatID         string `yaml:"ChatID"`
	RequestURL     string
	Request        *fasthttp.Request
	Response       *fasthttp.Response
	FastHTTPClient *fasthttp.Client

	props map[string]interface{}

	messagesStack []tbMessageBuffer
	instance      string
	messId        int64
	lock          sync.RWMutex
}

//struct for telegram reply_markup keyboard
type TelegramKeyboard struct {
	Keyboard        [][]string `json:"keyboard"`
	OneTimeKeyboard bool       `json:"one_time_keyboard"`
	ResizeKeyboard  bool       `json:"resize_keyboard"`
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

//
type tbMessageBuffer struct {
	messageText []byte
	messageTime time.Time
}

func (tbResp *TbResponseMessageStruct) String() string {
	if tbResp.Ok == true {
		return fmt.Sprintf("message request is Ok, ResponseStruct: {Ok: %v, Result: {MessageId: %v, Chat:{Id: %v, Title: %v, Username: %v, Type: %v}, Date: %v, Text: %v }}",
			tbResp.Ok, tbResp.Result.MessageId, tbResp.Result.Chat.Id, tbResp.Result.Chat.Title,
			tbResp.Result.Chat.Username, tbResp.Result.Chat.Type, tbResp.Result.Date, tbResp.Result.Text)
	}
	return fmt.Sprintf("message request is not Ok, ErrorCode:%v, %v", tbResp.ErrorCode, tbResp.Description)
}

func (tb *TelegramBot) String() string {
	return fmt.Sprintf("TelegramBot: {Token: %v, ChatID: %v, RequestURL: %v, Request: %v, Response: %v, FastHTTPClient: %v, instance: %v, messagesStack: %v}",
		tb.Token, tb.ChatID, tb.RequestURL, tb.Request, tb.Response, tb.FastHTTPClient, tb.instance, tb.messagesStack)
}

func (messbuf *tbMessageBuffer) String() string {
	return fmt.Sprintf("{messageText: %v, messageTime: %v}", string(messbuf.messageText), messbuf.messageTime)
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
		instance:       "[[#" + filepath.Base(os.Args[0]) + "]] ",
	}
	tb.Request.Header.SetMethod(fasthttp.MethodPost)

	err, resp := tb.SendMessage("Telegram Bot ready", false)
	if err == ErrBadTelegramBot {
		return nil, errors.Errorf(
			"%s, StatusCode: %d Description: %s",
			ErrBadTelegramBot,
			resp.ErrorCode,
			resp.Description)
	} else if err != nil {
		return nil, err
	}

	return tb, nil

}

// setRequestURL makes url for request
func (tbot *TelegramBot) setRequestURL(action string) {
	newUrl := tbot.RequestURL + tbot.Token + "/" + action
	if string(tbot.Request.Header.Method()) == "GET" {
		newUrl += "?"
	}
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
func (tbot *TelegramBot) SendMessage(message string, markdown bool, keys ...interface{}) (err error, response *TbResponseMessageStruct) {
	if err := tbot.checkBot(); err != nil {
		return err, nil
	}

	if string(tbot.Request.Header.Method()) == "GET" {
		strings.Replace(message, " ", "%20", -1)
	}

	requestParams := map[string]string{
		"chat_id": tbot.ChatID,
		"text":    tbot.instance + stripansi.Strip(strings.TrimSpace(message)),
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

	switch messLen := len(requestParams["text"]); {
	case messLen == 0:
		logs.ErrorStack(errors.Wrap(ErrEmptyMessText, message))
		return nil, nil
	case messLen > 4090:
		oldMess := []byte(requestParams["text"])
		prefix := ""
		endix := fmt.Sprintf(" MESS %v CONTINUE->", tbot.messId)

		for i := 0; i < len(oldMess); i += maxMessLength {
			end := i + maxMessLength
			if end > len(oldMess) {
				end = len(oldMess)
			}
			requestParams["text"] = prefix + string(oldMess[i:end]) + endix
			if len(requestParams["text"]) == 0 {
				logs.ErrorStack(errors.Wrap(ErrEmptyMessText, message), i, end)
				return nil, nil
			}

			err, response = tbot.FastRequest(cmdSendMes, requestParams)
			if err != nil {

				return err, response
			}

			prefix = fmt.Sprintf("MESS %v CONTINUE->", tbot.messId)
			if len(oldMess)-(i+maxMessLength) <= maxMessLength {
				endix = fmt.Sprintf("MESS %v ENDED", tbot.messId)
			}
		}
		return err, response
	default:
		err, response = tbot.FastRequest(cmdSendMes, requestParams)
		return err, response
	}

}

// checks bot params, used in send message, can be used in other methods
func (tbot *TelegramBot) checkBot() error {
	if tbot.Token == "" {
		return ErrBadBotParams{"TelegramBot.Token empty"}
	}
	if tbot.ChatID == "" {
		return ErrBadBotParams{"TelegramBot.ChatID empty"}
	}
	if tbot.FastHTTPClient == nil {
		return ErrBadBotParams{"TelegramBot.FastHTTPClient == nil"}
	}
	if tbot.Request == nil {
		return ErrBadBotParams{"TelegramBot.Request == nil"}
	}
	if tbot.Response == nil {
		return ErrBadBotParams{"TelegramBot.Response == nil"}
	}

	return nil
}

// TelegramBotHandler reads bot params from configPath and accepts some log struct to find if its needed to print some mess to telegram bot
func (tbot *TelegramBot) Write(message []byte) (int, error) {
	if len(tbot.messagesStack) > 0 && len(tbot.messagesStack) < 30 {
		if tbot.messagesStack[len(tbot.messagesStack)-1].messageTime != time.Now().Round(1*time.Second) {
			tbot.messagesStack = []tbMessageBuffer{}
		} else {
			for _, v := range tbot.messagesStack {
				if bytes.Equal(v.messageText, message) {
					return len(message), nil
				}
			}
		}
	} else if len(tbot.messagesStack) >= 30 {
		time.Sleep(1 * time.Second)
		tbot.messagesStack = []tbMessageBuffer{}
	}

	err, _ := tbot.SendMessage(string(message), false)
	if err == ErrBadTelegramBot {
		return len(message), logs.ErrBadWriter
	} else if err != nil {
		return -1, err
	}

	tbot.messagesStack = append(tbot.messagesStack, tbMessageBuffer{
		messageText: message,
		messageTime: time.Now().Round(1 * time.Second)})

	return len(message), nil

}

// FastRequest make fasthttp request
func (tbot *TelegramBot) FastRequest(action string, params map[string]string) (error, *TbResponseMessageStruct) {
	tbot.lock.Lock()
	defer tbot.lock.Unlock()

	tbot.setRequestURL(action)
	err := tbot.setMultipartData(params)
	if err != nil {
		return err, nil
	}
	tryCounter := 0

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
				if strings.Contains(resp.Description, "message text is empty") {
					logs.ErrorStack(errors.Wrap(ErrEmptyMessText, ""), params, resp.ErrorCode, resp.Description)
					return nil, resp
				} else if strings.Contains(resp.Description, "message is too long") {
					logs.ErrorStack(errors.Wrap(ErrTooLongMessText, ""))
					return nil, resp
				}
				logs.DebugLog("tb response 400, ResponseStruct:", resp.ErrorCode, resp.Description)
				return ErrBadTelegramBot, resp
			case 404:
				logs.DebugLog("tb response 404, ResponseStruct:", resp.ErrorCode, resp.Description)
				return ErrBadTelegramBot, resp
			case 429:
				<-time.After(time.Second * 1)
			case 500:
				if tryCounter > 100 {
					return ErrTelegramBotMultiple500, resp
				} else {
					tryCounter += 1
					<-time.After(time.Second * 10)
				}
			default:
				if !resp.Ok {
					// todo: add parsing error response
					logs.DebugLog(resp)
				}

				if action == cmdSendMes {
					tbot.messId += 1
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
