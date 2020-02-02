package core

import (
	"fmt"
	"testing"

	"github.com/UniversityRadioYork/bifrost-go/message"
)

// ExampleParseIamaResponse is a testable example for ParseIamaResponse.
func ExampleParseIamaResponse() {
	m := message.New(message.TagBcast, RsIama).AddArgs("player/file")
	if o, err := ParseIamaResponse(m); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Role:", o.Role)
	}

	// Output:
	// Role: player/file
}

var iamaResponseRoundTripCases = []IamaResponse{
	IamaResponse{"list"}, IamaResponse{"player/file"}, IamaResponse{"x/foobar"},
}

// TestParseIamaResponse_roundTrip checks that parsing the result of OhaiResponse's Message method returns a similar
// IamaResponse.
func TestParseIamaResponse_roundTrip(t *testing.T) {
	for _, c := range iamaResponseRoundTripCases {
		m := c.Message(message.TagBcast)

		if got, err := ParseIamaResponse(m); err != nil {
			t.Errorf("parse error: %v", err)
		} else if got.Role != c.Role {
			t.Errorf("got role %q; want %q", got.Role, c.Role)
		}
	}
}

// TestParseIamaResponse_wordError checks that ParseIamaResponse handles word errors correctly.
func TestParseIamaResponse_wordError(t *testing.T) {
	// Hopefully this is a decently representative set of variations on OHAI.
	cases := []string{
		"",
		"iama",
		"IAM",
		"Iama",
		"IAMAI",
	}

	for _, word := range cases {
		m := message.New(message.TagBcast, word).AddArgs("test")

		_, err := ParseIamaResponse(m)
		testParserWordError(t, err, RsIama, word)
	}
}

// TestParseIamaResponse_arityErrors checks that ParseIamaResponse handles arity errors correctly.
func TestParseIamaResponse_arityErrors(t *testing.T) {
	cases := [][]string{
		{},
		{"player/file", "but wait there's more!"},
	}

	for _, args := range cases {
		m := message.New(message.TagBcast, RsIama).AddArgs(args...)

		_, err := ParseIamaResponse(m)
		testArgUnpackError(t, "ParseIamaResponse", 1, len(args), err)
	}
}
