package main

import (
	"fmt"
	"log"

	"otodom-scraper/internal/dbcleanup"
	"otodom-scraper/internal/exporter"
	"otodom-scraper/internal/scraper"
	"otodom-scraper/internal/storage"

	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := storage.SetupSchema(db); err != nil {
		log.Fatalf("Error setting up schema: %v", err)
	}

	totalIDs := 0
	scrapedIDs := 0

	if err := db.QueryRow("SELECT COUNT(*) FROM Apartment_IDs").Scan(&totalIDs); err != nil {
		log.Fatalf("Error checking ID table: %v", err)
	}

	if totalIDs == 0 {
		fmt.Println("ID database is empty. Starting to collect IDs from 339 pages...")
		scraper.RunIDScraper(db)

		if err := db.QueryRow("SELECT COUNT(*) FROM Apartment_IDs").Scan(&totalIDs); err != nil {
			log.Fatalf("Error rechecking ID table: %v", err)
		}
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM Apartment_IDs WHERE is_scraped = 1").Scan(&scrapedIDs); err != nil {
		log.Fatalf("Error checking scraped IDs: %v", err)
	}

	if totalIDs == 0 {
		fmt.Println("No IDs after phase 1. Exiting.")
		return
	}

	if scrapedIDs < totalIDs {
		fmt.Printf("Found %d/%d scraped IDs. Starting Phase 2: Fetching apartment details...\n", scrapedIDs, totalIDs)
		scraper.RunDetailScraper(db)
	}

	invalidCount, err := dbcleanup.CountInvalidDistrictRows(db)
	if err != nil {
		log.Fatalf("Error checking invalid district rows: %v", err)
	}

	if invalidCount > 0 {
		fmt.Printf("Found %d records to clean. Running cleanup...\n", invalidCount)
		dbcleanup.RunCleanup(db)
	} else {
		fmt.Println("No records to clean.")
	}

	csvPath, xlsxPath, err := exporter.ExportApartmentsV2(db)
	if err != nil {
		log.Fatalf("Export error: %v", err)
	}
	log.Printf("CSV: %s, XLSX: %s", csvPath, xlsxPath)
}