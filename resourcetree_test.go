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
	want := DirectoryResourceNode(make(map[string]ResourceNoder))
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
			NewEntryResourceNode(BifrostTypeString("lol")),
			"/foo/bar/baz",
			DirectoryResourceNode(map[string]ResourceNoder{
				"foo": DirectoryResourceNode(map[string]ResourceNoder{
					"bar": DirectoryResourceNode(map[string]ResourceNoder{
						"baz": NewEntryResourceNode(BifrostTypeString("lol")),
					}),
				}),
			}),
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

// TestResourcify ensures that the stock resource tree nodes are giving us
// decent []Resource lists.
func TestResourcify(t *testing.T) {
	cases := []struct {
		have ResourceNoder
		want []Resource
	}{
		{
			NewEntryResourceNode(BifrostTypeString("fus ro dah")),
			[]Resource{
				Resource{[]string{}, BifrostTypeString("fus ro dah")},
			},
		},
		{
			NewEntryResourceNode(BifrostTypeInt(8675309)),
			[]Resource{
				Resource{[]string{}, BifrostTypeInt(8675309)},
			},
		},
		{
			NewDirectoryResourceNode(
				map[string]ResourceNoder{
					"we're":  NewEntryResourceNode(BifrostTypeString("only")),
					"making": NewEntryResourceNode(BifrostTypeString("plans")),
					"for":    NewEntryResourceNode(BifrostTypeString("Nigel")),
				},
			),
			[]Resource{
				// 3 entries
				Resource{[]string{}, BifrostTypeDirectory{3}},
				Resource{[]string{"we're"}, BifrostTypeString("only")},
				Resource{[]string{"making"}, BifrostTypeString("plans")},
				Resource{[]string{"for"}, BifrostTypeString("Nigel")},
			},
		},
	}

	for _, c := range cases {
		got := ToResource([]string{}, c.have)

		if len(got) != len(c.want) {
			t.Fatalf("bad resourcify: have %q, got %q, want %q", c.have, got, c.want)
		}

		// Slightly nasty own-rolled loop, as DeepEqual tests order of slice
		// as well, which we can't guarantee.
		// TODO: Check for duplicates
		for i := range got {
			equal := false
			for j := range c.want {
				if reflect.DeepEqual(got[i], c.want[j]) {
					equal = true
					break
				}
			}
			if !equal {
				t.Fatalf("bad resourcify: have %q, got %q, want %q", c.have, got, c.want)
			}
		}
	}
}
