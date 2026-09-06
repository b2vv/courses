package ports

import "context"

type TranscodingQueue interface {
	Enqueue(context.Context, string, int) error
}
