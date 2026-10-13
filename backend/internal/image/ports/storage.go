package ports

import (
	"context"
	"io"
)

// Storage writes uploaded image bytes to the durable image tree. The serve path reads the file
// directly through nginx, so no read method is needed here. The file is named <id>.<extension> so
// nginx derives the Content-Type from the extension under the global X-Content-Type-Options: nosniff
// (a bare id would be served as octet-stream and refused by the browser). Put returns bytes written.
type Storage interface {
	Put(ctx context.Context, id, extension string, r io.Reader) (int64, error)
}
