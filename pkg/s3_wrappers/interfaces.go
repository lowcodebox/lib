package s3_wrappers

import (
	"context"
	"io"
	"net/url"
	"time"
)

// Location represents a storage location.
type Location interface {
	io.Closer
	// CreateContainer creates a new Container with the
	// specified name.
	CreateContainer(ctx context.Context, name string) (Container, error)
	// Containers gets a page of containers
	// with the specified prefix from this Location.
	// The specified cursor is a pointer to the start of
	// the containers to get. It it obtained from a previous
	// call to this method, or should be CursorStart for the
	// first page.
	// count is the number of items to return per page.
	// The returned cursor can be checked with IsCursorEnd to
	// decide if there are any more items or not.
	Containers(ctx context.Context, prefix string, cursor string, count int) ([]Container, string, error)
	// Container gets the Container with the specified
	// identifier.
	Container(ctx context.Context, id string) (Container, error)
	// RemoveContainer removes the container with the specified ID.
	RemoveContainer(ctx context.Context, id string) error
	// ItemByURL gets an Item at this location with the
	// specified URL.
	ItemByURL(url *url.URL) (Item, error)
}

// Container represents a container.
type Container interface {
	// ID gets a unique string describing this Container.
	ID() string
	// Name gets a human-readable name describing this Container.
	Name() string
	// Item gets an item by its ID.
	Item(ctx context.Context, id string) (Item, error)
	// Items gets a page of items with the specified
	// prefix for this Container.
	// The specified cursor is a pointer to the start of
	// the items to get. It it obtained from a previous
	// call to this method, or should be CursorStart for the
	// first page.
	// count is the number of items to return per page.
	// The returned cursor can be checked with IsCursorEnd to
	// decide if there are any more items or not.
	Items(ctx context.Context, prefix, cursor string, count int) ([]Item, string, error)
	// RemoveItem removes the Item with the specified ID.
	RemoveItem(ctx context.Context, id string) error
	// Put creates a new Item with the specified name, and contents
	// read from the reader.
	Put(ctx context.Context, name string, r io.Reader, size int64, metadata map[string]interface{}) (Item, error)
}

// Item represents an item inside a Container.
// Such as a file.
type Item interface {
	// ID gets a unique string describing this Item.
	ID() string
	// Name gets a human-readable name describing this Item.
	Name() string
	// URL gets a URL for this item.
	// For example:
	// local: file:///path/to/something
	// azure: azure://host:port/api/something
	//    s3: s3://host:post/etc
	URL(ctx context.Context) (*url.URL, error)
	// Size gets the size of the Item's contents in bytes.
	Size(ctx context.Context) (int64, error)
	// Open opens the Item for reading.
	// Calling code must close the io.ReadCloser.
	Open(ctx context.Context) (io.ReadCloser, error)
	// ETag is a string that is different when the Item is
	// different, and the same when the item is the same.
	// Usually this is the last modified datetime.
	ETag(ctx context.Context) (string, error)
	// LastMod returns the last modified date of the file.
	LastMod(ctx context.Context) (time.Time, error)
	// Metadata gets a map of key/values that belong
	// to this Item.
	Metadata(ctx context.Context) (map[string]interface{}, error)
}
