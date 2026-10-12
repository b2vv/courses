package ports

import (
	"context"
	"io"
)

// Storage writes uploaded image bytes to the durable image tree. The serve path reads the file
// directly through nginx, so no read method is needed here. Put returns the number of bytes written.
type Storage interface {
	Put(ctx context.Context, id string, r io.Reader) (int64, error)
}
