package core

// File core/ohai.go describes parsing and emitting routines for the OHAI core request.

import (
	"github.com/UniversityRadioYork/bifrost-go/message"
)

const (
	// RsOhai is the Bifrost response word OHAI.
	RsOhai = "OHAI"

	// ThisProtocolVer represents the Bifrost protocol version this library represents.
	ThisProtocolVer = "bifrost-0.0.0"
)

// OhaiResponse represents the information contained within an OHAI response.
type OhaiResponse struct {
	// ProtocolVer is the semantic-version identifier for the Bifrost protocol.
	ProtocolVer string
	// ProtocolVer is the semantic-version identifier for the server itself.
	ServerVer string
}

// Message converts an OhaiResponse into an OHAI message with tag tag.
func (o *OhaiResponse) Message(tag string) *message.Message {
	return message.New(tag, RsOhai).AddArgs(o.ProtocolVer, o.ServerVer)
}

// ParseOhaiResponse tries to parse an arbitrary message as an OHAI response.
func ParseOhaiResponse(m *message.Message) (resp *OhaiResponse, err error) {
	if err = CheckWord(RsOhai, m); err != nil {
		return nil, err
	}

	var pv, sv string
	if pv, sv, err = TwoArgs(m); err != nil {
		return nil, err
	}

	r := OhaiResponse{ProtocolVer: pv, ServerVer: sv}
	return &r, nil
}
