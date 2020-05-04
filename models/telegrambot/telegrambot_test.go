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
	"sync"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/logs"

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

	as.Equal("bottoken", tb.Token, "Token from env wrong")
	as.Equal("chatid", tb.ChatID, "ChatID from env wrong")

	newError := errors.New("NewERROR")
	newErrorWraped := errors.Wrap(newError, "Wraped")

	parametrsCheck := []string{"chat_id", "text"}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// ===== Simple net.http/ListenAndServe server with specific handler to read out telgrambot request =====
	http.HandleFunc("/bottoken/sendMessage",
		func(w http.ResponseWriter, r *http.Request) {
			log.Println("HandleFunc r.URL", r.URL)

			if r.Method == "GET" {
				for i, paramName := range parametrsCheck {
					params, ok := r.URL.Query()[paramName]

					if !ok || len(params[0]) < 1 {
						log.Println("Url Param", paramName, "is missing")
						t.Fail()
						wg.Done()
						return
					}

					if i == 0 {
						as.Equal(string(params[0]), tb.ChatID, "ChatID in request is wrong")
					} else {
						if strings.Contains(string(params[0]), strings.Replace(newError.Error(), " ", "%20", -1)) ||
							strings.Contains(string(params[0]), strings.Replace(newErrorWraped.Error(), " ", "%20", -1)) {
							wg.Done()
							return
						}
					}
				}
			}

			if r.Method == "POST" {
				if as.Equal(tb.ChatID, r.FormValue("chat_id"), "ChatID in request is wrong") {
					if strings.Contains(r.FormValue("text"), newError.Error()) ||
						strings.Contains(r.FormValue("text"), newErrorWraped.Error()) {
						wg.Done()
						return
					}
				}
			}
		})

	go http.ListenAndServe(useTestLocalPort, nil)
	// =========================================

	//// another server version, but hadler hasn't written for the errors and waitgroup
	//go FastHTTPServer()

	_, err = tb.Write([]byte(newError.Error()))
	as.Nil(err, "error writing tb.Write([]byte(newError.Error()))")

	_, err = tb.Write([]byte(newErrorWraped.Error()))
	as.Nil(err, "error writing tb.Write([]byte(newErrorWraped.Error()))")

	wg.Wait()

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
	logs.SetWriters(tb, logs.FgErr)

	logs.ErrorLog(newError)
	<-ch

	logs.ErrorLog(newErrorWrapped)
	<-ch
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
	}()

	<-ch
	<-ch
}

func TestTelegramBot_getPartMes(t *testing.T) {
	tb := newTestBot(botToken, chatId, "")
	prefix := " part 1 "
	r := strings.NewReader(longMess)

	for i := 1; r.Len() > 0; i++ {
		mes, err := tb.getPartMes(r, prefix)
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

			go mockHandling(t, ch, c)
		}
	}()

	return s, nil
}

const (
	head = `HTTP/1.1 200 OK 
Host: localhost;
Content-Type: application/json;
Content-Length:12;

`
)

var (
	eProto      = []byte("POST /bottoken/sendMessage HTTP/1.1")
	cntType     = []byte("Content-Type")
	cntLen      = []byte("Content-Length:")
	mpType      = []byte("multipart/form-data")
	mpStart     = []byte("--")
	regForm     = regexp.MustCompile(`Content-Disposition:\s*form-data;\s*name\="text"`)
	regNewError = regexp.MustCompile(`test\[\[ERROR\]\]\s*\d+:\d+:\d+\s+telegrambot_test.go:\d+:\s+init\(\)([\w\s]+:)?\s+NewERROR`)
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

	assert.Equal(t, eProto, proto, "proto is wrong")

	for isRun, mode := true, 0; isRun; {

		switch line, isPrefix, err := r.ReadLine(); {
		case err != nil:
			t.Error(err)
			isRun = false
		case mode == 1:
			mode++
		case mode == 2:
			assert.True(t, regNewError.Match(line), "wrong error message")
			isRun = false
		case bytes.HasPrefix(line, cntType):
			assert.True(t, bytes.Contains(line, mpType), " Content-Type is wrong")
		case bytes.HasPrefix(line, cntLen):
		case bytes.HasPrefix(line, mpStart):
		case regForm.Match(line):
			mode++
		case isPrefix:
			t.Log("line too long", string(line))
		default:
			t.Log(string(line))
		}
	}

	// todo: implements struct fill & marshal into response
	// resp := &TbResponseMessageStruct{
	// 	Ok: true,
	// }

	w := bufio.NewWriter(conn)
	_, err = w.Write([]byte(head + `{"ok": true}`))
	if err != nil {
		t.Error(err)
	}

	err = w.Flush()
	if err != nil {
		t.Error(err)
	}
}
