package ports

import (
	"context"
	"io"
)

// Storage writes uploaded image bytes to the durable image tree. The serve path reads the file
// directly through nginx, so no read method is needed here. The file is named <id>.<extension> so
// nginx derives the Content-Type from the extension under the global X-Content-Type-Options: nosniff
// (a bare id would be served as octet-stream and refused by the browser). Put returns bytes written.
//
// Delete removes a stored object by its key. The uploader calls it to compensate a failed row insert
// after a successful Put, so a failed upload does not orphan bytes (LB09a). It is not a general
// deletion API: the id is always one the caller just minted and never referenced from a row.
type Storage interface {
	Put(ctx context.Context, id, extension string, r io.Reader) (int64, error)
	Delete(ctx context.Context, id, extension string) error
}
