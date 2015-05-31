package bifrost

import "testing"
import "reflect"

// cmpWords is defined in tokeniser_test.
// TODO(CaptainHayashi): move cmpWords elsewhere?

func TestMessageWord(t *testing.T) {
	cases := []struct {
		str     string
		word    MessageWord
		unknown bool
	}{
		// Ok, a request
		{"read", RqRead, false},
		// Ok, a response
		{"OHAI", RsOhai, false},
		// Unknown, but a request
		{"uwot", RqUnknown, true},
		// Unknown, but a response
		{"MATE", RsUnknown, true},
		// Unknown, and unclear what type of message
		{"MaTe", BadWord, true},
	}

	for _, c := range cases {
		gotword := LookupWord(c.str)
		if gotword != c.word {
			t.Errorf("LookupWord(%q) == %q, want %q", c.str, gotword, c.word)
		}
		if c.word.IsUnknown() != c.unknown {
			t.Errorf("%q.IsUnknown() == %q, want %q", c.word, !c.unknown, c.unknown)
		}
		// Only do the other direction if it's a valid response
		if !c.unknown {
			gotstr := c.word.String()
			if gotstr != c.str {
				t.Errorf("%q.String() == %q, want %q", c.word, gotstr, c.str)
			}
		}
	}
}

func TestMessage(t *testing.T) {
	cases := []struct {
		words []string
		msg   *Message
	}{
		// Empty request
		{[]string{"write"}, NewMessage(RqWrite)},
		// Request with one argument
		{[]string{"read", "/control/state"}, NewMessage(RqRead).AddArg("/control/state")},
		// Request with multiple argument
		{[]string{"write", "/player/time", "0"},
			NewMessage(RqWrite).AddArg("/player/time").AddArg("0"),
		},
		// Empty response
		{[]string{"RES"}, NewMessage(RsRes)},
		// Response with one argument
		{[]string{"OHAI", "playd 1.0.0"}, NewMessage(RsOhai).AddArg("playd 1.0.0")},
		// Response with multiple argument
		{[]string{"ACK", "int", "OK", "1337"},
			NewMessage(RsAck).AddArg("int").AddArg("OK").AddArg("1337"),
		},
	}

	for _, c := range cases {
		gotslice := c.msg.AsSlice()
		if !cmpWords(gotslice, c.words) {
			t.Errorf("%q.ToSlice() == %q, want %q", c.msg, gotslice, c.words)
		}
		gotword := LookupWord(c.words[0])
		if gotword != c.msg.Word() {
			t.Errorf("LookupWord(%q) == %q, but Word() == %q", c.words[0], gotword, c.msg.Word())
		}
	}

	// And now, test args.
	// TODO(CaptainHayashi): refactor the above to integrate this test
	args := []string{"bibbity", "bobbity", "boo"}
	msg := NewMessage(RsUnknown)
	for _, arg := range args {
		msg.AddArg(arg)
	}

	// Bounds checking
	for _, i := range []int{-1, len(args)} {
		if _, err := msg.Arg(i); err == nil {
			t.Errorf("Managed to get %dth arg of a %d-arged Message", i, len(args))
		}
	}

	for i, want := range args {
		got, err := msg.Arg(i)
		if err != nil {
			t.Errorf("unexpected error with Arg(%d)", i)
		} else if got != want {
			t.Errorf("Arg(%d) = %q, want %q", i, got, want)
		}
	}

	for _, c := range cases {
		got, err := LineToMessage(c.words)
		if err != nil {
			t.Errorf("unexpected error with: %q", got)
		} else if !reflect.DeepEqual(got, c.msg) {
			t.Errorf("Got %q, wanted %q", got, c.msg)
		}
	}
}
