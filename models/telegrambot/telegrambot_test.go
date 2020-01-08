// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"io"
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

	as.Nil(err, "%v", err)
	as.Equal("bottoken", tb.Token, "Token from env wrong") 
	as.Equal("chatid", tb.ChatID, "ChatID from env wrong")
	

	err = tb.SendMessage("some mess", false)
	as.Nil(err, "%v", err)

	_, err = tb.Write([]byte("some mess"))
	as.Nil(err, "%v", err)

	tb = &TelegramBot{}

	_, err = tb.Write([]byte("some mess"))
	if as.NotNil(err) {
		as.EqualError(err, "TelegramBot.Token empty")
	}
}
