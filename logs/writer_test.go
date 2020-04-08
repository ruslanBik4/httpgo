// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

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

	ErrorLog(err, "test err", 2)
	ErrorLog(err, "test err %s format %d", 2)
	ErrorLog(err, "test err format missing %d", 2, time.Now())

	err1 := errors.Wrap(err, "mess for error")
	ErrorLog(err1, "test err wrap", 1)
	ErrorLog(err1, "test err wrap format %d", 1)
	ErrorLog(err1, "test err wrap format missing %d", 1, 3)
}

type fakeWriter struct {
	wg *sync.WaitGroup
}

func newFakeWriter(wg *sync.WaitGroup) *fakeWriter {
	wg.Add(1)

	return &fakeWriter{wg: wg}
}

func (w fakeWriter) Write(b []byte) (int, error) {

	fmt.Println(boldcolors[WARNING] + "fake writer" + string(b) + LogEndColor)

	w.wg.Done()

	return len(b), nil
}

func TestErrorLogOthers(t *testing.T) {
	runtime.GOMAXPROCS(3)
	t.Run("testErrtoOthers", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		fwriter := fakeWriter{wg}
		mw := logErr.toOther.(*multiWriter)

		SetWriters(fwriter, FgErr)
		SetWriters(fwriter, FgErr)

		var err fakeErr

		ErrorLog(err, "test err", 1)

		wg.Wait()

		wg.Add(1)
		DeleteWriters(fwriter, FgErr)
		if len(mw.writers) == 0 {
			wg.Done()
		}

		wg.Wait()
	})

}

// Func to check error occurance line
func FuncStack(number int) error {
	err := InnerErrorFunc(number)
	if number < 5 && err != nil {
		err = errors.New("number < 5 and InnerErrorFunc() error occured")
		ErrorLog(err)
		return err
	}
	if number >= 5 && err != nil {
		return err
	}
	return err
}

// Func to raise error inside
func InnerErrorFunc(number int) error {
	err := InnerErrorFuncLower(number)
	if number < 5 && err != nil {
		ErrorLog(err)
		return err
	}

	if number < 5 {
		return nil
	}
	err = errors.New("-InnerErrorFunc()- error occured")
	ErrorLog(err)
	return err
}

func InnerErrorFuncLower(number int) error {
	if number < 3 {
		return nil
	}
	err := errors.New("-InnerErrorFuncLower()- error occured")
	ErrorLog(err)
	ErrorStack(err)
	return err
}

func TestLogErr(t *testing.T) {
	err := FuncStack(7)
	ErrorLog(errors.Wrap(err, "uhd3ekuiwe"))
	err = FuncStack(3)
	ErrorLog(errors.Wrap(err, "yw"))
}

func TestErrStack(t *testing.T) {
	err := FuncStack(7)
	ErrorStack(errors.Wrap(err, "uhd3ekuiwe"))
	err = FuncStack(3)
	ErrorStack(errors.Wrap(err, "yw"), "khef")
}

func RemoveTest(mw *multiWriter, w io.Writer, wg *sync.WaitGroup) {
	old_len := len(mw.writers)
	mw.Remove(w)
	if old_len == len(mw.writers)+1 {
		wg.Done()
	}
}

func TestRemoveOneMultiwriter(t *testing.T) {
	runtime.GOMAXPROCS(2)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	b := testWriter2{wg, "Second"}
	m := MultiWriter(b)

	mw := m.(*multiWriter)

	RemoveTest(mw, b, wg)

	data := []byte("Hello ")
	_, e := m.Write(data)
	if e != nil {
		panic(e)
	}
	time.Sleep(10 * time.Millisecond)
	wg.Wait()
}

func TestLogsMultiwriter(t *testing.T) {
	runtime.GOMAXPROCS(3)
	wg := &sync.WaitGroup{}

	wg.Add(2)
	a := testWriter{wg, "first"}
	b := testWriter2{wg, "Second"}
	m := MultiWriter(a, a)
	m = MultiWriter(m, a)

	mw := m.(*multiWriter)

	mw.Append(a, b)
	mw.Remove(a)
	mw.Append(b)

	data := []byte("Hello ")
	_, e := m.Write(data)
	if e != nil {
		panic(e)
	}
	wg.Wait()
}

func TestErrorsMultiwriter(t *testing.T) {
	a := testErrorWriter{}
	b := testBadWriter{}
	m := MultiWriter(a, a)
	m = MultiWriter(m, a)
	m = MultiWriter(m, b)

	for i := 0; i < 2; i++ {
		_, e := m.Write([]byte("Hello "))
		if errMultiwriter, ok := e.(MultiwriterErr); ok {
			assert.Equal(t, 3, len(errMultiwriter.ErrorsList))
			for _, writerErr := range errMultiwriter.ErrorsList {
				fmt.Printf("error: %v, writer:%v\n", writerErr.Err, writerErr.Wr)
			}
		} else if e != nil {
			fmt.Println("error: ", e)
		}
	}

	assert.Equal(t, 3, len(m.(*multiWriter).writers))
}

type testErrorWriter struct{}

func (tew testErrorWriter) Write(b []byte) (int, error) {
	fmt.Println("testErrorWriter: ", string(b))
	return len(b), errors.New("TestErrorWrite Error occured")
}

type testBadWriter struct{}

func (tbw testBadWriter) Write(b []byte) (int, error) {
	fmt.Println("testBadWriter: ", string(b))
	return len(b), BadWriter
}

type testWriter struct {
	wg   *sync.WaitGroup
	name string
}

func (tw testWriter) Write(b []byte) (int, error) {
	fmt.Println(boldcolors[WARNING] + tw.name + "|| testWriter writer: " + string(b) + LogEndColor)
	return len(b), nil
}

type testWriter2 struct {
	wg   *sync.WaitGroup
	name string
}

func (tw2 testWriter2) Write(b []byte) (int, error) {
	fmt.Println(boldcolors[DEBUG] + tw2.name + "|| testWriter2 writer: " + string(b) + LogEndColor)
	tw2.wg.Done()
	return len(b), nil
}

func TestLogsWithSentry(t *testing.T) {
	err := SetSentry("https://5gerstge5rgtry.io/18wstger4tge5rg13325", "ertgesrg")
	fmt.Println(err)
	ErrorLog(errors.New("Test SetSentry"))

	err = fakeErr{}
	ErrorLog(err)
}

func BenchmarkErrorLog(b *testing.B) {

	b.Run("single log", func(b *testing.B) {
		ErrorLog(errors.New("BenchmarkErrorLog testing"), 1)
		b.ReportAllocs()
	})

	wg := &sync.WaitGroup{}
	b.Run("fakewriter", func(b *testing.B) {

		SetWriters(newFakeWriter(wg), FgErr)
		ErrorLog(errors.New("BenchmarkErrorLog testing with "), 1)

		b.ReportAllocs()
	})
	wg.Wait()

}

func TestDebug(t *testing.T) {
	DebugLog("deiwhd", "sduh")
}
