package fetcher

import "context"

type HTMLRetriever interface {
	Fetch(context.Context, string) (string, error)
	Close() error
}
