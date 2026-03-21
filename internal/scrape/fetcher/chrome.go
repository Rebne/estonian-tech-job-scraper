package fetcher

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)


func FetchHTMLViaChrome(ctx context.Context, url string) (string, error) {
	ctx, cancel := chromedp.NewContext(ctx)
		defer cancel()

		var html string

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.OuterHTML("html", &html),
		)
		if err != nil {
			return "", fmt.Errorf("chromedp run ended with and error: %w", err)
		}

		return html, nil
}
