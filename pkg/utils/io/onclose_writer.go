package io

import (
	"bytes"
	"io"
)

type LenReadWriter interface {
	io.ReadWriter
	Len() int
}

type OnCloseWriter struct {
	buff    LenReadWriter
	onClose func(r io.Reader, rlen int64) error
}

func (w *OnCloseWriter) Write(in []byte) (n int, err error) {
	return w.buff.Write(in)
}

func (w *OnCloseWriter) Close() (err error) {
	return w.onClose(w.buff, int64(w.buff.Len()))
}

// Returns a write-closer whose results (and length) will be fed to
// the named function at the close of the returned descriptor.
//
// You may specify any 'bytes.Buffer' like object that offers Read/Write/Len functions
// or leave it nil, and a default bytes.NewBuffer() will be created
func NewOnCloseWriter(buff LenReadWriter, f func(r io.Reader, len int64) (err error)) io.WriteCloser {
	if buff == nil {
		buff = bytes.NewBuffer(nil)
	}
	return &OnCloseWriter{buff: buff, onClose: f}
}
