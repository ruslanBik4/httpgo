// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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

	tb, err := NewTelegramBotFromEnv()
	as.EqualError(err, "Bad TelegramBot parameters, StatusCode: 404 Description: Not Found")

	if tb != nil {
		as.Equal("bottoken", tb.Token, "Token from env wrong")
		as.Equal("chatid", tb.ChatID, "ChatID from env wrong")
		err, _ = tb.SendMessage("some mess", false)
		as.EqualError(err, BadTelegramBot.Error())

		_, err = tb.Write([]byte("some mess"))
		as.Nil(err, "%v", err)

		tb = &TelegramBot{}

		_, err = tb.Write([]byte("some mess"))
		if as.NotNil(err) {
			as.EqualError(err, "TelegramBot.Token empty")
		}
	}
}

//server from fasthttp example
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

func TestErrorLogTelegramWritesSecondVersion(t *testing.T) {
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
	wg := &sync.WaitGroup{}
	wg.Add(2)

	l, err := net.Listen("tcp", useTestLocalPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	//defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}

			reader := bufio.NewReader(c)

			netData, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			if strings.Contains(netData, newError.Error()) ||
				strings.Contains(netData, strings.Replace(newErrorWraped.Error(), " ", "%20", -1)) {
				fmt.Println("strings.Contains(netData, Error())")
				wg.Done()
			}
		}
	}()

	//// === check with logs
	logs.SetWriters(tb, logs.FgErr)

	logs.ErrorLog(newError)
	logs.ErrorLog(newErrorWraped)

	wg.Wait()

}
