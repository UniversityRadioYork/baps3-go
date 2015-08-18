package bifrost

import "testing"

func TestState(t *testing.T) {
	cases := []struct {
		str  string
		want State
	}{
		{
			"Ready",
			StReady,
		},
		{
			"Ejected",
			StEjected,
		},
		// Test unknown state
		{
			"Rustling",
			StUnknown,
		},
	}

	for _, c := range cases {
		got, err := LookupState(c.str)
		if (err != nil) != (c.want == StUnknown) { // If err is not what we expect (not nil iff StUnknown)
			if err != nil {
				t.Errorf("Got err when expecting nil: %q", err)
			} else {
				t.Errorf("Got nil when expecting unknown state error")
			}
		}
		if got != c.want {
			t.Errorf("LookupState(%q) == %q, want %q", c.str, got, c.want)
		}

		if got != StUnknown { // Only do the other way if valid state
			if gotstr := c.want.String(); gotstr != c.str {
				t.Errorf("%q.String() == %q, want %q", c.want, gotstr, c.str)
			}
		}
	}

}
