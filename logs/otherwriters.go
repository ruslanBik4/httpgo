package logs

import (
	"errors"
	"io"
)

var BadWriter = errors.New("BadWriter, it will delete from multiwriter")

type MultiwriterErr struct {
	ErrorsList []WriterErr
}

func (mwe MultiwriterErr) Error() string {
	return "MultiwriterErr: some writers have errors"
}

type WriterErr struct {
	Err error
	Wr  io.Writer
}

type multiWriter struct {
	writers []io.Writer
}

func (t *multiWriter) Write(p []byte) (n int, err error) {
	errList := []WriterErr{}
	for _, w := range t.writers {
		n, err = w.Write(p)

		if err == BadWriter {
			t.Remove(w)
			ErrorLog(BadWriter, w)
			err = nil
		} else if err != nil {
			errList = append(errList, WriterErr{err, w})
			err = nil
		}

		if n != len(p) {
			errList = append(errList, WriterErr{io.ErrShortWrite, w})
		}
	}

	if len(errList) > 0 {
		return len(p), MultiwriterErr{errList}
	}

	return len(p), nil
}

// Removes all writers that are identical to the writer we need to remove
func (t *multiWriter) Remove(writers ...io.Writer) {
	//todo: если нужно будет, чтоб удалялся только один врайтер, то можно поменять местами циклы
	for i := len(t.writers) - 1; i >= 0; i-- {
		for _, v := range writers {
			if t.writers[i] == v {
				t.writers = append(t.writers[:i], t.writers[i+1:]...)
				break
			}
		}
	}
}

func (t *multiWriter) Append(writers ...io.Writer) {
	t.writers = append(t.writers, writers...)
}

var _ io.StringWriter = (*multiWriter)(nil)

func (t *multiWriter) WriteString(s string) (n int, err error) {
	p := []byte(s)

	for _, w := range t.writers {
		if sw, ok := w.(io.StringWriter); ok {
			n, err = sw.WriteString(s)
		} else {
			n, err = w.Write(p)
		}

		if err != nil {
			ErrorLog(err, w)
			err = nil
		}

		if n != len(s) {
			ErrorLog(io.ErrShortWrite, "while WriteString() to single Writer in multiWriter", w)
		}
	}

	return len(s), nil
}

func MultiWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}

	return &multiWriter{allWriters}
}
