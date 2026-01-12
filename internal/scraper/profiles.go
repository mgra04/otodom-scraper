package scraper

import (
	"math/rand"

	"github.com/bogdanfinn/tls-client/profiles"
)

type BrowserProfile struct {
    UA              string
    ClientHints     map[string]string
    TLSProfile      profiles.ClientProfile
}

func GetRandomProfile() BrowserProfile {
    profilesList := []BrowserProfile{
		{
			UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			TLSProfile: profiles.Chrome_131,
			ClientHints: map[string]string{
				"upgrade-insecure-requests": "1",
				"sec-ch-ua":                 `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"Windows"`,
				"accept-language":           "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7",
				"accept-encoding": "gzip, deflate, br, zstd",
			},
		},
		{
			UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			TLSProfile: profiles.Chrome_124,
			ClientHints: map[string]string{
				"upgrade-insecure-requests": "1",
				"sec-ch-ua":                 `"Google Chrome";v="124", "Chromium";v="124", "Not_A Brand";v="99"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"Windows"`,
				"sec-ch-ua-full-version":    `"124.0.6367.60"`,
				"accept-language":           "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7",
				"accept-encoding": "gzip, deflate, br, zstd",
			},
		},
		{
			UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			TLSProfile: profiles.Chrome_120,
			ClientHints: map[string]string{
				"upgrade-insecure-requests": "1",
				"sec-ch-ua":                 `"Google Chrome";v="120", "Chromium";v="120", "Not_A Brand";v="8"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"Windows"`,
				"accept-language":           "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7",
				"accept-encoding": "gzip, deflate, br, zstd",
			},
		},
		{
			UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
			TLSProfile: profiles.Chrome_117,
			ClientHints: map[string]string{
				"upgrade-insecure-requests": "1",
				"sec-ch-ua":                 `"Google Chrome";v="117", "Chromium";v="117", "Not_A Brand";v="99"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"Windows"`,
				"accept-language":           "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7",
				"accept-encoding": "gzip, deflate, br, zstd",
			},
		},
		{
			UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
			TLSProfile: profiles.Chrome_112,
			ClientHints: map[string]string{
				"upgrade-insecure-requests": "1",
				"sec-ch-ua":                 `"Google Chrome";v="112", "Chromium";v="112", "Not_A Brand";v="24"`,
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        `"Windows"`,
				"accept-language":           "pl-PL,pl;q=0.9,en-US;q=0.8,en;q=0.7",
				"accept-encoding": "gzip, deflate, br, zstd",
			},
		},
    }
    return profilesList[rand.Intn(len(profilesList))]
}