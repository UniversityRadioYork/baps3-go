package baps3

import (
	"reflect"
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

func TestFeatureSetFromMsg(t *testing.T) {
	cases := []struct {
		msg         *Message
		want        FeatureSet
		shoulderror bool
	}{
		// Test error on not a features message
		{
			NewMessage(RsTime).AddArg("Ceci n'est pas une caractéristique"),
			nil,
			true,
		},
		// Test general functionality
		{
			NewMessage(RsFeatures).AddArg("PlayStop").AddArg("End").AddArg("FileLoad"),
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}, FtFileLoad: struct{}{}},
			false,
		},
		// Test duplicated args
		{
			NewMessage(RsFeatures).AddArg("PlayStop").AddArg("PlayStop"),
			FeatureSet{FtPlayStop: struct{}{}},
			false,
		},
		// Test error on unknown feature
		{
			NewMessage(RsFeatures).AddArg("JimmyRustle").AddArg("PlayStop"),
			FeatureSet{FtUnknown: struct{}{}, FtPlayStop: struct{}{}},
			true,
		},
	}

	for _, c := range cases {
		got, err := FeatureSetFromMsg(c.msg)
		if c.shoulderror != (err != nil) {
			if err != nil {
				t.Errorf("FeatureSetFromMsg(%q) returned err when should be nil(%s)", c.msg, err.Error())
			} else {
				t.Errorf("FeatureSetFromMsg(%q) returned nil when should be err", c.msg)
			}
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("FeatureSetFromMsg(%q) == %q, want %q", c.msg, got, c.want)
		}
	}
}

func TestToMessage(t *testing.T) {
	cases := []struct {
		fs   FeatureSet
		want *Message
	}{
		// N.B. the wanted message needs to have arguments in alphabetical order, as ToMessage output is sorted
		{
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}},
			NewMessage(RsFeatures).AddArg("End").AddArg("PlayStop"),
		},
		{
			FeatureSet{},
			NewMessage(RsFeatures),
		},
	}

	for _, c := range cases {
		got := c.fs.ToMessage()
		if got.String() != c.want.String() {
			t.Errorf("%q.ToMessage() == %q, want %q", c.fs, got, c.want)
		}
	}
}

func TestAddDelFeature(t *testing.T) {
	cases := []struct {
		fs   FeatureSet
		want FeatureSet
	}{
		// Test add
		{
			FeatureSet{}.AddFeature(FtPlayStop).AddFeature(FtEnd),
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}},
		},
		// Test duplicate adds
		{
			FeatureSet{}.AddFeature(FtPlayStop).AddFeature(FtPlayStop),
			FeatureSet{FtPlayStop: struct{}{}},
		},
		// Test Delete
		{
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}}.DelFeature(FtPlayStop),
			FeatureSet{FtEnd: struct{}{}},
		},
		// Test duplicate delete
		{
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}}.DelFeature(FtPlayStop).DelFeature(FtPlayStop),
			FeatureSet{FtEnd: struct{}{}},
		},
		// Test delete nonexistant
		{
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}}.DelFeature(FtFileLoad),
			FeatureSet{FtPlayStop: struct{}{}, FtEnd: struct{}{}},
		},
	}

	for _, c := range cases {
		if !reflect.DeepEqual(c.fs, c.want) {
			t.Errorf("TestAddDelFeature: %q != %q", c.fs, c.want)
		}
	}
}
