package comm

import (
	"github.com/UniversityRadioYork/bifrost-go/core"
	"github.com/UniversityRadioYork/bifrost-go/message"
	"reflect"
	"testing"
)

var exampleMessageTable = []struct {
	message *message.Message
	want    Messager
}{
	{message.New(message.TagBcast, core.RsAck).AddArgs(core.WordWhat, "description here"),
		&core.AckResponse{
			Status:      core.StatusWhat,
			Description: "description here",
		},
	},
	{
		message.New(message.TagBcast, core.RsIama).AddArgs("player/file"),
		&core.IamaResponse{Role: "player/file"},
	},
	{message.New(message.TagBcast, core.RsOhai).AddArgs("test-0.2.0", "example-42.0.0"),
		&core.OhaiResponse{
			ProtocolVer: "test-0.2.0",
			ServerVer:   "example-42.0.0",
		},
	},
}

// TestParseMessage_Valid tests ParseMessage on various valid messages.
func TestParseMessage_Valid(t *testing.T) {
	for _, c := range exampleMessageTable {
		if got, err := ParseMessage(c.message); err != nil {
			t.Errorf("unexpected parse error on %q: %v", c.message.Word(), err)
		} else if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parse %q=%v; want %v", c.message.Word(), got, c.want)
		}
	}
}
