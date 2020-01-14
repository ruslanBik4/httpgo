// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTelegramBot(t *testing.T) {
	b, err := NewTelegramBot("bot_test_vals.yml")
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}
	assert.NotEmpty(t, b.Token)
	if assert.Equal(t, "bottoken", b.Token) && assert.Equal(t,"chatid", b.ChatID) {return}

	err = b.GetUpdates()
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	t.Log("updates", b.props)

	err = b.GetChat(b.ChatID)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}
	t.Log(b.props)

	err = b.GetChatMemberCount(b.ChatID)
	if !assert.Nil(t, err) {
		t.Fatal(err)
		fmt.Println()
	}

	c := b.props["result"].(float64)
	t.Log(fmt.Sprintf("try to sum users of group - %.0f", c))
	// for i := ; i < c; i++ {
	b.GetChatMember(b.ChatID, "91653754")
	// b.SendMessage(fmt.Sprintf("user - %v", b.props["result"]), true)
	// }
	// b.SendMessage(fmt.Sprintf("try to sum users of group - %f", c), true)
	t.Log(b.Response.String())
}
