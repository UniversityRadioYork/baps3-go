package baps3

import (
	"fmt"
	"testing"
)

// TestHasFeature tests whether ServiceState.hasFeature seems to work.
func TestHasFeature(t *testing.T) {
	cases := []struct {
		feat    Feature
		present bool
	}{
		// We check the presence of some features and the absence of
		// others.  This is a shuffled, but even distribution of both.
		{FtFileLoad, true},
		{FtPlayStop, true},
		{FtSeek, false},
		{FtEnd, true},
		{FtTimeReport, false},
		{FtPlaylist, true},
		{FtPlaylistAutoAdvance, false},
		{FtPlaylistTextItems, false},
	}

	// This is for collecting the features we do want to enable.
	presents := []Feature{}

	// All features should be absent on a new ServiceState.
	srv := InitServiceState()

	for _, c := range cases {
		if srv.HasFeature(c.feat) {
			t.Errorf("initial ServiceState shouldn't have feature %q", c.feat)
		}
		if c.present {
			presents = append(presents, c.feat)
		}
	}

	// Now set the features we want.
	msg := NewMessage(RsFeatures)
	for _, p := range presents {
		msg.AddArg(p.String())
	}

	if err := srv.Update(*msg); err != nil {
		t.Errorf("error when setting features: %s", err)
	}

	// Now check if HasFeature works (!)
	for _, d := range cases {
		has := srv.HasFeature(d.feat)
		if has && !d.present {
			t.Errorf("service should not have feature %q, but does", d.feat)
		} else if !has && d.present {
			t.Errorf("service should have feature %q, but does not", d.feat)
		}
	}
}

// TestServiceStateUpdateFail tests the behaviour of a ServiceState when it
// receives a malformed message.
func TestServiceStateUpdateFail(t *testing.T) {
	// TODO(CaptainHayashi): maybe test what the error actually is
	cases := []struct {
		msg    *Message
		hasErr bool
	}{
		// Request where response was expected
		{
			NewMessage(RqLoad).AddArg("/quux"),
			false, // TODO(CaptainHayashi): error on requests?
		},
		// Too few arguments
		{
			NewMessage(RsFile),
			true,
		},
		// Too many arguments
		{
			NewMessage(RsTime).AddArg("3003").AddArg("lol"),
			true,
		},
		// Unknown request (should be ignored)
		{
			NewMessage(RsUnknown).AddArg("heh"),
			false,
		},
	}

	for _, c := range cases {
		err := InitServiceState().Update(*c.msg)
		if c.hasErr && (err == nil) {
			t.Errorf("expected %q to produce error, none produced", c.msg)
		} else if !c.hasErr && (err != nil) {
			t.Errorf("expected %q not to produce error, one produced", c.msg)
		}
	}
}

// TestServiceStateUpdate tests the updating of a ServiceState by messages.
func TestServiceStateUpdate(t *testing.T) {
	// TODO(CaptainHayashi): test failure states as well as successes

	cases := []struct {
		msg *Message
		cmp func(*ServiceState) error
	}{
		{
			NewMessage(RsFeatures).AddArg("End").AddArg("FileLoad"),
			func(s *ServiceState) (err error) {
				_, endIn := s.Features[FtEnd]
				_, flIn := s.Features[FtFileLoad]

				if !endIn || !flIn {
					err = fmt.Errorf(
						"features should contain End and Fileload, got %d",
						s.Features,
					)
				}

				return
			},
		},
		{
			NewMessage(RsFile).AddArg("/home/foo/bar.mp3"),
			func(s *ServiceState) (err error) {
				if s.File != "/home/foo/bar.mp3" {
					err = fmt.Errorf(
						"file should be %d, got %d",
						"/home/foo/bar.mp3",
						s.File,
					)
				}

				return
			},
		},
		{
			NewMessage(RsState).AddArg("Ejected"),
			func(s *ServiceState) (err error) {
				if s.State != "Ejected" {
					err = fmt.Errorf(
						"state should be %d, got %d",
						"Ejected",
						s.State,
					)
				}

				return
			},
		},
		{
			NewMessage(RsTime).AddArg("1337000000"),
			func(s *ServiceState) (err error) {
				if s.Time.Seconds() != 1337 {
					err = fmt.Errorf(
						"time should be %i secs, got %i",
						1337,
						s.Time,
					)
				}

				return
			},
		},
	}

	for _, c := range cases {
		st := InitServiceState()
		if err := st.Update(*c.msg); err != nil {
			t.Errorf("error when sending %d: %s", c.msg, err)
		}
		if err := c.cmp(st); err != nil {
			t.Errorf("sent %d, but got error: %s", c.msg, err)
		}
	}

}
