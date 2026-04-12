package fetcher

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	internalerrors "github.com/Rebne/scrapy_project_v2/internal/errors"
	"github.com/playwright-community/playwright-go"
)

const playwrightFetchTimeoutMilliseconds = 60_000

var errPlaywrightFetcherClosed = errors.New("playwright fetcher closed")

type PlaywrightFetcher struct {
	once      sync.Once
	closeOnce sync.Once
	proxyURL  string

	mu      sync.RWMutex
	closed  bool
	initErr error
	pw      *playwright.Playwright
	browser playwright.Browser
}

type PlaywrightFetcherOptions struct {
	ProxyURL string
}

func NewPlaywrightFetcher(options PlaywrightFetcherOptions) (*PlaywrightFetcher, error) {
	proxyURL := options.ProxyURL
	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		if parsedProxyURL.Scheme == "" || parsedProxyURL.Host == "" {
			return nil, fmt.Errorf("missing scheme or host in proxyUrl")
		}
	}

	return &PlaywrightFetcher{proxyURL: proxyURL}, nil
}

func (pf *PlaywrightFetcher) initBrowser() {
	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	}

	if pf.proxyURL != "" {
		proxy, err := newPlaywrightProxy(pf.proxyURL)
		if err != nil {
			pf.mu.Lock()
			pf.initErr = err
			pf.mu.Unlock()
			return
		}
		launchOptions.Proxy = proxy
	}

	pw, err := playwright.Run()
	if err != nil {
		pf.mu.Lock()
		pf.initErr = fmt.Errorf("starting playwright failed: %w", err)
		pf.mu.Unlock()
		return
	}

	browser, err := pw.Chromium.Launch(launchOptions)
	if err != nil {
		_ = pw.Stop()
		pf.mu.Lock()
		pf.initErr = fmt.Errorf("launching playwright browser failed: %w", err)
		pf.mu.Unlock()
		return
	}

	pf.mu.Lock()
	pf.pw = pw
	pf.browser = browser
	if pf.closed {
		_ = browser.Close()
		_ = pw.Stop()
		pf.initErr = errPlaywrightFetcherClosed
	}
	pf.mu.Unlock()
}

func (pf *PlaywrightFetcher) ensureReady() error {
	pf.once.Do(pf.initBrowser)

	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if pf.closed {
		return errPlaywrightFetcherClosed
	}

	return pf.initErr
}

func (pf *PlaywrightFetcher) Fetch(ctx context.Context, targetURL string) (string, error) {
	if err := pf.ensureReady(); err != nil {
		return "", err
	}

	pf.mu.RLock()
	browser := pf.browser
	pf.mu.RUnlock()

	browserContext, err := browser.NewContext()
	if err != nil {
		return "", fmt.Errorf("creating playwright context failed: %w", err)
	}
	defer browserContext.Close()

	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("creating playwright page failed: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		page.SetDefaultNavigationTimeout(float64(timeUntilMilliseconds(deadline)))
	}

	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(playwrightFetchTimeoutMilliseconds),
	})
	if err != nil {
		if errors.Is(err, playwright.ErrTimeout) {
			return "", fmt.Errorf("%w: %w", internalerrors.ErrPlaywrightTimeout, err)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("playwright goto failed: %w", err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle}); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("waiting for playwright networkidle failed: %w", err)
	}

	html, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("reading playwright page content failed: %w", err)
	}

	return html, nil
}

func (pf *PlaywrightFetcher) Close() error {
	var closeErr error

	pf.closeOnce.Do(func() {
		pf.mu.Lock()
		defer pf.mu.Unlock()

		pf.closed = true
		if pf.browser != nil {
			closeErr = errors.Join(closeErr, pf.browser.Close())
		}
		if pf.pw != nil {
			closeErr = errors.Join(closeErr, pf.pw.Stop())
		}
	})

	return closeErr
}

func newPlaywrightProxy(rawProxyURL string) (*playwright.Proxy, error) {
	parsedProxyURL, err := url.Parse(rawProxyURL)
	if err != nil {
		return nil, err
	}

	proxy := &playwright.Proxy{
		Server: fmt.Sprintf("%s://%s", parsedProxyURL.Scheme, parsedProxyURL.Host),
	}

	if parsedProxyURL.User != nil {
		proxy.Username = playwright.String(parsedProxyURL.User.Username())
		password, hasPassword := parsedProxyURL.User.Password()
		if hasPassword {
			proxy.Password = playwright.String(password)
		}
	}

	return proxy, nil
}

func timeUntilMilliseconds(deadline time.Time) int {
	remaining := time.Until(deadline).Milliseconds()
	if remaining <= 0 {
		return 1
	}

	return int(remaining)
}
