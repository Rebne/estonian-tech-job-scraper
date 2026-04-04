package fetcher

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/chromedp/chromedp"
)

var errChromeFetcherClosed = errors.New("chrome fetcher closed")

type ChromeFetcher struct {
	once      sync.Once
	closeOnce sync.Once

	mu            sync.RWMutex
	closed        bool
	initErr       error
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	browserCtx    context.Context
	browserCancel context.CancelFunc
}

func NewChromeFetcher() (*ChromeFetcher, error) {
	return &ChromeFetcher{}, nil
}

func (cf *ChromeFetcher) initBrowser() {
	allocatorOptions := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1366,768"),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"),
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	cf.mu.Lock()
	cf.allocCtx = allocCtx
	cf.allocCancel = allocCancel
	cf.browserCtx = browserCtx
	cf.browserCancel = browserCancel
	if cf.closed {
		browserCancel()
		allocCancel()
		cf.initErr = errChromeFetcherClosed
	}
	cf.mu.Unlock()
}

func (cf *ChromeFetcher) ensureReady() error {
	cf.once.Do(cf.initBrowser)

	cf.mu.RLock()
	defer cf.mu.RUnlock()

	if cf.closed {
		return errChromeFetcherClosed
	}

	return cf.initErr
}

func (cf *ChromeFetcher) Fetch(ctx context.Context, url string) (string, error) {
	if err := cf.ensureReady(); err != nil {
		return "", err
	}

	cf.mu.RLock()
	browserCtx := cf.browserCtx
	cf.mu.RUnlock()

	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	runCtx, runCancel := context.WithCancel(tabCtx)
	defer runCancel()

	go func() {
		select {
		case <-ctx.Done():
			runCancel()
		case <-runCtx.Done():
		}
	}()

	var html string
	err := chromedp.Run(runCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("chromedp run ended with an error: %w", err)
	}

	return html, nil
}

func (cf *ChromeFetcher) Close() error {
	cf.closeOnce.Do(func() {
		cf.mu.Lock()
		defer cf.mu.Unlock()

		cf.closed = true
		if cf.browserCancel != nil {
			cf.browserCancel()
		}
		if cf.allocCancel != nil {
			cf.allocCancel()
		}
	})

	return nil
}
