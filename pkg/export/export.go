package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"kata/pkg/stats"
)

type ExportData struct {
	ExportDate    time.Time         `json:"export_date"`
	AverageWPM    float64           `json:"average_wpm"`
	Sessions      []stats.Session   `json:"sessions"`
	KeyStatistics []stats.KeyStat   `json:"key_statistics"`
}

func ToJSON(db *stats.DB, outputFile string) error {
	data, err := gatherData(db)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func ToCSV(db *stats.DB, outputFile string) error {
	sessions, err := db.GetRecentSessions(1000) // Get all sessions
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Timestamp", "WPM", "Accuracy", "Duration", "ErrorCount"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write sessions
	for _, s := range sessions {
		record := []string{
			fmt.Sprintf("%d", s.ID),
			s.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", s.WPM),
			fmt.Sprintf("%.2f", s.Accuracy),
			fmt.Sprintf("%.2f", s.Duration),
			fmt.Sprintf("%d", s.ErrorCount),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

func gatherData(db *stats.DB) (ExportData, error) {
	data := ExportData{
		ExportDate: time.Now(),
	}

	// Get average WPM
	avgWPM, err := db.GetAverageWPM()
	if err == nil {
		data.AverageWPM = avgWPM
	}

	// Get all sessions
	sessions, err := db.GetRecentSessions(1000)
	if err != nil {
		return data, fmt.Errorf("failed to get sessions: %w", err)
	}
	data.Sessions = sessions

	// Get all key statistics
	keyStats, err := db.GetAllKeyStats()
	if err != nil {
		return data, fmt.Errorf("failed to get key stats: %w", err)
	}
	data.KeyStatistics = keyStats

	return data, nil
}
