package comm

import (
	"fmt"
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
	case core.RsAck:
		return core.ParseAckResponse(m)
	case core.RsIama:
		return core.ParseIamaResponse(m)
	case core.RsOhai:
		return core.ParseOhaiResponse(m)
	}
	return nil, fmt.Errorf("unknown word: %s", m.Word())
}
