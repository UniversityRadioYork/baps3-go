package comm

import (
	"fmt"

	"github.com/UniversityRadioYork/bifrost-go/role/list"

	"github.com/UniversityRadioYork/bifrost-go/core"
	"github.com/UniversityRadioYork/bifrost-go/message"
)

// Messager is the type of things that can be converted into a message.
type Messager interface {
	Message(tag string) *message.Message
}

// ParseMessage tries to understand message m as a known Bifrost message.
func ParseMessage(m *message.Message) (Messager, error) {
	switch m.Word() {
	// core
	case core.RsAck:
		return core.ParseAckResponse(m)
	case core.RsIama:
		return core.ParseIamaResponse(m)
	case core.RsOhai:
		return core.ParseOhaiResponse(m)
	// list
	case list.RsCountL:
		return list.ParseCountLResponse(m)
	}
	return nil, fmt.Errorf("unknown word: %s", m.Word())
}

// ReadMessage reads a line from tokeniser r, then converts it to a Message.
func ReadMessage(r *message.Reader) (*message.Message, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	return message.NewFromLine(line)
}

// ReadAndParse reads a line from tokeniser r, converts it to a Message, then parses it.
func ReadAndParse(r *message.Reader) (Messager, error) {
	m, err := ReadMessage(r)
	if err != nil {
		return nil, err
	}
	return ParseMessage(m)
}
