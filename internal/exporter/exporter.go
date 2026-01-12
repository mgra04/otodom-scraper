package exporter

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

func ExportApartmentsV2(db *sql.DB) (string, string, error) {
	const table = "Apartmentsv2"

	if err := os.MkdirAll("data", 0o755); err != nil {
		return "", "", fmt.Errorf("create data dir: %w", err)
	}

	csvPath := filepath.Join("data", "apartments.csv")
	xlsxPath := filepath.Join("data", "apartments.xlsx")

	rows, err := db.Query("SELECT * FROM " + table + " ORDER BY id")
	if err != nil {
		return "", "", fmt.Errorf("query table: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", "", fmt.Errorf("columns: %w", err)
	}

	csvFile, err := os.Create(csvPath)
	if err != nil {
		return "", "", fmt.Errorf("create csv: %w", err)
	}
	defer csvFile.Close()
	csvW := csv.NewWriter(csvFile)

	xlsx := excelize.NewFile()
	const sheet = "Sheet1"
	stream, err := xlsx.NewStreamWriter(sheet)
	if err != nil {
		return "", "", fmt.Errorf("excel stream: %w", err)
	}

	// Header row
	if err := csvW.Write(cols); err != nil {
		return "", "", fmt.Errorf("csv header: %w", err)
	}
	header := make([]interface{}, len(cols))
	for i, c := range cols {
		header[i] = c
	}
	if err := stream.SetRow("A1", header); err != nil {
		return "", "", fmt.Errorf("excel header: %w", err)
	}

	scanDest := make([]interface{}, len(cols))
	for i := range scanDest {
		var v any
		scanDest[i] = &v
	}

	excelRow := 2
	for rows.Next() {
		if err := rows.Scan(scanDest...); err != nil {
			return "", "", fmt.Errorf("scan: %w", err)
		}

		record := make([]string, len(cols))
		rowVals := make([]interface{}, len(cols))
		for i := range cols {
			val := *(scanDest[i].(*any))
			str := formatValue(val)
			record[i] = str
			rowVals[i] = str
		}

		if err := csvW.Write(record); err != nil {
			return "", "", fmt.Errorf("csv write: %w", err)
		}

		cellRef := fmt.Sprintf("A%d", excelRow)
		if err := stream.SetRow(cellRef, rowVals); err != nil {
			return "", "", fmt.Errorf("excel write row %d: %w", excelRow, err)
		}
		excelRow++
	}

	if err := rows.Err(); err != nil {
		return "", "", fmt.Errorf("rows error: %w", err)
	}

	csvW.Flush()
	if err := csvW.Error(); err != nil {
		return "", "", fmt.Errorf("csv flush: %w", err)
	}

	if err := stream.Flush(); err != nil {
		return "", "", fmt.Errorf("excel stream flush: %w", err)
	}

	if err := xlsx.SaveAs(xlsxPath); err != nil {
		return "", "", fmt.Errorf("save xlsx: %w", err)
	}

	return csvPath, xlsxPath, nil
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return fmt.Sprint(val)
	}
}