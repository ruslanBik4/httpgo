// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

func TestWrapKitLogger_Printf(t *testing.T) {
	var flag errLogPrint = true

	l := NewWrapKitLogger("[[test]]", 0)

	l.Printf(flag, "test true flag", 0, "\n")

	flag = false

	l.Printf(flag, "test false flag", 0)

	l.Printf("test not flag", 0)
}

type fakeErr struct{}

func (f fakeErr) Error() string {
	return "fake error"
}

func TestErrorLog(t *testing.T) {

	var err fakeErr

	ErrorLog(err, "test err", 1)

	err1 := errors.Wrap(err, "mess for error")
	ErrorLog(err1, "test err wrap", 1)
}

type fakeWriter struct {
	wg *sync.WaitGroup
}

func (w fakeWriter) Write(b []byte) (int, error) {
	fmt.Println(boldcolors[WARNING]+"fake writer", string(b))

	w.wg.Done()

	return 0, nil
}

func TestErrorLogOthers(t *testing.T) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	SetWriters(fakeWriter{wg}, fgErr)

	var err fakeErr

	ErrorLog(err, "test err", 1)

	wg.Wait()
}