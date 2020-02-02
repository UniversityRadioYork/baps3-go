package comm

import (
	"context"
	"github.com/UniversityRadioYork/bifrost-go/core"
	"github.com/UniversityRadioYork/bifrost-go/message"
	"net"
)

// Client is a wrapper around an IoEndpoint that stores various pieces of information about a Bifrost server.
type Client struct {
	// ServerVer stores the server version of the client.
	ServerVer string

	// Role stores the initial role of the client.
	Role string

	// Endpoint is the raw message-based endpoint that can be used to interact with this client's server.
	Endpoint Endpoint

	// ServerIo represents the connection to the external server.
	ServerIo IoEndpoint
}

// Dial connects to a Bifrost server at address, and, if successful, constructs a new ExternalService over it.
func Dial(ctx context.Context, address string, errCh chan<- error) (c *Client, err error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	cliEnd, srvEnd := NewEndpointPair()
	ioEnd := IoEndpoint{Endpoint: srvEnd, Io: conn}
	ioEnd.Run(ctx, errCh)
	return NewClient(ctx, cliEnd, ioEnd)
}

// NewClient tries to spin up a Client connected to a Bifrost server through serverIo.
// It expects the endpoint to be running its loops.
func NewClient(ctx context.Context, cliEnd *Endpoint, serverIo IoEndpoint) (*Client, error) {
	serverVer, role, err := handshake(ctx, cliEnd)
	if err != nil {
		return nil, err
	}

	c := &Client{ServerVer: serverVer, Role: role, ServerIo: serverIo}
	return c, nil
}

// handshake performs the Bifrost handshake with whichever Bifrost service is on the other end of cliEnd.
func handshake(ctx context.Context, cliEnd *Endpoint) (serverVer, role string, err error) {
	// TODO(@MattWindsor91): make this more symmetric with the way it's done on the client side
	if serverVer, err = recvOhai(ctx, cliEnd); err != nil {
		return "", "", err
	}
	if role, err = recvIama(ctx, cliEnd); err != nil {
		return "", "", err
	}
	return serverVer, role, nil
}

func recvOhai(ctx context.Context, cliEnd *Endpoint) (serverVer string, err error) {
	var (
		ohaiMsg *message.Message
		ohai    *core.OhaiResponse
	)
	if ohaiMsg, err = cliEnd.Recv(ctx); err != nil {
		return "", err
	}
	if ohai, err = core.ParseOhaiResponse(ohaiMsg); err != nil {
		return "", err
	}
	// TODO(@MattWindsor91): check protocol version
	return ohai.ServerVer, nil
}

func recvIama(ctx context.Context, cliEnd *Endpoint) (role string, err error) {
	var (
		iamaMsg *message.Message
		iama    *core.IamaResponse
	)
	if iamaMsg, err = cliEnd.Recv(ctx); err != nil {
		return "", err
	}
	if iama, err = core.ParseIamaResponse(iamaMsg); err != nil {
		return "", err
	}
	return iama.Role, nil
}
