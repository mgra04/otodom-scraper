package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"otodom-scraper/internal/models"
	"otodom-scraper/internal/utils"

	"github.com/joho/godotenv"

	_ "github.com/microsoft/go-mssqldb"
)

func InitDB() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	server := os.Getenv("DB_SERVER")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s",
		server, user, password, port)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	return db, err
}

func SetupSchema(db *sql.DB) error {
	schema, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

func SaveIDsToDB(db *sql.DB, ids []int) {
	for _, id := range ids {
		query := `
            IF NOT EXISTS (SELECT 1 FROM Apartment_IDs WHERE id = @id)
            BEGIN
                INSERT INTO Apartment_IDs (id) VALUES (@id)
            END`
		_, err := db.Exec(query, sql.Named("id", id))

		if err != nil {
			log.Printf("Błąd podczas zapisu ID %d: %v", id, err)
			continue
		}
	}
}

func SaveApartmentDetails(db *sql.DB, data *models.NextPageDataDetail) error {
	ad := data.Props.PageProps.Ad

	info := make(map[string][]string)
	for _, item := range append(ad.TopInformation, ad.AdditionalInformation...) {
		info[item.Label] = item.Values
	}

	rooms := utils.ExtractInt(utils.GetFirst(info["rooms_num"]))
	buildYear := utils.ExtractInt(utils.GetFirst(info["build_year"]))

	floorStr := utils.GetFirst(info["floor"])
	floor := utils.ExtractInt(utils.CleanValue(floorStr))

	totalFloors := 0
	if len(info["floor"]) > 1 {
		totalFloors = utils.ExtractInt(info["floor"][1])
	}

	area := utils.ParseFloat(utils.GetFirst(info["area"]))

	hasElevator := 0
	if utils.GetFirst(info["lift"]) == "::y" {
		hasElevator = 1
	}

	var availableFrom interface{}
	dateStr := utils.GetFirst(info["free_from"])

	if dateStr != "" {
		availableFrom = dateStr
	} else {
		availableFrom = nil
	}

	query := `
        INSERT INTO Apartments (
            id, source, district, latitude, longitude, total_price, 
            square_footage, rooms, floor, total_floors, finishing_state, 
            market_type, advertiser_type, build_year, has_elevator, 
            heating_type, avaiable_from
        ) VALUES (
            @id, @source, @district, @lat, @long, @price, 
            @area, @rooms, @floor, @total_floors, @finish, 
            @market, @adv, @year, @lift, @heat, @free
        );
        UPDATE Apartment_IDs SET is_scraped = 1 WHERE id = @id;`

	_, err := db.Exec(query,
		sql.Named("id", ad.ID),
		sql.Named("source", fmt.Sprintf("https://www.otodom.pl/%d", ad.ID)),
		sql.Named("district", ad.Location.Address.District.Name),
		sql.Named("lat", ad.Location.Coordinates.Latitude),
		sql.Named("long", ad.Location.Coordinates.Longitude),
		sql.Named("price", ad.Target.Price),
		sql.Named("area", area),
		sql.Named("rooms", rooms),
		sql.Named("floor", floor),
		sql.Named("total_floors", totalFloors),
		sql.Named("finish", utils.CleanValue(utils.GetFirst(info["construction_status"]))),
		sql.Named("market", utils.CleanValue(utils.GetFirst(info["market"]))),
		sql.Named("adv", utils.CleanValue(utils.GetFirst(info["advertiser_type"]))),
		sql.Named("year", buildYear),
		sql.Named("lift", hasElevator),
		sql.Named("heat", utils.CleanValue(utils.GetFirst(info["heating"]))),
		sql.Named("free", availableFrom),
	)
	return err
}

func MarkIDScraped(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE Apartment_IDs SET is_scraped = 1 WHERE id = @id", sql.Named("id", id))
	return err
}
