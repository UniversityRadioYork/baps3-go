package message

import "io"

// Reader wraps a ReadCloser to provide message-level reading functionality.
type Reader struct {
	tok    *Tokeniser
	reader io.ReadCloser
	buf    [4096]byte
	pos    int
	max    int
}

// Close closes the Reader's underlying ReadCloser.
func (r *Reader) Close() error {
	return r.reader.Close()
}

// tokeniseUntilLine drains t's internal buffer into its tokeniser until it runs out or produces a line.
func (r *Reader) tokeniseUntilLine() (line []string, lineok bool) {
	var nread int
	for r.pos < r.max && !lineok {
		nread, lineok, line = r.tok.TokeniseBytes(r.buf[r.pos:r.max])
		r.pos += nread
	}
	return
}

// fillFromReader fills t's internal buffer using its reader.
// It can fail with errors from the reader.
func (r *Reader) fillFromReader() (err error) {
	r.pos = 0
	r.max, err = r.reader.Read(r.buf[:])
	return
}

// ReadLine reads a tokenised line from the Reader.
// ReadLine may return an error if the Reader chokes.
func (r *Reader) ReadLine() ([]string, error) {
	for {
		if line, lineok := r.tokeniseUntilLine(); lineok {
			return line, nil
		}
		if err := r.fillFromReader(); err != nil {
			return []string{}, err
		}
	}
}

// NewReader creates and returns a new, empty Reader.
// The Reader will read from the given ReadCloser when Tokenise is called.
// If closing is not required, use ioutil.NopCloser.
func NewReader(reader io.ReadCloser) *Reader {
	return &Reader{
		tok:    NewTokeniser(),
		reader: reader,
		pos:    0,
		max:    0,
	}
}
