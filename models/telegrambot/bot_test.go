// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTelegramBot(t *testing.T) {
	b, err := NewTelegramBot("cfg.yml")
	assert.Nil(t, err)
	assert.NotNil(t, b)
}
