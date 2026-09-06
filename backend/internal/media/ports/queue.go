package ports

import "context"

type Job struct {
	AssetID    string
	Attempt    int
	RetryCount int
	MaxRetry   int
}

type TranscodingQueue interface {
	Enqueue(context.Context, string, int) error
}
