package io

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/UniversityRadioYork/bifrost-go/message"
)

// HungUpError is the error sent by an Io when its transmission loop has hung up.
var HungUpError = errors.New("client has hung up")

// Io represents a Bifrost endpoint that sends and receives messages along an I/O connection.
type Conn struct {
	// Io holds the internal I/O connection.
	Io io.ReadWriteCloser

	// Bifrost holds the Bifrost channel pair used by the Io.
	Bifrost *Endpoint
}

func (c *Conn) Close() error {
	// TODO(@MattWindsor91): make sure we close everything
	close(c.Bifrost.Tx)
	return c.Io.Close()
}

// Run spins up the client's receiver and transmitter loops.
// It takes a channel to notify the caller asynchronously of any errors, and a client
// and the server's client hangup and done channels.
// It closes errors once both loops are done.
func (c *Conn) Run(ctx context.Context, errCh chan<- error) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		c.runTx(ctx, errCh)
		c.sendError(ctx, errCh, HungUpError)
		wg.Done()
	}()

	go func() {
		c.runRx(ctx, errCh)
		wg.Done()
	}()

	wg.Wait()
	close(errCh)
}

// runRx runs the client's message receiver loop.
// This writes messages to the socket.
func (c *Conn) runRx(ctx context.Context, errCh chan<- error) {
	// We don't have to check c.Bifrost.Done here:
	// client always drops both Rx and Done when shutting down.
	for m := range c.Bifrost.Rx {
		mbytes, err := m.Pack()
		if err != nil {
			c.sendError(ctx, errCh, err)
			continue
		}

		if _, err := c.Io.Write(mbytes); err != nil {
			c.sendError(ctx, errCh, err)
			break
		}
	}
}

// runTx runs the client's message transmitter loop.
func (c *Conn) runTx(ctx context.Context, errCh chan<- error) {
	r := message.NewReaderTokeniser(c.Io)

	for {
		if err := c.txLine(ctx, r); err != nil {
			c.sendError(ctx, errCh, err)
			return
		}
	}
}

// txLine transmits a line from the ReaderTokeniser r
func (c *Conn) txLine(ctx context.Context, r *message.ReaderTokeniser) (err error) {
	var line []string
	if line, err = r.ReadLine(); err != nil {
		return err
	}

	var msg *message.Message
	if msg, err = message.NewFromLine(line); err != nil {
		return err
	}

	if !c.Bifrost.Send(ctx, *msg) {
		return errors.New("client died while sending message on %s")
	}

	return nil
}

// sendError tries to send an error e to the error channel errCh.
// It silently fails if the underlying Client's Done channel is closed.
func (c *Conn) sendError(ctx context.Context, errCh chan<- error, e error) {
	done := ctx.Done()
	select {
	case errCh <- e:
	case <-done:
	}
}
