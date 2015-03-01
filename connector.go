package baps3

import (
	"bufio"
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
	buf       *bufio.Reader
	resCh     chan<- Message
	ReqCh     chan Message
	name      string
	wg        *sync.WaitGroup
	logger    *log.Logger
}

// InitConnector creates and returns a Connector.
// The returned Connector shall have the given name, send responses through the
// response channel resCh, report termination via the wait group waitGroup, and
// log to logger.
func InitConnector(name string, resCh chan Message, waitGroup *sync.WaitGroup, logger *log.Logger) *Connector {
	c := new(Connector)
	c.tokeniser = NewTokeniser()
	c.resCh = resCh
	c.ReqCh = make(chan Message)
	c.name = name
	c.wg = waitGroup
	c.logger = logger
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
	c.buf = bufio.NewReader(c.conn)
}

// Run is the main connector loop, reading bytes off the wire, tokenising and handling
// responses.
func (c *Connector) Run() {
	lineCh := make(chan [][]string, 3)
	errCh := make(chan error)

	// Spin up a goroutine to accept and tokenise incoming bytes, and spit them
	// out in a channel
	go func(lineCh chan [][]string, eCh chan error) {
		for {
			data, err := c.buf.ReadBytes('\n')
			if err != nil {
				errCh <- err
			}
			// TODO(CaptainHayashi): more robust handling of an
			// error from Tokenise?
			lines, _, err := c.tokeniser.Tokenise(data)
			if err != nil {
				errCh <- err
			}
			lineCh <- lines
		}
	}(lineCh, errCh)

	// Main run loop, select on new received lines, errors or incoming requests
	for {
		select {
		case lines := <-lineCh:
			c.handleResponses(lines)
		case err := <-errCh:
			c.logger.Fatal(err)
		case req, ok := <-c.ReqCh:
			if !ok { // Other end closed the channel, shut down
				c.logger.Println(c.name + " Connector shutting down")
				err := c.conn.Close()
				if err != nil {
					c.logger.Println(err)
				}
				c.wg.Done()
				return
			}
			data, err := req.Pack()
			if err != nil {
				c.logger.Println(err)
			} else {
				c.conn.Write(data)
			}
		}
	}
}

// handleResponses handles a series of response lines from the BAPS3 server.
func (c *Connector) handleResponses(lines [][]string) {
	for _, line := range lines {
		msg, err := LineToMessage(line)
		if err != nil {
			c.logger.Println(err)
			continue
		}

		if msg.Word().IsUnknown() {
			continue
		}

		c.resCh <- *msg
	}
}

// PrettyDuration pretty-prints a duration in the form minutes:seconds.
// The seconds part is zero-padded; the minutes part is not.
func PrettyDuration(dur time.Duration) string {
	return fmt.Sprintf("%d:%02d", int(dur.Minutes()), int(math.Mod(dur.Seconds(), 60)))
}
