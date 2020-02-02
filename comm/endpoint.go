package comm

import (
	"context"

	"github.com/UniversityRadioYork/bifrost-go/message"
)

// Note: we use the Endpoint structs in both sides of a client/server communication,
// hence why their channels are called Tx and Rx and not something more indicative (eg 'RequestTx' or 'ResponseRx').

// Endpoint describes a message-level Bifrost endpoint.
type Endpoint struct {
	// Rx is the channel for receiving messages intended for the endpoint.
	Rx <-chan message.Message

	// Tx is the channel for transmitting messages from the endpoint.
	Tx chan<- message.Message
}

// Send tries to send a request on an Endpoint, modulo a context.
// It returns false if the given context has been cancelled.
//
// Send is just sugar over a Select between Tx and ctx.Done(), and it is
// ok to do this manually using the channels themselves.
func (e *Endpoint) Send(ctx context.Context, r message.Message) bool {
	select {
	case <-ctx.Done():
		return false
	case e.Tx <- r:
	}
	return true
}

// NewEndpointPair creates a pair of Bifrost client channel sets.
func NewEndpointPair() (*Endpoint, *Endpoint) {
	res := make(chan message.Message)
	req := make(chan message.Message)

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
