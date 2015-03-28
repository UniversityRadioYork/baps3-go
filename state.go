package baps3

import "errors"

type State int

const (
	StUnknown State = iota
	StReady
	StQuitting
	StEjected
	StPlaying
	StStopped
)

var stateStrings = []string{
	"<UNKNOWN STATE>", // StUnknown
	"Ready",           // StReady
	"Quitting",        // StQuitting
	"Ejected",         // StEjected
	"Playing",         // StPlaying
	"Stopped",         // StStopped
}

func (state State) String() string {
	return stateStrings[int(state)]
}

// LookupState returns the corresponding State for a string, or StUnknown if unrecognised.
func LookupState(statestr string) (s State, err error) {
	for i, str := range stateStrings {
		if str == statestr {
			return State(i), nil
		}
	}
	return StUnknown, errors.New("Unrecognised state")
}
