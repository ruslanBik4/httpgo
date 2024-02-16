/*
 * Copyright (c) 2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"sync"

	"github.com/valyala/fasthttp"
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
