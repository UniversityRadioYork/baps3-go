package list

// SetAutoModeRequest requests an automode change.
type SetAutoModeRequest struct {
	// AutoMode represents the new AutoMode to use.
	AutoMode AutoMode
}

// SetSelectRequest requests a selection change.
type SetSelectRequest struct {
	// Index represents the index to select.
	Index Index
}

// AddItemRequest requests that the given item be enqueued in front of the given index.
type AddItemRequest struct {
	// Index is the index at which we want to enqueue this item.
	Index Index

	// Item is the item itself.
	Item Item
}
