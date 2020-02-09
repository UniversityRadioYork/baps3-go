package list

// Index represents a Bifrost list index: a pair of list position and inserter-chosen 'hash' string.
type Index struct {
	// Position represents the physical position of the item in the list.
	Position int

	// Hash represents the 'hash' of the item being indexed.
	// While its exact contents are up to the inserter, it should serve to prevent selection races.
	Hash string
}
