package bifrost

import (
	"context"

	"github.com/UniversityRadioYork/bifrost-go/msgproto"
)

// File bifrost/endpoint.go describes clients that communicate at the level of Bifrost messages.

// Note: we use the Client and Endpoint structs in both sides of a client/server communication,
// hence why their channels are called Tx and Rx and not something more indicative (eg 'RequestTx' or 'ResponseRx').

// Endpoint contains the opposite end of a Client's channels.
type Endpoint struct {
	// Rx is the channel for receiving messages intended for the endpoint.
	Rx <-chan msgproto.Message

	// Tx is the channel for transmitting messages from the endpoint.
	Tx chan<- msgproto.Message
}

// Send tries to send a request on an Endpoint, modulo a context.
// It returns false if the given context has been cancelled.
//
// Send is just sugar over a Select between Tx and ctx.Done(), and it is
// ok to do this manually using the channels themselves.
func (e *Endpoint) Send(ctx context.Context, r msgproto.Message) bool {
	select {
	case <-ctx.Done():
		return false
	case e.Tx <- r:
	}
	return true
}

// NewEndpointPair creates a pair of Bifrost client channel sets.
func NewEndpointPair() (*Endpoint, *Endpoint) {
	res := make(chan msgproto.Message)
	req := make(chan msgproto.Message)

	left := Endpoint{
		Rx: res,
		Tx: req,
	}

	right := Endpoint{
		Tx: res,
		Rx: req,
	}

	return &left, &right
}
