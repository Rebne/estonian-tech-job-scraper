package fetcher

import "context"

type HTMLRetriever func(context context.Context, url string) (string, error)
