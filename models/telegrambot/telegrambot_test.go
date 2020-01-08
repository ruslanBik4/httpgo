// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestISWrite(t *testing.T) {
	b := &TelegramBot{}

	assert.Implements(t, (*io.Writer)(nil), b)
}

func TestNewTelegramBotFromEnv(t *testing.T) {
	os.Setenv("TBTOKEN", "bottoken")
	os.Setenv("TBCHATID", "chatid")
	tb, err := NewTelegramBotFromEnv()
	as := assert.New(t)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	if tb.Token != "bottoken" || tb.ChatID != "chatid" {
		log.Println(err)
		t.Fail()
	}

	err = tb.SendMessage("some mess", false)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	_, err = tb.Write([]byte("some mess"))
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	tb = &TelegramBot{}

	_, err = tb.Write([]byte("some mess"))
	if as.NotNil(err) {
		as.EqualError(err, "TelegramBot{} is empty struct")
	}

}
