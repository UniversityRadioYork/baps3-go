package bifrost

import "fmt"

// MessageWord is a token representing a message word known to baps3-go.
// While the BAPS3 API allows for arbitrarily many message words to exist, we
// only handle a small, finite set of them.  For simplicity of later
// comparison, we 'intern' the ones we know by converting them to a MessageWord
// upon creation of the Message representing their parent message.
type StatusCode int

const (

	// BadCode denotes a message with an unknown and ill-formed word.
	BadCode StatusCode = iota

	// - Core

	// StatusOk denotes a successful request
	StatusOk

	// StatusNotfound denotes a invalid resource in a request
	StatusNotFound

	// StatusInvalid is ???
	StatusInvalid

	// StatusNotAllowed denotes a request is not allowed on the given resource (e.g. delete on /control/serverid)
	StatusNotAllowed

	StatusError
)

var statusStrings = []string{
	"<BAD CODE>", // BadWord
	"OK",         // StatusOk
	"NOTFOUND",   // StatusNotFound
	"INVALID",    // StatusInvalid
	"NOTALLOWED", // StatusNotAllowed
	"ERROR",      // StatusError
}

func (status StatusCode) String() string {
	return statusStrings[int(status)]
}

func LookupStatus(status string) StatusCode {
	// This is O(n) on the size of StatusStrings, which is unfortunate, but
	// probably ok.
	for i, str := range statusStrings {
		if str == status {
			return StatusCode(i)
		}
	}
	return BadCode
}

type Status struct {
	Code    StatusCode
	Message string
}

func NewStatus(code StatusCode) *Status {
	s := new(Status)
	s.Code = code
	return s
}

func (s *Status) String() string {
	return fmt.Sprintf("%s %s", s.Code.String(), s.Message)
}
