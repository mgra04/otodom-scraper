package scraper

import (
	"encoding/json"
	"fmt"
	"otodom-scraper/internal/models"

	"github.com/PuerkitoBio/goquery"
	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func fetchOfferIDs(client tls_client.HttpClient, url string) ([]int, error) {
	req, _ := fhttp.NewRequest("GET", url, nil)

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.otodom.pl/pl/wyniki/sprzedaz/mieszkanie/malopolskie/krakow")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status error: %d", resp.StatusCode)
	}

	var data models.NextPageData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.PageProps.Tracking.Listing.AdImpressions, nil
}

func FetchDetailsFromHTML(client tls_client.HttpClient, id int, profile BrowserProfile, referer string) (*models.NextPageDataDetail, error) {
	url := fmt.Sprintf("https://www.otodom.pl/%d", id)
	req, _ := fhttp.NewRequest("GET", url, nil)

	req.Header.Set("user-agent", profile.UA)
	for k, v := range profile.ClientHints {
		req.Header.Set(k, v)
	}

	req.Header.Set("Referer", referer)

	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("accept-language", profile.ClientHints["accept-language"])
	req.Header.Set("accept-encoding", profile.ClientHints["accept-encoding"])
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-user", "?1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("403 Forbidden")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	scriptContent := doc.Find("script#__NEXT_DATA__").Text()
	if scriptContent == "" {
		return nil, fmt.Errorf("tag __NEXT_DATA__ not found")
	}

	var data models.NextPageDataDetail
	if err := json.Unmarshal([]byte(scriptContent), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func SessionWarmUp(client tls_client.HttpClient, profile BrowserProfile) error {
	urlAddr := "https://www.otodom.pl/pl/wyniki/sprzedaz/mieszkanie/malopolskie/krakow"
	req, _ := fhttp.NewRequest("GET", urlAddr, nil)

	req.Header.Set("user-agent", profile.UA)
	for k, v := range profile.ClientHints {
		req.Header.Set(k, v)
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Warm up ended successfully.")
	return nil
}
