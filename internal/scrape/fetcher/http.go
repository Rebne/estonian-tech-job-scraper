package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type HTTPFetcher struct {
	client *http.Client
}

type HTTPFetcherOptions struct {
	ProxyURL string
}

func NewHTTPFetcher(options HTTPFetcherOptions) (*HTTPFetcher, error) {
	proxyURL := options.ProxyURL
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		if parsedProxyURL.Scheme == "" || parsedProxyURL.Host == "" {
			return nil, fmt.Errorf("missing scheme or host in proxyUrl")
		}

		transport.Proxy = http.ProxyURL(parsedProxyURL)
	}

	return &HTTPFetcher{
		client: &http.Client{Transport: transport},
	}, nil
}

func (hf *HTTPFetcher) Fetch(ctx context.Context, url string) (string, error) {
	return hf.fetchHTMLViaHTTP(ctx, url)
}

func (hf *HTTPFetcher) Close() error {
	return nil
}

func (hf *HTTPFetcher) fetchHTMLViaHTTP(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")

	resp, err := hf.client.Do(req)
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
