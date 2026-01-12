package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"otodom-scraper/internal/storage"
	"otodom-scraper/internal/utils"
	"strings"
	"sync"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	proxyList     []string
	proxyLastUsed = map[string]time.Time{}
	proxyInitOnce sync.Once
	proxyCooldown = 5 * time.Minute
)

func RunIDScraper(db *sql.DB) {
	buildID := os.Getenv("OTODOM_BUILD_ID")

	currentProfile := GetRandomProfile()
	client, err := GetSafeClient(currentProfile.TLSProfile, "")
	if err != nil {
		log.Fatalf("Client initialize failed: %v", err)
	}

	for page := 1; page <= 339; page++ {
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
		url := fmt.Sprintf("https://www.otodom.pl/_next/data/%s/pl/wyniki/sprzedaz/mieszkanie/malopolskie/krakow/krakow/krakow.json?page=%d", buildID, page)

		ids, err := fetchOfferIDs(client, url)
		if err != nil {
			log.Printf("Error on page %d: %v", page, err)
			continue
		}

		storage.SaveIDsToDB(db, ids)
	}
}

func RunDetailScraper(db *sql.DB) {
	const maxRequestsPerSession = 6
	var lastFailedID int
	var repeatFail int

	proxyInitOnce.Do(func() {
		list := strings.TrimSpace(os.Getenv("SCRAPER_PROXIES"))
		if list == "" {
			proxyList = nil
			return
		}
		parts := strings.Split(list, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				proxyList = append(proxyList, p)
			}
		}
	})

	newSession := func() (tls_client.HttpClient, BrowserProfile, error) {
		profile := GetRandomProfile()
		proxy := getRandomProxy()
		client, err := GetSafeClient(profile.TLSProfile, proxy)
		if err != nil {
			return nil, BrowserProfile{}, err
		}

		if proxy != "" {
			fmt.Printf("Using proxy: %s\n", proxy)
		}

		cookies := GetFreshCookies(profile.UA)
		u, _ := url.Parse("https://www.otodom.pl")
		client.SetCookies(u, cookies)

		if err := SessionWarmUp(client, profile); err != nil {
			return nil, BrowserProfile{}, err
		}

		time.Sleep(utils.SleepTimeFromRangeMs(1200, 2600))

		return client, profile, nil
	}

	for {
		client, profile, err := newSession()
		if err != nil {
			log.Printf("Failed to prepare session: %v", err)
			time.Sleep(45 * time.Second)
			continue
		}

		fmt.Println("New session.")

		for i := 0; i < maxRequestsPerSession; i++ {
			var id int
			err := db.QueryRow("SELECT TOP 1 id FROM Apartment_IDs WHERE is_scraped = 0").Scan(&id)
			if err == sql.ErrNoRows {
				fmt.Println("All IDs scraped. Exiting.")
				return
			}
			if err != nil {
				log.Printf("Error fetching ID from DB: %v", err)
				time.Sleep(15 * time.Second)
				break
			}

			if lastFailedID == id {
				repeatFail++
			} else {
				lastFailedID = id
				repeatFail = 1
			}

			if repeatFail >= 2 {
				log.Printf("ID %d repeated again – marking as scraped", id)
				_ = storage.MarkIDScraped(db, id)
				time.Sleep(utils.SleepTimeFromRangeMs(800, 1800))
				continue
			}

			referer := utils.GetRandomReferer()
			aptData, err := FetchDetailsFromHTML(client, id, profile, referer)
			if err != nil {
				if strings.Contains(err.Error(), "403") {
					fmt.Println("403 Forbidden – reseting session.")
					time.Sleep(utils.SleepTimeFromRangeMs(5000, 10000))
					break
				}
				if strings.Contains(err.Error(), "410") {
					log.Printf("ID %d deleted (410) – marking as scraped.", id)
					_ = storage.MarkIDScraped(db, id)
					time.Sleep(utils.SleepTimeFromRangeMs(800, 1800))
					continue
				}
				if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "http: server gave HTTP response to HTTPS client") {
					fmt.Println("Timeout – resetting session")
					time.Sleep(utils.SleepTimeFromRangeMs(5000, 10000))
					break
				}
				
				continue
			}
			repeatFail = 0

			storage.SaveApartmentDetails(db, aptData)
			fmt.Printf("Offer data with id: %d saved\n", id)

			time.Sleep(utils.SleepTimeFromRangeMs(2500, 5500))
		}

		fmt.Println("Session done, restarting...")
		time.Sleep(utils.SleepTimeFromRangeMs(5000, 9000))
	}
}

func getRandomProxy() string {
	if len(proxyList) == 0 {
		return ""
	}
	now := time.Now()
	eligible := make([]string, 0, len(proxyList))
	for _, p := range proxyList {
		if t, ok := proxyLastUsed[p]; !ok || now.Sub(t) > proxyCooldown {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) == 0 {
		p := proxyList[rand.Intn(len(proxyList))]
		proxyLastUsed[p] = now
		return p
	}
	p := eligible[rand.Intn(len(eligible))]
	proxyLastUsed[p] = now
	return p
}
