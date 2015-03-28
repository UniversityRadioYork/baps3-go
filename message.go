package baps3

import (
	"fmt"
	"strings"
)

// MessageWord is a token representing a message word known to Bifrost.
// While the BAPS3 API allows for arbitrarily many message words to exist, we
// only handle a small, finite set of them.  For simplicity of later
// comparison, we 'intern' the ones we know by converting them to a MessageWord
// upon creation of the Message representing their parent message.
type MessageWord int

const (
	/* Message word constants.
	 *
	 * When adding to this, also add the string equivalent to LookupRequest and
	 * LookupResponse.
	 *
	 * Also note that the use of the same iota-run of numbers for requests,
	 * responses and errors is intentional, because all three series of
	 * message are conveyed in the same struct, parsed by the same
	 * functions, and consequently the message word is referenced by code
	 * with no understanding of whether the word pertains to a request, a
	 * response, or something completely different.
	 */

	// BadWord denotes a message with an unknown and ill-formed word.
	BadWord MessageWord = iota

	// - Requests

	// RqUnknown denotes a message with an unknown but valid request word.
	RqUnknown

	/* -- Core
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/core.html#requests
	 */

	// RqQuit denotes a 'quit' request message.
	RqQuit

	// RqDump denotes a 'dump' request message.
	RqDump

	/* -- PlayStop feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-playstop.html#requests
	 */

	// RqPlay denotes a 'play' request message.
	RqPlay

	// RqStop denotes a 'stop' request message.
	RqStop

	/* -- FileLoad feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-fileload.html#requests
	 */

	// RqEject denotes an 'eject' request message.
	RqEject

	// RqLoad denotes a 'load' request message.
	RqLoad

	/* -- Seek feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-seek.html#requests
	 */

	// RqSeek denotes a 'seek' request message.
	RqSeek

	/* -- Playlist feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-playlist.html#requests
	 */

	// RqDequeue denotes a 'dequeue' request message.
	RqDequeue

	// RqEnqueue denotes an 'enqueue' request message.
	RqEnqueue

	// RqList denotes a 'list' request message.
	RqList

	// RqSelect denotes a 'select' request message.
	RqSelect

	/* -- Playlist.AutoAdvance feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-autoadvance.html#requests
	 */

	// RqAutoAdvance denotes an 'autoadvance' request message.
	RqAutoAdvance

	// - Responses

	// RsUnknown denotes a message with an unknown but valid response word.
	RsUnknown

	/* -- Core
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/core.html#responses
	 */

	// RsOk denotes a message with the 'OK' response.
	RsOk

	// RsFail denotes a message with the 'FAIL' response.
	RsFail

	// RsWhat denotes a message with the 'WHAT' response.
	RsWhat

	// RsOhai denotes a message with the 'OHAI' response.
	RsOhai

	// RsFeatures denotes a message with the 'FEATURES' response.
	RsFeatures

	// RsState denotes a message with the 'STATE' response.
	RsState

	/* -- End feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-end.html#responses
	 */

	// RsEnd denotes a message with the 'END' response.
	RsEnd

	/* -- FileLoad feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-fileload.html#responses
	 */

	// RsFile denotes a message with the 'FILE' response.
	RsFile

	/* -- TimeReport feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-timereport.html#responses
	 */

	// RsTime denotes a message with the 'TIME' response.
	RsTime

	/* -- Playlist feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-playlist.html#responses
	 */

	// RsCount denotes a message with the 'COUNT' response.
	RsCount

	// RsDequeue denotes a message with the 'DEQUEUE' response.
	RsDequeue

	// RsEnqueue denotes a message with the 'ENQUEUE' response.
	RsEnqueue

	// RsItem denotes a message with the 'ITEM' response.
	RsItem

	// RsSelect denotes a message with the 'SELECT' response.
	RsSelect

	/* -- Playlist.AutoAdvance feature
	 * http://universityradioyork.github.io/baps3-spec/comms/internal/feature-autoadvance.html#responses
	 */

	// RqAutoAdvance denotes a message with the 'AUTOADVANCE' response.
	RsAutoAdvance
)

// Yes, a global variable.
// Go can't handle constant arrays.
var wordStrings = []string{
	"<BAD WORD>",         // BadWord
	"<UNKNOWN REQUEST>",  // RqUnknown
	"quit",               // RqQuit
	"dump",               // RqDump
	"play",               // RqPlay
	"stop",               // RqStop
	"eject",              // RqEject
	"load",               // RqLoad
	"seek",               // RqSeek
	"dequeue",            // RqDequeue
	"enqueue",            // RqEnqueue
	"list",               // RqList
	"select",             // RqSelect
	"autoadvance",        // RqAutoAdvance
	"<UNKNOWN RESPONSE>", // RsUnknown
	"OK",                 // RsOk
	"FAIL",               // RsFail
	"WHAT",               // RsWhat
	"OHAI",               // RsOhai
	"FEATURES",           // RsFeatures
	"STATE",              // RsState
	"END",                // RsEnd
	"FILE",               // RsFile
	"TIME",               // RsTime
	"COUNT",              // RsCount
	"DEQUEUE",            // RsDequeue
	"ENQUEUE",            // RsEnqueue
	"ITEM",               // RsItem
	"SELECT",             // RsSelect
	"AUTOADVANCE",        // RsAutoAdvance
}

// IsUnknown returns whether word represents a message word unknown to Bifrost.
func (word MessageWord) IsUnknown() bool {
	return word == BadWord || word == RqUnknown || word == RsUnknown
}

func (word MessageWord) String() string {
	return wordStrings[int(word)]
}

// LookupWord finds the equivalent MessageWord for a string.
// If the message word is not known to Bifrost, it will check whether the word
// is a valid request (all lowercase) or a valid response (all uppercase),
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

// NewMessage creates and returns a new Message with the given message word.
// The message will initially have no arguments; use AddArg to add arguments.
func NewMessage(word MessageWord) *Message {
	m := new(Message)
	m.word = word
	return m
}

// AddArg adds the given argument to a Message in-place.
// The given Message-pointer is returned, to allow for chaining.
func (m *Message) AddArg(arg string) *Message {
	m.args = append(m.args, arg)
	return m
}

// AsSlice outputs the given Message as a string slice.
// The slice contains the string form of the Message's word in index 0, and
// the arguments as index 1 upwards, if any.
func (m *Message) AsSlice() []string {
	slice := []string{m.word.String()}
	for _, arg := range m.args {
		slice = append(slice, arg)
	}
	return slice
}

// Pack outputs the given Message as raw bytes representing a BAPS3 message.
// These bytes can be sent down a TCP connection to a BAPS3 server, providing
// they are terminated using a line-feed character.
func (m *Message) Pack() ([]byte, error) {
	return Pack(m.word.String(), m.args)
}

// Word returns the MessageWord of the given Message.
func (m *Message) Word() MessageWord {
	return m.word
}

// Arg returns the index-th argument of the given Message.
// The first argument is argument 0.
// If the argument does not exist, an error is returned via err.
func (m *Message) Arg(index int) (arg string, err error) {
	if index < 0 {
		err = fmt.Errorf("Arg got negative index %d", index)
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
		msg = NewMessage(LookupWord(line[0]))
		for _, arg := range line[1:] {
			msg.AddArg(arg)
		}
	}

	return
}
