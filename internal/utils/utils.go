package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func CleanValue(val string) string {
    if strings.Contains(val, "::") {
        parts := strings.Split(val, "::")
        return parts[1]
    }
    return val
}

func GetFirst(slice []string) string {
    if len(slice) > 0 {
        return slice[0]
    }
    return ""
}

func ExtractInt(val string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(val)
	if match == "" {
		return 0
	}
	i, _ := strconv.Atoi(match)
	return i
}

func ParseFloat(val string) float64 {
	val = strings.Replace(val, ",", ".", -1)
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func SleepTimeFromRangeMs(min int, max int) time.Duration {
	randomMs := rand.Intn(max - min + 1)
	sleepDuration := time.Duration(min + randomMs) * time.Millisecond
	return sleepDuration
}

func GetRandomReferer() string {
    page := rand.Intn(10) + 1
    return fmt.Sprintf("https://www.otodom.pl/pl/wyniki/sprzedaz/mieszkanie/malopolskie/krakow?page=%d", page)
}