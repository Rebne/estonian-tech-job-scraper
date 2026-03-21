package fetcher

import (
	"context"
	"io"
	"net/http"
)

type HTTPFetcher struct{}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{}
}

func (hf *HTTPFetcher) Fetch(ctx context.Context, url string) (string, error) {
	return FetchHTMLViaHTTP(ctx, url)
}

func (hf *HTTPFetcher) Close() error {
	return nil
}

func FetchHTMLViaHTTP(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
