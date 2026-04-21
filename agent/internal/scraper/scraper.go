package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ScrapeError is a classified error from page fetching.
type ScrapeError struct {
	Type    string // matches error_type enum values on the orchestrator
	Message string
}

func (e *ScrapeError) Error() string { return fmt.Sprintf("%s: %s", e.Type, e.Message) }

// Result holds the rendered page data captured by the scraper.
type Result struct {
	HTML       string
	Screenshot []byte // raw PNG bytes; may be empty if capture failed
}

// Scraper fetches fully-rendered pages via headless Chrome.
type Scraper struct {
	logger *slog.Logger
	run    func(context.Context, ...chromedp.Action) error
}

// New returns a Scraper ready to fetch pages.
func New(logger *slog.Logger) *Scraper {
	return &Scraper{logger: logger, run: chromedp.Run}
}

// Fetch navigates to url, optionally waits for waitSelector to appear in the
// DOM, then returns the fully rendered HTML and a viewport screenshot.
// On page_timeout the partial result (whatever loaded) is returned alongside the error.
func (s *Scraper) Fetch(ctx context.Context, rawURL, waitSelector string, timeout time.Duration) (*Result, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		// Stealth: suppress automation signals that trigger bot-detection.
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-infobars", true),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)
	defer cancelAlloc()

	tabCtx, cancelTab := chromedp.NewContext(allocCtx)
	defer cancelTab()

	timeoutCtx, cancelTimeout := context.WithTimeout(tabCtx, timeout)
	defer cancelTimeout()

	var html string
	var screenshot []byte

	tasks := chromedp.Tasks{
		// Hide navigator.webdriver before any page script runs.
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(
				`Object.defineProperty(navigator,'webdriver',{get:()=>undefined})`,
			).Do(ctx)
			return err
		}),
		chromedp.Navigate(rawURL),
		// Wait for the body to exist, then allow JS/CSS to paint.
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(1500 * time.Millisecond),
		// Scroll down then back to top to trigger lazy-loaded content.
		chromedp.Evaluate(`window.scrollTo(0, (document.body||document.documentElement).scrollHeight)`, nil),
		chromedp.Sleep(500 * time.Millisecond),
		chromedp.Evaluate(`window.scrollTo(0, 0)`, nil),
		chromedp.Sleep(300 * time.Millisecond),
	}

	if waitSelector != "" {
		tasks = append(tasks, chromedp.WaitVisible(waitSelector, chromedp.ByQuery))
	}

	tasks = append(tasks,
		chromedp.CaptureScreenshot(&screenshot),
		chromedp.OuterHTML("html", &html),
	)

	if err := s.run(timeoutCtx, tasks...); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Best-effort: capture whatever loaded using the non-timeout tab context.
			_ = s.run(tabCtx,
				chromedp.CaptureScreenshot(&screenshot),
				chromedp.OuterHTML("html", &html),
			)
			return &Result{HTML: html, Screenshot: screenshot}, &ScrapeError{
				Type:    "page_timeout",
				Message: fmt.Sprintf("page did not load within %s", timeout),
			}
		}

		return nil, &ScrapeError{
			Type:    "navigation_error",
			Message: err.Error(),
		}
	}

	s.logger.Debug("page fetched", "url", rawURL, "html_bytes", len(html), "screenshot_bytes", len(screenshot))

	return &Result{HTML: html, Screenshot: screenshot}, nil
}
