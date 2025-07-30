package utils

import "io"

type ReadSeekWrapper struct {
	io.ReadSeeker
	io.Closer
}
