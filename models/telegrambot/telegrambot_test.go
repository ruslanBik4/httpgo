// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var useTestLocalPort string = ":21476"

func TestISWrite(t *testing.T) {
	b := &TelegramBot{}

	assert.Implements(t, (*io.Writer)(nil), b)
}

func TestSome(t *testing.T) {
	as := assert.New(t)
	as.Containsf("true false dunno", "true", "dunn")
}

func TestNewTelegramBotFromEnv(t *testing.T) {
	as := assert.New(t)

	err := os.Setenv("TBTOKEN", "bottoken")
	as.Nil(err, "%v", err)
	err = os.Setenv("TBCHATID", "chatid")
	as.Nil(err, "%v", err)

	fmt.Println(ErrBadTelegramBot)

	e := errors.Wrapf(
		ErrBadTelegramBot,
		"StatusCode: %d Description: %s",
		404,
		"Not Found")

	fmt.Println(e)

	tb, err := NewTelegramBotFromEnv()
	fmt.Println(err)
	as.EqualError(err, "Bad TelegramBot parameters, StatusCode: 404 Description: Not Found")

	if tb != nil {
		as.Equal("bottoken", tb.Token, "Token from env wrong")
		as.Equal("chatid", tb.ChatID, "ChatID from env wrong")
		err, _ = tb.SendMessage("some mess", false)
		as.EqualError(err, ErrBadTelegramBot.Error())

		_, err = tb.Write([]byte("some mess"))
		as.Nil(err, "%v", err)

		tb = &TelegramBot{}

		_, err = tb.Write([]byte("some mess"))
		if as.NotNil(err) {
			as.EqualError(err, "TelegramBot.Token empty")
		}
	}
}

// server from fasthttp example
func FastHTTPServer() {
	ln, err := net.Listen("tcp", useTestLocalPort)
	if err != nil {
		log.Fatalf("error in net.Listen: %s", err)
	}
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		fmt.Println(ctx, "Requested path is")
	}
	if err := fasthttp.Serve(ln, requestHandler); err != nil {
		log.Fatalf("error in Serve: %s", err)
	}
}

func TestErrorLogTelegramWrite(t *testing.T) {
	as := assert.New(t)

	err := os.Setenv("TBTOKEN", "bottoken")
	as.Nil(err, "%v", err)
	err = os.Setenv("TBCHATID", "chatid")
	as.Nil(err, "%v", err)

	tbtoken := os.Getenv("TBTOKEN")
	tbchatid := os.Getenv("TBCHATID")

	//tb, err := NewTelegramBotFromEnv()
	tb := &TelegramBot{
		Token:          tbtoken,
		ChatID:         tbchatid,
		Response:       &fasthttp.Response{},
		RequestURL:     "http://localhost" + useTestLocalPort + "/",
		Request:        &fasthttp.Request{},
		FastHTTPClient: &fasthttp.Client{},
	}
	tb.Request.Header.SetMethod(fasthttp.MethodPost)
	tb.allocStack()

	as.Equal("bottoken", tb.Token, "Token from env wrong")
	as.Equal("chatid", tb.ChatID, "ChatID from env wrong")

	newError := errors.New("NewERROR")
	newErrorWraped := errors.Wrap(newError, "Wraped")

	paramNames := []string{"chat_id", "text"}

	ch := make(chan string, 2)

	// ===== Simple net.http/ListenAndServe server with specific handler to read out telegrambot request =====
	http.HandleFunc("/bottoken/sendMessage",
		func(w http.ResponseWriter, r *http.Request) {
			t.Log("HandleFunc r.URL", r.URL)

			switch r.Method {
			case "GET":
				for i, paramName := range paramNames {
					params, ok := r.URL.Query()[paramName]

					if !ok || len(params[0]) < 1 {

						resp := fmt.Sprintf("Url Param %s is missing", paramName)
						w.WriteHeader(http.StatusBadRequest)
						_, err = w.Write([]byte(resp))
						if err != nil {
							t.Error(err)
						}
						ch <- resp
						return
					}

					msg := params[0]
					if i == 0 {
						as.Equal(msg, tb.ChatID, "ChatID in request is wrong")
					} else if strings.Contains(msg, strings.Replace(newError.Error(), " ", "%20", -1)) ||
						strings.Contains(msg, strings.Replace(newErrorWraped.Error(), " ", "%20", -1)) {
						ch <- "ok"
						return

					}
				}
			case "POST":
				chatID := r.FormValue("chat_id")
				msg := r.FormValue("text")
				if as.Equal(tb.ChatID, chatID, "ChatID in request is wrong") &&
					(strings.Contains(msg, newError.Error()) ||
						strings.Contains(msg, newErrorWraped.Error())) {
					w.WriteHeader(http.StatusOK)

					ch <- "ok"
					return
				}
			default:
				ch <- "unknown method"
			}
			w.WriteHeader(http.StatusNotFound)
			ch <- "not found"
		})

	go func() {
		err := http.ListenAndServe(useTestLocalPort, nil)
		if err != nil {
			t.Fatal(err)
		}
	}()
	// =========================================

	//// another server version, but hadler hasn't written for the errors and waitgroup
	//go FastHTTPServer()

	go func() {
		sendMsg(tb, newError, as, newErrorWraped)
		close(ch)
	}()

	for str := range ch {
		switch str {
		case "ok":
			t.Log(str)
		default:
			t.Fatal(str)
		}
	}

}

func sendMsg(tb *TelegramBot, newError error, as *assert.Assertions, newErrorWrapped error) {
	_, err := tb.Write([]byte(newError.Error()))
	as.Nil(err, "error writing tb.Write([]byte(newError.Error()))")

	_, err = tb.Write([]byte(newErrorWrapped.Error()))
	as.Nil(err, "error writing tb.Write([]byte(newErrorWrapped.Error()))")

	_, err = tb.Write([]byte(""))
	as.Nil(err, "error nil if message empty")
}

const (
	botToken = "bottoken"
	chatId   = "chatid"
)

var (
	newError        = errors.New("NewERROR")
	newErrorWrapped = errors.Wrap(newError, "Wrapped")
	longMess        = "begin" + strings.Repeat("o", maxMessLength-8) + "end"
)

func TestErrorLogTelegramWritesSecondVersion(t *testing.T) {
	as := assert.New(t)

	err := os.Setenv("TBTOKEN", botToken)
	as.Nil(err, "%v", err)
	err = os.Setenv("TBCHATID", chatId)
	as.Nil(err, "%v", err)

	tbtoken := os.Getenv("TBTOKEN")
	tbchatid := os.Getenv("TBCHATID")

	ch := make(chan struct{})

	p, err := mockTelegramServer(t, ch)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// tb, err := NewTelegramBotFromEnv()
	tb := newTestBot(tbtoken, tbchatid, p)

	as.Equal(botToken, tb.Token, "Token from env wrong")
	as.Equal(chatId, tb.ChatID, "ChatID from env wrong")

	// todo: remove logs methods!
	_, err = tb.Write([]byte(newError.Error()))
	as.Nil(err, "%v", err)
	<-ch
	//select {
	//case <-ch:
	//case <-time.After(time.Second*10):
	//	t.Error("timeout")
	//}

	_, err = tb.Write([]byte(newErrorWrapped.Error()))
	as.Nil(err, "%v", err)
	<-ch
	//select {
	//case <-ch:
	//case <-time.After(time.Second*10):
	//	t.Error("timeout")
	//}

}

func TestTelegramBot_SendMessage(t *testing.T) {
	ch := make(chan struct{})

	p, err := mockTelegramServer(t, ch)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	tb := newTestBot(botToken, chatId, p)

	go func() {
		err, resp := tb.SendMessage(longMess, true)

		assert.Nil(t, err, "error must be nil")
		t.Log(resp)
		close(ch)
	}()

	for isRun := true; isRun; {
		select {
		case _, ok := <-ch:
			t.Log("request finished")
			if !ok {
				isRun = false
			}
		case <-time.After(time.Second * 10):
			t.Log("timeout")
			isRun = false
		}
	}
}

func TestTelegramBot_SendEmptyMessage(t *testing.T) {
	ch := make(chan struct{})

	p, err := mockTelegramServer(t, ch)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	tb := newTestBot(botToken, chatId, p)

	err, resp := tb.SendMessage("", true)

	assert.Nil(t, err, "error must be nil")
	t.Log("response", resp)

}

func TestTelegramBot_getPartMes(t *testing.T) {
	tb := newTestBot(botToken, chatId, "")
	prefix := " part 1 "
	r := strings.NewReader(longMess)

	num := tb.messId

	for i := 1; r.Len() > 0; i++ {
		mes, err := tb.getPartMes(r, prefix, num+1)
		assert.Nil(t, err)
		t.Log(len(mes), strings.Replace(mes, "o", "", -1))
		prefix = fmt.Sprintf(" MESS #%v part %d ", tb.messId, i+1)
	}
}

func newTestBot(tbtoken string, tbchatid string, port string) *TelegramBot {
	tb := &TelegramBot{
		Token:          tbtoken,
		ChatID:         tbchatid,
		Response:       &fasthttp.Response{},
		RequestURL:     "http://localhost" + port + "/",
		Request:        &fasthttp.Request{},
		FastHTTPClient: &fasthttp.Client{},
		instance:       "test",
	}
	tb.Request.Header.SetMethod(fasthttp.MethodPost)

	return tb
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func mockTelegramServer(t *testing.T, ch chan struct{}) (string, error) {
	p, err := GetFreePort()
	if err != nil {
		return "", err
	}

	s := ":" + strconv.Itoa(p)
	l, err := net.Listen("tcp", s)
	if err != nil {
		return "", err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				t.Error(err)
				return
			}
			t.Log("con begin")
			go mockHandling(t, ch, c)
		}
	}()

	return s, nil
}

const (
	headStatus = `HTTP/1.1 200 OK`
	headBody   = `
Host: localhost;
Content-Type: application/json;
Content-Length:12;

`
)

var (
	eProto        = []byte("POST /bottoken/sendMessage HTTP/1.1")
	cntType       = []byte("Content-Type")
	cntLen        = []byte("Content-Length:")
	mpType        = []byte("multipart/form-data")
	mpStart       = []byte("--")
	regForm       = regexp.MustCompile(`Content-Disposition:\s*form-data;\s*name\="text"`)
	regFormChatId = regexp.MustCompile(`Content-Disposition:\s*form-data;\s*name\="chat_id"`)
	regNewError   = regexp.MustCompile(`(test)*(.*?)(NewERROR)*?`)
)

func mockHandling(t *testing.T, ch chan struct{}, conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Error(err)
			return
		}

		ch <- struct{}{}
	}()

	r := bufio.NewReader(conn)
	proto, _, err := r.ReadLine()
	if err != nil {
		t.Error(err)
		return
	}

	text, chat, isChatId, l := "", "", 0, 0
	assert.Equal(t, eProto, proto, "proto is wrong")

	for isRun, mode := true, 0; isRun; {

		switch line, isPrefix, err := r.ReadLine(); {
		case err != nil:
			t.Error(err)
			isRun = false
		case mode == 1:
			mode++
		case mode == 2:
			if !assert.True(t, regNewError.Match(line), "wrong error message") {
				t.Log(string(line))
			}
			text = string(line)
			mode = 0
			// isRun = false
		case bytes.HasPrefix(line, cntType):
			assert.True(t, bytes.Contains(line, mpType), " Content-Type is wrong")
		case bytes.HasPrefix(line, cntLen):
			l, err = strconv.Atoi(string(bytes.Trim(bytes.TrimPrefix(line, cntLen), " ")))
			if err != nil {
				t.Errorf("%v %s", err, line)
			} else {
				t.Log(l)
			}
		case bytes.HasPrefix(line, mpStart) && bytes.HasSuffix(line, mpStart):
			t.Log(string(line))
			isRun = false
			b, err := r.Peek(r.Buffered())
			if err != nil {
				t.Errorf("%v %s", err, line)
			} else if len(b) < l {
				t.Logf("read only %d from %d", len(b), l)
			}
		case regForm.Match(line):
			mode++
		case regFormChatId.Match(line):
			isChatId++
		case isPrefix:
			t.Log("line too long", string(line))
		default:
			if isChatId == 2 {
				chat = string(line)
			}
			if isChatId >= 1 {
				isChatId++
			}
			t.Log(string(line))
		}
	}

	resp := ""
	if len(text) == 0 {
		resp = `HTTP/1.1 400 Bad Request` + headBody + `{"ok":false,"error_code":400,"description":"Bad Request: message text is empty"}`
	} else if len(text) > 4050 {
		resp = `HTTP/1.1 400 Bad Request` + headBody + `{"ok":false,"error_code":400,"description":"Bad Request: message is too long"}`
	} else {
		resp = headStatus + headBody + `{"ok":true,"result":{"message_id":324,"chat":{"id":"` + chat + `","title":"","username":"","type":"channel"},"date":` + strconv.FormatInt(time.Now().Unix(), 10) + `,"text":"` + text + `"}}`
	}

	w := bufio.NewWriter(conn)
	_, err = w.Write([]byte(resp))
	if err != nil {
		t.Error(err)
	}

	err = w.Flush()
	if err != nil {
		t.Error(err)
	}

}
