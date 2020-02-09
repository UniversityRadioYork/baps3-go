package list

// ItemType is the type of types of item.
type ItemType int

const (
	// ItemNone represents a nonexistent item.
	ItemNone ItemType = iota
	// ItemTrack represents a track item.
	// Track items can be selected.
	ItemTrack
	// ItemText represents a textual item.
	// Text items cannot be selected.
	ItemText
)

// String gets the descriptive name of an ItemType as a string.
func (i ItemType) String() string {
	switch i {
	case ItemNone:
		return "none"
	case ItemTrack:
		return "track"
	case ItemText:
		return "text"
	default:
		return "?unknown?"
	}
}

// Item represents a baps3d list item.
type Item struct {
	// Payload is the data component of the item.
	Payload string
	// Type is the type of the item.
	Type ItemType
}
