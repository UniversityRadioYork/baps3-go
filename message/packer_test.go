package message

import "testing"

func cmpByteSlices(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, abyte := range a {
		if abyte != b[i] {
			return false
		}
	}

	return true
}

func TestPack(t *testing.T) {
	cases := []struct {
		word string
		args []string
		want []byte
	}{
		// Unescaped command
		{
			"load",
			[]string{"/home/donald/wjaz.mp3"},
			[]byte("load /home/donald/wjaz.mp3\n"),
		},
		// Backslashes
		{
			"load",
			[]string{`C:\silly\windows\is\silly`},
			[]byte(`load 'C:\silly\windows\is\silly'` + "\n"),
		},
		// No args
		{
			"play",
			[]string{},
			[]byte("play\n"),
		},
		// Spaces
		{
			"load",
			[]string{"/home/donald/01 The Nightfly.mp3"},
			[]byte("load '/home/donald/01 The Nightfly.mp3'\n"),
		},
		// Single quotes
		{
			"foo",
			[]string{"a'bar'b"},
			[]byte(`foo 'a'\''bar'\''b'` + "\n"),
		},
		// Double quotes
		{
			"foo",
			[]string{`a"bar"b`},
			[]byte(`foo 'a"bar"b'` + "\n"),
		},
	}

	for _, c := range cases {
		got, err := Pack(c.word, c.args)
		if err != nil {
			t.Errorf("Pack(%q, %q) encountered error %q", c.word, c.args, err)
		}
		if !cmpByteSlices(c.want, got) {
			t.Errorf("Pack(%q, %q) == %q, want %q", c.word, c.args, got, c.want)
		}
	}
}
