package bifrost

import (
	"bytes"
	"io"
	"unicode"
)

// quoteType represents one of the types of quoting used in the BAPS3 protocol.
type quoteType int

const (
	// none represents the state between quoted parts of a BAPS3 message.
	none quoteType = iota

	// single represents 'single quoted' parts of a BAPS3 message.
	single

	// double represents "double quoted" parts of a BAPS3 message.
	double

	// bufsize is the number of bytes the Tokeniser will try to read from
	// its Reader in one go.
	bufsize int = 4096
)

// Tokeniser holds the state of a Bifrost protocol tokeniser.
type Tokeniser struct {
	// raw is the current back-buffer of bytes to tokenise.
	raw []byte
	// rawpos is the current position in `raw`.
	rawpos int
	// rawcount is the number of valid bytes in `raw`.
	rawcount int

	inWord           bool
	escapeNextChar   bool
	currentQuoteType quoteType
	word             *bytes.Buffer
	words            []string
	lineDone         bool
	err              error
	reader           io.Reader
}

// NewTokeniser creates and returns a new, empty Tokeniser.
// The Tokeniser will read from the given Reader when Tokenise is called.
func NewTokeniser(reader io.Reader) *Tokeniser {
	return NewTokeniserWith(reader, bufsize)
}

// NewTokeniserWith is NewTokeniser, but with a custom internal buffer size.
func NewTokeniserWith(reader io.Reader, size int) *Tokeniser {
	t := new(Tokeniser)

	t.raw = make([]byte, size)
	t.rawpos = 0
	t.rawcount = 0

	t.escapeNextChar = false
	t.currentQuoteType = none
	t.word = new(bytes.Buffer)
	t.inWord = false
	t.words = []string{}
	t.lineDone = false
	t.err = nil
	t.reader = reader
	return t
}

func (t *Tokeniser) endLine() {
	// We might still be in the middle of a word.
	t.endWord()
	t.lineDone = true
}

func (t *Tokeniser) endWord() {
	if !t.inWord {
		// Don't add an empty word.
		return
	}

	// This ensures any non-UTF8 is replaced with the Unicode replacement
	// character.  We could use String(), but this would permit invalid
	// UTF8.
	uword := []rune{}
	for {
		r, _, err := t.word.ReadRune()
		if err != nil {
			break
		}
		uword = append(uword, r)
	}

	t.words = append(t.words, string(uword))
	t.word.Truncate(0)
	t.inWord = false
}

// Tokenise reads a tokenised line from the Reader.
//
// Tokenise may return an error if its current word gets over-full, or the Reader chokes.
// In the former case, the Tokeniser will need to be replaced.
func (t *Tokeniser) Tokenise() ([]string, error) {
	// Have we previously suffered a permanent tokenising error?
	// If so, bail with it.
	if t.err != nil {
		return []string{}, t.err
	}

	for {
		// First, tokenise everything in our buffer.
		for ; t.rawpos < t.rawcount; t.rawpos++ {
			t.tokeniseByte(t.raw[t.rawpos])
			if t.err != nil {
				return []string{}, t.err
			}

			// Have we finished a line?
			// If so, clean up for another tokenising, and return it.
			if t.lineDone {
				// The t.rawpos++ above won't fire if we leave now.
				// Thus, we need to do it here.
				t.rawpos++

				t.lineDone = false
				line := t.words
				t.words = []string{}
				return line, nil
			}
		}
		// We've run out of buffer now, so prod the Reader.
		n, err := t.reader.Read(t.raw)
		if err != nil {
			return []string{}, err
		}
		t.rawpos, t.rawcount = 0, n
	}
}

// tokeniseByte tokenises a single byte.
func (t *Tokeniser) tokeniseByte(b byte) {
	if t.escapeNextChar {
		t.put(b)
		t.escapeNextChar = false
		return
	}

	switch t.currentQuoteType {
	case none:
		t.tokeniseNoQuotes(b)
	case single:
		t.tokeniseSingleQuotes(b)
	case double:
		t.tokeniseDoubleQuotes(b)
	}
}

// tokeniseNoQuotes tokenises a single byte outside quote characters.
func (t *Tokeniser) tokeniseNoQuotes(b byte) {
	switch b {
	case '\'':
		// Switching into single quotes mode starts a word.
		// This is to allow '' to represent the empty string.
		t.inWord = true
		t.currentQuoteType = single
	case '"':
		// Switching into double quotes mode starts a word.
		// This is to allow "" to represent the empty string.
		t.inWord = true
		t.currentQuoteType = double
	case '\\':
		t.escapeNextChar = true
	case '\n':
		t.endLine()
	default:
		// Note that this will only check for ASCII
		// whitespace, because we only pass it one byte
		// and non-ASCII whitespace is >1 UTF-8 byte.
		if unicode.IsSpace(rune(b)) {
			t.endWord()
		} else {
			t.put(b)
		}
	}
}

// tokeniseSingleQuotes tokenises a single byte within single quotes.
func (t *Tokeniser) tokeniseSingleQuotes(b byte) {
	switch b {
	case '\'':
		t.currentQuoteType = none
	default:
		t.put(b)
	}
}

// tokeniseDoubleQuotes tokenises a single byte within double quotes.
func (t *Tokeniser) tokeniseDoubleQuotes(b byte) {
	switch b {
	case '"':
		t.currentQuoteType = none
	case '\\':
		t.escapeNextChar = true
	default:
		t.put(b)
	}
}

// put adds a byte to the Tokeniser's buffer.
// If the buffer is too big, an error will be raised and propagated to the
// Tokeniser's user.
func (t *Tokeniser) put(b byte) {
	t.err = t.word.WriteByte(b)
	t.inWord = true
}
