package scraper

import (
	"context"
	"log"
	"os"
	"strings"

	"otodom-scraper/internal/utils"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func GetFreshCookies(userAgent string) []*fhttp.Cookie {
	headless := true
	if strings.ToLower(os.Getenv("SCRAPER_HEADLESS")) == "false" {
		headless = false
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.UserAgent(userAgent),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1920,1080"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var cookies []*fhttp.Cookie

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate("https://www.otodom.pl/pl/wyniki/sprzedaz/mieszkanie/malopolskie/krakow"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(utils.SleepTimeFromRangeMs(7000, 11000)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromeCookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			for _, c := range chromeCookies {
				cookies = append(cookies, &fhttp.Cookie{
					Name:   c.Name,
					Value:  c.Value,
					Domain: c.Domain,
					Path:   c.Path,
				})
			}
			return nil
		}),
	)

	if err != nil {
		log.Printf("chromedp error: %v", err)
	}

	return cookies
}
