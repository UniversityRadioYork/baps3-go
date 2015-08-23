package bifrost

import (
	"reflect"
	"testing"
)

func TestSplitPath(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{
			"/",
			[]string{},
		},
		{
			"reelin/in/the/years",
			[]string{
				"reelin",
				"in",
				"the",
				"years",
			},
		},
		{
			"/stowing/away/the/time",
			[]string{
				"stowing",
				"away",
				"the",
				"time",
			},
		},
		{
			"gatherin/all/the/fears/",
			[]string{
				"gatherin",
				"all",
				"the",
				"fears",
			},
		},
		{
			"//had//enough/of////mine///",
			[]string{
				"had",
				"enough",
				"of",
				"mine",
			},
		},
	}

	for _, c := range cases {
		got := splitPath(c.path)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitPath(%q) == %q, want %q", c.path, got, c.want)
		}
	}
}

func TestNew(t *testing.T) {
	got := NewDirectoryResourceNode(make(map[string]ResourceNoder))
	want := &DirectoryResourceNode{
		ResourceNode{},
		make(map[string]ResourceNoder),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewDirectoryResourceNode == %q want %q", got, want)
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		have      ResourceNoder
		toadd     ResourceNoder
		path      string
		want      ResourceNoder
		shouldErr bool
	}{
		{
			NewDirectoryResourceNode(make(map[string]ResourceNoder)),
			// TODO(CaptainHayashi): bodge job to plaster over missing type function
			NewEntryResourceNode(ToResource([]string{}, "lol")[0].value),
			"/foo/bar/baz",
			&DirectoryResourceNode{
				ResourceNode{},
				map[string]ResourceNoder{
					"foo": &DirectoryResourceNode{
						ResourceNode{},
						map[string]ResourceNoder{
							"bar": &DirectoryResourceNode{
								ResourceNode{},
								map[string]ResourceNoder{
									"baz": &EntryResourceNode{
										ResourceNode{},
										BifrostTypeString("lol"),
									},
								},
							},
						},
					},
				},
			},
			false,
		},
	}

	for _, c := range cases {
		err := Add(c.have, c.path, c.toadd)
		if (err != nil) != c.shouldErr {
			if err != nil {
				t.Errorf("Add errored when it shouldn't have, got: %q", err)
			} else {
				t.Errorf("Add didn't error when it should have")
			}
		}
		if !reflect.DeepEqual(c.have, c.want) {
			t.Errorf("Add: got %q, want %q", c.have, c.want)
		}
	}
}
