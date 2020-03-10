package logs

import (
	"errors"
	"io"
)

var BadWriter = errors.New("BadWriter, needs to be deleted from mutiwriter")

type multiWriter struct {
	writers []io.Writer
}

func (t *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			if err == BadWriter {
				ErrorLog(BadWriter, w)
				t.Remove(w)
				err = nil
			} else {
				return
			}
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
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
	var p []byte
	for _, w := range t.writers {
		if sw, ok := w.(io.StringWriter); ok {
			n, err = sw.WriteString(s)
		} else {
			if p == nil {
				p = []byte(s)
			}
			n, err = w.Write(p)
		}
		if err != nil {
			return
		}
		if n != len(s) {
			err = io.ErrShortWrite
			return
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
