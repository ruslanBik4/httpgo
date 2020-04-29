package logs

import (
	"errors"
	"fmt"
	"io"
)

var ErrBadWriter = errors.New("ErrBadWriter, it will be deleted from multiwriter")

type MultiwriterErr struct {
	ErrorsList []WriterErr
}

func (mwe MultiwriterErr) Error() string {
	return "MultiwriterErr: some writers have errors:\n" + mwe.String()
}

func (mwe MultiwriterErr) String() string {
	endl := ""
	retStr := ""
	for _, writerErr := range mwe.ErrorsList {
		retStr += fmt.Sprintf("%sError during write toOther: %v, writer: %v",
			endl, writerErr.Err, writerErr.Wr)
		endl = "\n"
	}

	return retStr
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

		if err == ErrBadWriter {
			t.Remove(w)
			ErrorLog(ErrBadWriter, w)
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
	for i := len(t.writers) - 1; i >= 0; i-- {
		for _, v := range writers {
			if t.writers[i] == v {
				t.writers = append(t.writers[:i], t.writers[i+1:]...)
				break
			}
		}
	}
}

// Appends each writer passed as single writer entity. If multiwriter is passed, appends it as single writer.
func (t *multiWriter) Append(writers ...io.Writer) {
	t.writers = append(t.writers, writers...)
}

// If multiwriter is passed, appends each writer of multiwriter separately
func (t *multiWriter) AppendWritersSeparately(writers ...io.Writer) {
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			t.writers = append(t.writers, mw.writers...)
		} else {
			t.writers = append(t.writers, w)
		}
	}
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

//creates a multiwriter
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
