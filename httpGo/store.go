/*
 * Copyright (c) 2024-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"html"
	"io"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/logs"
)

// how long a finished log stays readable (late clients, reconnects) before it
// is removed from the Store
const sseKeepAfterDone = 10 * time.Minute

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

// SetNew stores value under a fresh random id and returns the key.
//
// Set() uses ctx.ConnID(), which is wrong for anything that must outlive or
// be told apart from a request: a keep-alive connection serves many requests
// with the SAME ConnID (a second job on that connection silently replaced the
// first one's log), and ConnIDs are sequential, so anyone could read any log by
// counting. The id is a random 31-bit number (the sse route reads it as int32),
// re-drawn if it is taken.
func (s *Store) SetNew(name string, value any) (StoreKey, error) {
	s.Lock()
	defer s.Unlock()

	var b [4]byte
	for range 100 {
		if _, err := rand.Read(b[:]); err != nil {
			return StoreKey{}, err
		}
		key := StoreKey{Id: uint64(binary.BigEndian.Uint32(b[:]) & 0x7fffffff), Name: name}
		if _, taken := s.store[key]; key.Id == 0 || taken {
			continue
		}
		s.store[key] = value
		return key, nil
	}
	return StoreKey{}, fmt.Errorf("store: no free id for %q", name)
}

func (s *Store) Get(id uint64, name string) any {
	s.RLock()
	defer s.RUnlock()

	return s.store[StoreKey{
		Id:   id,
		Name: name,
	}]
}

func (s *Store) Delete(key StoreKey) {
	s.Lock()
	delete(s.store, key)
	s.Unlock()
}

func (s *Store) Len() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.store)
}

// StartSSELog (demo): runs fnc in a goroutine and streams to the browser
// EVERYTHING the application logs (info + error levels) meanwhile, plus what
// fnc writes to w. Careful: logs.SetWriters is global, so this also shows log
// lines of other requests - use StartSSEJob for anything user-specific.
func (s *Store) StartSSELog(ctx *fasthttp.RequestCtx, startMsg []byte, fnc func(w io.Writer)) StoreKey {
	return s.startSSE(startMsg, true, func(_ context.Context, w *LogWriter) { fnc(w) })
}

// StartSSEJob runs fnc in a goroutine and streams ONLY what fnc sends to w
// (w.Send for ready HTML, w.Write for raw text/ANSI output). Nothing else of
// the app's logging leaks in. ctx is cancelled when fnc returns.
//
// Return the key to the client as `/httpgo/store/sse?id=<key.Id>`.
func (s *Store) StartSSEJob(startMsg string, fnc func(ctx context.Context, w *LogWriter)) StoreKey {
	return s.startSSE(gotools.StringToBytes(startMsg), false, fnc)
}

func (s *Store) startSSE(startMsg []byte, captureAppLogs bool, fnc func(ctx context.Context, w *LogWriter)) StoreKey {
	l := NewLogWriter()
	l.add(startMsg) // takes ownership of startMsg

	key, err := s.SetNew(logName, l)
	if err != nil {
		logs.ErrorLog(err)
		l.Send("cannot start the log stream: " + html.EscapeString(err.Error()))
		l.Close()
		return key
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// deferred calls run bottom-up
		defer time.AfterFunc(sseKeepAfterDone, func() { s.Delete(key) }) // 5. forget it later
		defer l.Close()                                                  // 4. clients get "closed" after the last line
		defer cancel()                                                   // 3.
		defer func() {                                                   // 2.
			switch err := recover().(type) {
			case nil:
			case error:
				logs.ErrorLog(err)
				l.Send("panic: " + html.EscapeString(err.Error()))
			default:
				logs.StatusLog("recover: %v", err)
				l.Send("panic: " + html.EscapeString(fmt.Sprint(err)))
			}
		}()
		if captureAppLogs {
			logs.SetWriters(l, logs.FgInfo, logs.FgErr)
			defer logs.DeleteWriters(l, logs.FgAll) // 1.
		}

		fnc(ctx, l)
	}()

	return key
}
