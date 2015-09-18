package bifrost

import (
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

// Connector is a struct containing the internal state of a BAPS3 connector.
type Connector struct {
	state     string
	time      time.Duration
	tokeniser *Tokeniser
	conn      net.Conn
	resCh     chan<- Message
	ReqCh     chan Message
	name      string
	logger    *log.Logger
	quit      chan struct{}
	wg        *sync.WaitGroup
}

// InitConnector creates and returns a Connector.
// The returned Connector shall have the given name, send responses through the
// response channel resCh, report termination via the wait group waitGroup, and
// log to logger.
func InitConnector(name string, resCh chan Message, logger *log.Logger) *Connector {
	c := new(Connector)
	c.resCh = resCh
	c.ReqCh = make(chan Message)
	c.name = name
	c.logger = logger
	c.quit = make(chan struct{})
	c.wg = new(sync.WaitGroup)
	return c
}

// Connect connects an existing Connector to the BAPS3 server at hostport (in
// the format host:port).
func (c *Connector) Connect(hostport string) {
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		c.logger.Fatal(err)
	}
	c.conn = conn
	c.tokeniser = NewTokeniser(c.conn)
}

// Quit synchronously terminates the connector, gracefully disconnecting from downstream
func (c *Connector) Quit() {
	c.quit <- struct{}{}
	c.wg.Wait()
}

// Run is the main connector loop, reading bytes off the wire, tokenising and handling
// responses.
func (c *Connector) Run() {
	lineCh := make(chan []string, 3)
	errCh := make(chan error)

	// Spin up a goroutine to accept and tokenise incoming bytes, and spit them
	// out in a channel
	go func(lineCh chan []string, eCh chan error) {
		for {
			// TODO(CaptainHayashi): more robust handling of an
			// error from Tokenise?
			line, err := c.tokeniser.Tokenise()
			if err != nil {
				eCh <- err
			}
			lineCh <- line
		}
	}(lineCh, errCh)

	// Main run loop, select on new received lines, errors or incoming requests
	c.wg.Add(1)
	for {
		select {
		case line := <-lineCh:
			if err := c.handleResponse(line); err != nil {
				c.logger.Println(err)
			}
		case err := <-errCh:
			c.logger.Fatal(err)
		case req := <-c.ReqCh:
			data, err := req.Pack()
			if err != nil {
				c.logger.Println(err)
			} else {
				c.conn.Write(data)
			}
		case <-c.quit:
			c.logger.Println(c.name + " Connector shutting down")
			err := c.conn.Close()
			if err != nil {
				c.logger.Println(err)
			}
			c.wg.Done()
			return
		}
	}
}

// handleResponses handles a response line from the BAPS3 server.
func (c *Connector) handleResponse(line []string) error {
	msg, err := LineToMessage(line)
	if err != nil {
		return err
	}

	if !msg.Word().IsUnknown() {
		c.resCh <- *msg
	}

	return nil
}

// PrettyDuration pretty-prints a duration in the form minutes:seconds.
// The seconds part is zero-padded; the minutes part is not.
func PrettyDuration(dur time.Duration) string {
	return fmt.Sprintf("%d:%02d", int(dur.Minutes()), int(math.Mod(dur.Seconds(), 60)))
}
