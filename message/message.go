package message

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// MessageWord is a token representing a message word known to Bifrost.
// While the BAPS3 API allows for arbitrarily many message words to exist, we
// only handle a small, finite set of them.  For simplicity of later
// comparison, we 'intern' the ones we know by converting them to a MessageWord
// upon creation of the Message representing their parent message.
type MessageWord int

const (
	// Message word constants.
	//
	// When adding to this, also add the string equivalent to LookupRequest and
	// LookupResponse.
	//
	// Also note that the use of the same iota-run of numbers for requests,
	// responses and errors is intentional, because all three series of
	// message are conveyed in the same struct, parsed by the same
	// functions, and consequently the message word is referenced by code
	// with no understanding of whether the word pertains to a request, a
	// response, or something completely different.
	//

	// BadWord denotes a message with an unknown and ill-formed word.
	BadWord MessageWord = iota

	// - Requests

	// RqUnknown denotes a message with an unknown but valid request word.
	RqUnknown

	// RqRead denotes a 'read' request message.
	RqRead

	// RqWrite denotes a 'write' request message.
	RqWrite

	// RqDelete denotes a 'delete' request message.
	RqDelete

	// - Responses

	// RsUnknown denotes a message with an unknown but valid response word.
	RsUnknown

	// RsRes denotes a message with the 'RES' response.
	RsRes

	// RsUpdate denotes a message with the 'RES' response.
	RsUpdate

	// RsAck denotes a message with the 'ACK' response.
	RsAck

	// RsOhai denotes a message with the 'OHAI' response.
	RsOhai
)

// Go can't do constant arrays :(
var wordStrings = []string{
	"<BAD WORD>",         // BadWord
	"<UNKNOWN REQUEST>",  // RqUnknown
	"read",               // RqRead
	"write",              // RqWrite
	"delete",             // RqDelete
	"<UNKNOWN RESPONSE>", // RsUnknown
	"RES",                // RsRes
	"UPDATE",             // RsUpdate
	"ACK",                // RsAck
	"OHAI",               // RsOhai
}

func (word MessageWord) String() string {
	return wordStrings[int(word)]
}

// IsUnknown returns whether word represents an unknown message word.
func (word MessageWord) IsUnknown() bool {
	return word == BadWord || word == RqUnknown || word == RsUnknown
}

// LookupWord finds the equivalent MessageWord for a string.
// If the message word is not known, it will check whether the word is a
// valid request (all lowercase) or a valid response (all uppercase),
// returning RqUnknown or RsUnknown respectively.  Failing this, it will return
// BadWord.
func LookupWord(word string) MessageWord {
	// This is O(n) on the size of WordStrings, which is unfortunate, but
	// probably ok.
	for i, str := range wordStrings {
		if str == word {
			return MessageWord(i)
		}
	}

	// In BAPS3, lowercase words are requests; uppercase words are responses.
	if strings.ToLower(word) == word {
		return RqUnknown
	} else if strings.ToUpper(word) == word {
		return RsUnknown
	}
	return BadWord
}

// Message is a structure representing a full BAPS3 message.
// It is comprised of a word, which is stored as a MessageWord, and zero or
// more string arguments.
type Message struct {
	word MessageWord
	args []string
}

// New creates and returns a new Message with the given message word.
// The message will initially have no arguments; use AddArg to add arguments.
func New(word MessageWord) *Message {
	return &Message{
		word: word,
	}
}

// AddArg adds the given argument to a Message in-place.
// The given Message-pointer is returned, to allow for chaining.
func (m *Message) AddArg(arg string) *Message {
	m.args = append(m.args, arg)
	return m
}

func escapeArgument(input string) string {
	return "'" + strings.Replace(input, "'", `'\''`, -1) + "'"
}

// Pack outputs the given Message as raw bytes representing a BAPS3 message.
// These bytes can be sent down a TCP connection to a BAPS3 server, providing
// they are terminated using a line-feed character.
func (m *Message) Pack() (packed []byte, err error) {
	output := new(bytes.Buffer)

	_, err = output.WriteString(m.word.String())
	if err != nil {
		return
	}

	for _, a := range m.args {
		// Escape arg if needed
		for _, c := range a {
			if c < unicode.MaxASCII && (unicode.IsSpace(c) || strings.ContainsRune(`'"\`, c)) {
				a = escapeArgument(a)
				break
			}
		}

		_, err = output.WriteString(" " + a)
		if err != nil {
			return
		}
	}
	output.WriteString("\n")

	packed = output.Bytes()
	return
}

// Word returns the MessageWord of the given Message.
func (m *Message) Word() MessageWord {
	return m.word
}

// Args returns the slice of Arguments.
func (m *Message) Args() []string {
	return m.args
}

// Arg returns the index-th argument of the given Message.
// The first argument is argument 0.
// If the argument does not exist, an error is returned via err.
func (m *Message) Arg(index int) (arg string, err error) {
	if index < 0 {
		err = fmt.Errorf("got negative index %d", index)
	} else if len(m.args) <= index {
		err = fmt.Errorf("wanted argument %d, only %d arguments", index, len(m.args))
	} else {
		arg = m.args[index]
	}
	return
}

func (m *Message) String() (outstr string) {
	outstr = m.word.String()
	for _, s := range m.args {
		outstr += " " + s
	}
	return
}

// lineToMessage constructs a Message struct from a line of word-strings.
func LineToMessage(line []string) (msg *Message, err error) {
	if len(line) == 0 {
		err = fmt.Errorf("cannot construct message from zero words")
	} else {
		msg = New(LookupWord(line[0]))
		for _, arg := range line[1:] {
			msg.AddArg(arg)
		}
	}

	return
}
