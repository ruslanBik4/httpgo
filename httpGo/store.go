/*
 * Copyright (c) 2024-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"bytes"
	"io"
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/logs"
)

type StoreKey struct {
	Id   uint64
	Name string
}

type Store struct {
	sync.RWMutex
	store map[StoreKey]any
}

func NewStore() *Store {
	return &Store{store: make(map[StoreKey]any)}
}

func (s *Store) Set(ctx *fasthttp.RequestCtx, name string, value any) StoreKey {
	key := StoreKey{
		Id:   ctx.ConnID(),
		Name: name,
	}
	s.Lock()
	s.store[key] = value
	s.Unlock()

	return key
}

func (s *Store) Get(id uint64, name string) any {
	s.RLock()
	defer s.RUnlock()

	return s.store[StoreKey{
		Id:   id,
		Name: name,
	}]
}

func (s *Store) Len() int {
	return len(s.store)
}
func (s *Store) StartSSELog(ctx *fasthttp.RequestCtx, startMsg []byte, fnc func(w io.Writer)) StoreKey {
	l := &LogWriter{buf: bytes.NewBuffer(nil), ch: make(chan []byte, 100)}
	storeKey := s.Set(ctx, logName, l)
	l.ch <- startMsg
	go func() {
		defer close(l.ch)
		logs.SetWriters(l, logs.FgInfo, logs.FgErr)
		defer logs.DeleteWriters(l, logs.FgAll)

		fnc(l)
	}()

	return storeKey
}
