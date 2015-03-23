package baps3

import (
	"testing"
)

func TestLookupFeature(t *testing.T) {
	cases := []struct {
		featstr string
		want    Feature
	}{
		{
			"PlayStop",
			FtPlayStop,
		},
		{
			"ImATeapot",
			FtUnknown,
		},
		{
			"Playlist.TextItems",
			FtPlaylistTextItems,
		},
	}

	for _, c := range cases {
		got := LookupFeature(c.featstr)
		if got != c.want {
			t.Errorf("LookupFeature(%q) == %q, want %q", c.featstr, got, c.want)
		}
		if got.String() != c.want.String() {
			t.Errorf("Equivalent feature string mismatch, %q != %q", got.String(), c.want.String())
		}
	}
}
