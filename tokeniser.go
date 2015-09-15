package bifrost

import (
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
	word             []byte
	words            []string
	reader           io.Reader
}

// NewTokeniser creates and returns a new, empty Tokeniser.
// The Tokeniser will read from the given Reader when Tokenise is called.
func NewTokeniser(reader io.Reader) *Tokeniser {
	return &Tokeniser{
		escapeNextChar:   false,
		currentQuoteType: none,
		word:             []byte{},
		inWord:           false,
		words:            []string{},
		reader:           reader,
	}
}

func (t *Tokeniser) endWord() {
	if !t.inWord {
		// Don't add an empty word.
		return
	}

	t.words = append(t.words, string(t.word))
	t.word = []byte{}
	t.inWord = false
}

// Tokenise reads a tokenised line from the Reader.
//
// Tokenise may return an error if the Reader chokes.
func (t *Tokeniser) Tokenise() ([]string, error) {
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

		lineDone := t.tokeniseByte(bs[0])
		// Have we finished a line?
		// If so, clean up for another tokenising, and return it.
		if lineDone {
			line := t.words
			t.words = []string{}
			return line, nil
		}
	}
}

// tokeniseByte tokenises a single byte.
// It returns true if we've finished a line, which can only occur outside of
// quotes
func (t *Tokeniser) tokeniseByte(b byte) (endLine bool) {
	endLine = false

	if t.escapeNextChar {
		t.put(b)
		t.escapeNextChar = false
		return
	}

	switch t.currentQuoteType {
	case none:
		endLine = t.tokeniseNoQuotes(b)
	case single:
		t.tokeniseSingleQuotes(b)
	case double:
		t.tokeniseDoubleQuotes(b)
	}

	return
}

// tokeniseNoQuotes tokenises a single byte outside quote characters.
// It returns true if we've finished a line, and any error that occurred while
// tokenising.
func (t *Tokeniser) tokeniseNoQuotes(b byte) (endLine bool) {
	endLine = false

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
		// We're ending the current word as well as a line.
		t.endWord()
		endLine = true
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

	return
}

// tokeniseSingleQuotes tokenises a single byte within single quotes.
// It doesn't need to return whether we've finished a line, because we can't finish
// a line in quotes.
func (t *Tokeniser) tokeniseSingleQuotes(b byte) {
	switch b {
	case '\'':
		t.currentQuoteType = none
	default:
		t.put(b)
	}
}

// tokeniseDoubleQuotes tokenises a single byte within double quotes.
// It doesn't need to return whether we've finished a line, because we can't finish
// a line in quotes.
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

// put adds a byte to the Tokeniser's word.
func (t *Tokeniser) put(b byte) {
	t.inWord = true
	t.word = append(t.word, b)
}
