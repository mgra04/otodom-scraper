package models

type NextPageData struct {
	PageProps struct {
		Tracking struct {
			Listing struct {
				AdImpressions []int `json:"ad_impressions"`
			} `json:"listing"`
		} `json:"tracking"`
	} `json:"pageProps"`
}

type NextPageDataDetail struct {
	Props struct {
		PageProps struct {
			Ad struct {
				ID     int `json:"id"`
				Target struct {
					Price float64 `json:"Price"`
				} `json:"target"`
				Location struct {
					Coordinates struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
					} `json:"coordinates"`
					Address struct {
						District struct {
							Name string `json:"name"`
						} `json:"district"`
					} `json:"address"`
				} `json:"location"`
				TopInformation        []InfoItem `json:"topInformation"`
				AdditionalInformation []InfoItem `json:"additionalInformation"`
			} `json:"ad"`
		} `json:"pageProps"`
	} `json:"props"`
}

type InfoItem struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}