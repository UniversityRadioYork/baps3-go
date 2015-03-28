package baps3

// The part of the BAPS3 connector code responsible for updating the
// internal state.

import (
	"fmt"
	"strconv"
	"time"
)

// ServiceState is the struct of all known state for a BAPS3 service.
// TODO(CaptainHayashi): possibly segregate more by feature, so elements not
// relevant to the current feature set aren't allocated?
type ServiceState struct {
	// Core
	Identifier string
	Features   FeatureSet
	State      State

	// TimeReport
	Time time.Duration

	// FileLoad
	File string
}

// InitServiceState creates a new, blank, ServiceState.
func InitServiceState() (s *ServiceState) {
	s = new(ServiceState)
	s.Features = FeatureSet{}
	s.State = StReady

	return
}

// Maps response word to a handler function to update servicestate
var updateFunctionForResponse = map[MessageWord]func(*ServiceState, Message) error{
	RsOhai:     (*ServiceState).updateIdentifierFromMessage,
	RsFeatures: (*ServiceState).updateFeaturesFromMessage,
	RsFile:     (*ServiceState).updateFileFromMessage,
	RsState:    (*ServiceState).updateStateFromMessage,
	RsTime:     (*ServiceState).updateTimeFromMessage,
}

// Update updates a ServiceState according to an incoming service response.
func (s *ServiceState) Update(res Message) (err error) {
	updateFunc, ok := updateFunctionForResponse[res.Word()]
	if ok {
		err = updateFunc(s, res)
	}

	return
}

func (s *ServiceState) updateIdentifierFromMessage(res Message) (err error) {
	if len(res.AsSlice()[1:]) > 1 {
		return fmt.Errorf("Too many arguments in %q", res)
	}
	if ident, ok := res.Arg(0); ok != nil {
		s.Identifier = ""
		err = fmt.Errorf("No identifier present in %q", res)
	} else {
		s.Identifier = ident
	}
	return
}

func (s *ServiceState) updateFeaturesFromMessage(res Message) (err error) {
	feats, err := FeatureSetFromMsg(&res)
	if err == nil {
		s.Features = feats
	}
	return
}

func (s *ServiceState) updateFileFromMessage(res Message) (err error) {
	// Expecting only one argument
	if _, err := res.Arg(1); err == nil {
		return fmt.Errorf("too many arguments in %q", res)
	}

	file, err := res.Arg(0)
	if err != nil {
		s.File = ""
		return
	}

	s.File = file

	return
}

func (s *ServiceState) updateStateFromMessage(res Message) (err error) {
	// Expecting only one argument
	if _, err := res.Arg(1); err == nil {
		return fmt.Errorf("too many arguments in %q", res)
	}

	statestr, err := res.Arg(0)
	if err != nil {
		// TODO(CaptainHayashi): "Ready" is currently the most
		// valid fallback value in the spec.  Does the spec
		// need an 'I don't know what the state is' value?
		s.State = StReady
		return
	}

	state, err := LookupState(statestr)
	if err != nil {
		s.State = StReady // TODO(wlcx): see above todo
		return
	}
	s.State = state

	return
}

func (s *ServiceState) updateTimeFromMessage(res Message) (err error) {
	// Expecting only one argument
	if _, err := res.Arg(1); err == nil {
		return fmt.Errorf("too many arguments in %q", res)
	}

	usecs, err := res.Arg(0)
	if err != nil {
		return
	}

	usec, err := strconv.Atoi(usecs)
	if err != nil {
		return
	}

	s.Time = time.Duration(usec) * time.Microsecond

	return
}

// HasFeature returns whether the connected server advertises the given feature.
func (s *ServiceState) HasFeature(f Feature) bool {
	_, ok := s.Features[f]
	return ok
}
