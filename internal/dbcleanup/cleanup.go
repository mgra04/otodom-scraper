package dbcleanup

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

var districtNameByFile = map[string]string{
	"1-stare-miasto":                "Stare Miasto",
	"2-grzegórzki":                  "Grzegórzki",
	"3-prądnik-czerwony":            "Prądnik Czerwony",
	"4-prądnik-biały":               "Prądnik Biały",
	"5-krowodrza":                   "Krowodrza",
	"6-bronowice":                   "Bronowice",
	"7-zwierzyniec":                 "Zwierzyniec",
	"8-dębniki":                     "Dębniki",
	"9-łagiewniki-borek-fałęcki":     "Łagiewniki-Borek Fałęcki",
	"10-swoszowice":                 "Swoszowice",
	"11-podgórze-duchackie":         "Podgórze Duchackie",
	"12-biezanow-prokocim":          "Bieżanów-Prokocim",
	"13-podgórze":                   "Podgórze",
	"14-czyżyny":                    "Czyżyny",
	"15-mistrzejowice":              "Mistrzejowice",
	"16-bieńczyce":                  "Bieńczyce",
	"17-wzgórza-krzesławickie":      "Wzgórza Krzesławickie",
	"18-nowa-huta":                  "Nowa Huta",
}

var (
	districtPolygons     map[string]orb.MultiPolygon
	districtPolygonsOnce sync.Once
	districtPolygonsErr  error
)

func PriceCleanup(db *sql.DB) (string, error) {
	query := `
	DELETE FROM Apartmentsv2
	WHERE total_price = 0 OR total_price IS NULL;
	`

	result, err := db.Exec(query)
	if err != nil {
		return "Error during price cleanup", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return "Error during price cleanup", err
	}

	return fmt.Sprintf("Price cleanup done (removed %d rows)", rows), nil
}

func DistrictCleanup(db *sql.DB) (string, error) {
	const querySelect = `
SELECT id, latitude, longitude
FROM Apartmentsv2
WHERE district IS NULL
   OR LTRIM(RTRIM(district)) = ''
   OR district NOT IN (
        N'Stare Miasto',
        N'Grzegórzki',
        N'Prądnik Czerwony',
        N'Prądnik Biały',
        N'Krowodrza',
        N'Bronowice',
        N'Zwierzyniec',
        N'Dębniki',
        N'Łagiewniki-Borek Fałęcki',
        N'Swoszowice',
        N'Podgórze Duchackie',
        N'Bieżanów-Prokocim',
        N'Podgórze',
        N'Czyżyny',
        N'Mistrzejowice',
        N'Bieńczyce',
        N'Wzgórza Krzesławickie',
        N'Nowa Huta'
    );
`

	const queryUpdate = `UPDATE Apartmentsv2 SET district = @district WHERE id = @id;`

	rows, err := db.Query(querySelect)
	if err != nil {
		return "Error during district cleanup", err
	}
	defer rows.Close()

	var (
		updated int
		skipped int
		failed  int
	)

	for rows.Next() {
		var (
			id   int
			lat  sql.NullFloat64
			long sql.NullFloat64
		)

		if err := rows.Scan(&id, &lat, &long); err != nil {
			return "Error during district cleanup", err
		}

		if !lat.Valid || !long.Valid {
			skipped++
			continue
		}

		district, err := PointToDistrict(db, lat.Float64, long.Float64)
		if err != nil {
			failed++
			log.Printf("District lookup failed for id %d (lat=%.6f, long=%.6f): %v", id, lat.Float64, long.Float64, err)
			continue
		}

		if _, err := db.Exec(queryUpdate, sql.Named("district", district), sql.Named("id", id)); err != nil {
			failed++
			log.Printf("Failed to update district for id %d: %v", id, err)
			continue
		}

		updated++
	}

	if err := rows.Err(); err != nil {
		return "Error during district cleanup", err
	}

	return fmt.Sprintf("District cleanup done (updated: %d, skipped: %d, failed: %d)", updated, skipped, failed), nil
}

func PointToDistrict(db *sql.DB, lat float64, long float64) (string, error) {
	_ = db

	if err := loadDistrictPolygons(); err != nil {
		return "", err
	}

	point := orb.Point{long, lat}

	for name, polygon := range districtPolygons {
		if planar.MultiPolygonContains(polygon, point) {
			return name, nil
		}
	}

	return "", fmt.Errorf("point outside known Kraków districts (lat=%.6f, long=%.6f)", lat, long)
}

func loadDistrictPolygons() error {
	districtPolygonsOnce.Do(func() {
		dir := resolveDistrictsDir()

		entries, err := os.ReadDir(dir)
		if err != nil {
			districtPolygonsErr = fmt.Errorf("read districts dir: %w", err)
			return
		}

		districtPolygons = make(map[string]orb.MultiPolygon)

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".geojson" {
				continue
			}

			base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			name, ok := districtNameByFile[base]
			if !ok {
				name = base
			}

			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				districtPolygonsErr = fmt.Errorf("read %s: %w", entry.Name(), err)
				return
			}

			geom, err := geojson.UnmarshalGeometry(data)
			if err != nil {
				districtPolygonsErr = fmt.Errorf("parse %s: %w", entry.Name(), err)
				return
			}

			switch g := geom.Geometry().(type) {
			case orb.MultiPolygon:
				districtPolygons[name] = g
			case orb.Polygon:
				districtPolygons[name] = orb.MultiPolygon{g}
			default:
				districtPolygonsErr = fmt.Errorf("%s: unsupported geometry type %T", entry.Name(), geom.Geometry())
				return
			}
		}

		if len(districtPolygons) == 0 {
			districtPolygonsErr = errors.New("no district polygons loaded")
		}
	})

	return districtPolygonsErr
}

func resolveDistrictsDir() string {
	if _, err := os.Stat("/cracow-districts"); err == nil {
		return "/cracow-districts"
	}

	return "cracow-districts"
}

func DeleteInvalidDistrictRows(db *sql.DB) (string, error) {
	const queryDelete = `
DELETE FROM Apartmentsv2
WHERE district IS NULL
   OR LTRIM(RTRIM(district)) = ''
   OR district NOT IN (
		N'Stare Miasto',
		N'Grzegórzki',
		N'Prądnik Czerwony',
		N'Prądnik Biały',
		N'Krowodrza',
		N'Bronowice',
		N'Zwierzyniec',
		N'Dębniki',
		N'Łagiewniki-Borek Fałęcki',
		N'Swoszowice',
		N'Podgórze Duchackie',
		N'Bieżanów-Prokocim',
		N'Podgórze',
		N'Czyżyny',
		N'Mistrzejowice',
		N'Bieńczyce',
		N'Wzgórza Krzesławickie',
		N'Nowa Huta'
	);
`

	result, err := db.Exec(queryDelete)
	if err != nil {
		return "Error deleting invalid districts", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return "Error deleting invalid districts", err
	}

	return fmt.Sprintf("Invalid districts cleanup done (removed %d rows)", rows), nil
}

func CountInvalidDistrictRows(db *sql.DB) (int, error) {
	const query = `
SELECT COUNT(*)
FROM Apartmentsv2
WHERE district IS NULL
   OR LTRIM(RTRIM(district)) = ''
   OR district NOT IN (
        N'Stare Miasto',
        N'Grzegórzki',
        N'Prądnik Czerwony',
        N'Prądnik Biały',
        N'Krowodrza',
        N'Bronowice',
        N'Zwierzyniec',
        N'Dębniki',
        N'Łagiewniki-Borek Fałęcki',
        N'Swoszowice',
        N'Podgórze Duchackie',
        N'Bieżanów-Prokocim',
        N'Podgórze',
        N'Czyżyny',
        N'Mistrzejowice',
        N'Bieńczyce',
        N'Wzgórza Krzesławickie',
        N'Nowa Huta'
    );`

	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func RunCleanup(db *sql.DB) {
	if msg, err := PriceCleanup(db); err != nil {
		log.Printf("Błąd czyszczenia cen: %v", err)
	} else {
		log.Println(msg)
	}

	if msg, err := DistrictCleanup(db); err != nil {
		log.Printf("Błąd czyszczenia dzielnic: %v", err)
	} else {
		log.Println(msg)
	}

	if msg, err := DeleteInvalidDistrictRows(db); err != nil {
		log.Printf("Błąd usuwania rekordów z niepoprawną dzielnicą: %v", err)
	} else {
		log.Println(msg)
	}
}