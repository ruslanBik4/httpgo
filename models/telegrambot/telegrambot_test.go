// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"io"
	"testing"
	"log"

	"github.com/stretchr/testify/assert"
)

func TestISWrite(t *testing.T) {
	b := &TelegramBot{}

	assert.Implements(t, (*io.Writer)(nil), b)
}

func TestNilWrite(t *testing.T) {
	tb, err := NewTelegramBotFromEnv()
	log.Println(tb, err)
	code, errr := tb.Write([]byte("try mess"))
	log.Println(code, errr)
}

