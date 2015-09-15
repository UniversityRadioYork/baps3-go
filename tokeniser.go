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
)

// Tokeniser holds the state of a Bifrost protocol tokeniser.
type Tokeniser struct {
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
	return &Tokeniser{
		escapeNextChar:   false,
		currentQuoteType: none,
		word:             new(bytes.Buffer),
		inWord:           false,
		words:            []string{},
		lineDone:         false,
		err:              nil,
		reader:           reader,
	}
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

	// As per http://grokbase.com/t/gg/golang-nuts/139fgmycba
	var bs [1]byte

	for {
		// Constantly grab one byte out of the Reader.
		// Technically inefficient, but this will be done on network
		// connections mainly anyway, so this shouldn't be the
		// bottleneck.
		n, err := t.reader.Read(bs[:])
		if err != nil {
			return []string{}, err
		}
		// Spin until we get a byte.
		if n == 0 {
			continue
		}

		t.tokeniseByte(bs[0])
		if t.err != nil {
			return []string{}, t.err
		}

		// Have we finished a line?
		// If so, clean up for another tokenising, and return it.
		if t.lineDone {
			t.lineDone = false
			line := t.words
			t.words = []string{}
			return line, nil
		}
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
