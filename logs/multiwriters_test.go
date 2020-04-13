package logs

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

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
	return len(b), ErrBadWriter
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
