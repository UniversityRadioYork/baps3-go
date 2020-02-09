package list

// AutoModeResponse announces a change in AutoMode.
type AutoModeResponse struct {
	// AutoMode represents the new AutoMode.
	AutoMode AutoMode
}

// SelectResponse announces a change in selection.
type SelectResponse struct {
	// Index represents the selected index.
	Index Index
}

// ItemResponse announces the presence of a single list item.
type ItemResponse struct {
	// Index is the index of the item in the list.
	Index Index

	// Item is the item itself.
	Item Item
}
