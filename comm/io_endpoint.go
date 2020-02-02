package comm

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/UniversityRadioYork/bifrost-go/message"
)

// HungUpError is the error sent by an IoEndpoint when its transmission loop has hung up.
var HungUpError = errors.New("hung up")

// IoEndpoint represents a Bifrost endpoint that sends and receives messages along an I/O connection.
type IoEndpoint struct {
	// Io holds the internal I/O connection.
	Io io.ReadWriteCloser

	// Bifrost holds the Bifrost channel pair used by the Io.
	Endpoint *Endpoint
}

func (e *IoEndpoint) Close() error {
	// TODO(@MattWindsor91): make sure we close everything
	close(e.Endpoint.Tx)
	return e.Io.Close()
}

// Run spins up the client's receiver and transmitter loops.
// It takes a channel to notify the caller asynchronously of any errors, and a client
// and the server's client hangup and done channels.
// It closes errors once both loops are done.
func (e *IoEndpoint) Run(ctx context.Context, errCh chan<- error) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		e.runTx(ctx, errCh)
		e.sendError(ctx, errCh, HungUpError)
		wg.Done()
	}()

	go func() {
		e.runRx(ctx, errCh)
		wg.Done()
	}()

	wg.Wait()
	close(errCh)
}

// runRx runs the client's message receiver loop.
// This writes messages to the socket.
func (e *IoEndpoint) runRx(ctx context.Context, errCh chan<- error) {
	// We don't have to check e.Bifrost.Done here:
	// client always drops both Rx and Done when shutting down.
	for m := range e.Endpoint.Rx {
		mbytes, err := m.Pack()
		if err != nil {
			e.sendError(ctx, errCh, err)
			continue
		}

		if _, err := e.Io.Write(mbytes); err != nil {
			e.sendError(ctx, errCh, err)
			break
		}
	}
}

// runTx runs the client's message transmitter loop.
func (e *IoEndpoint) runTx(ctx context.Context, errCh chan<- error) {
	r := message.NewReaderTokeniser(e.Io)

	for {
		if err := e.txLine(ctx, r); err != nil {
			e.sendError(ctx, errCh, err)
			return
		}
	}
}

// txLine transmits a line from the ReaderTokeniser r
func (e *IoEndpoint) txLine(ctx context.Context, r *message.ReaderTokeniser) (err error) {
	var line []string
	if line, err = r.ReadLine(); err != nil {
		return err
	}

	var msg *message.Message
	if msg, err = message.NewFromLine(line); err != nil {
		return err
	}

	if !e.Endpoint.Send(ctx, *msg) {
		return errors.New("client died while sending message on %s")
	}

	return nil
}

// sendError tries to send an error e to the error channel errCh.
// It silently fails if the underlying Client's Done channel is closed.
func (e *IoEndpoint) sendError(ctx context.Context, errCh chan<- error, err error) {
	done := ctx.Done()
	select {
	case errCh <- err:
	case <-done:
	}
}
