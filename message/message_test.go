package message

import "testing"
import "reflect"

// It feels like there should be more tests here, but since Message is
// essentially just a []string, why bother?

func TestPack(t *testing.T) {
	cases := []struct {
		msg  Message
		want []byte
	}{
		// Read helper func, Unescaped command
		{
			Read("uuid", "/player/file"),
			[]byte("read uuid /player/file\n"),
		},
		// Write helper func, Backslashes
		{
			Write("uuid", "/player/file", `C:\silly\windows\is\silly`),
			[]byte(`write uuid /player/file 'C:\silly\windows\is\silly'` + "\n"),
		},
		// Delete helper func
		{
			Delete("uuid", "/player/file"),
			[]byte("delete uuid /player/file\n"),
		},
		// Spaces
		{
			Write("uuid", "/player/file", "/home/donald/01 The Nightfly.mp3"),
			[]byte("write uuid /player/file '/home/donald/01 The Nightfly.mp3'\n"),
		},
		// Single quotes
		{
			Message{RsOhai, "a'bar'b"},
			[]byte(`OHAI 'a'\''bar'\''b'` + "\n"),
		},
		// Double quotes
		{
			Message{RsOhai, `a"bar"b`},
			[]byte(`OHAI 'a"bar"b'` + "\n"),
		},
		// Single word (shouldn't ever be used)
		{
			Message{RsOhai},
			[]byte("OHAI\n"),
		},
	}

	for _, c := range cases {
		got := c.msg.Pack()
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("Message.Pack(%q) == %q, want %q", c.msg, got, c.want)
		}
	}
}
