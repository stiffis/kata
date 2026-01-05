package stats

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Session struct {
	ID        int
	Text      string
	WPM       float64
	Accuracy  float64
	Duration  float64
	ErrorCount int
	Timestamp time.Time
}

type KeyStat struct {
	Key           string
	Errors        int
	Successes     int
	LastPracticed time.Time
}

type ErrorAnalysis struct {
	CharErrors   map[string]int // individual characters that were typed wrong
	BigramErrors map[string]int // two-character sequences that were typed wrong
}

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		wpm REAL NOT NULL,
		accuracy REAL NOT NULL,
		duration REAL NOT NULL,
		error_count INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS key_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		errors INTEGER DEFAULT 0,
		successes INTEGER DEFAULT 0,
		last_practiced DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(key)
	);
	`
	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) SaveSession(session Session) error {
	query := `
	INSERT INTO sessions (text, wpm, accuracy, duration, error_count, timestamp)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, session.Text, session.WPM, session.Accuracy,
		session.Duration, session.ErrorCount, session.Timestamp)
	return err
}

func (db *DB) GetRecentSessions(limit int) ([]Session, error) {
	query := `
	SELECT id, text, wpm, accuracy, duration, error_count, timestamp
	FROM sessions
	ORDER BY timestamp DESC
	LIMIT ?
	`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Text, &s.WPM, &s.Accuracy, &s.Duration, &s.ErrorCount, &s.Timestamp); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (db *DB) GetAverageWPM() (float64, error) {
	var avg float64
	query := `SELECT AVG(wpm) FROM sessions`
	err := db.conn.QueryRow(query).Scan(&avg)
	return avg, err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// AnalyzeErrors compares user input with target text and returns analysis of errors
func AnalyzeErrors(target, input string) ErrorAnalysis {
	analysis := ErrorAnalysis{
		CharErrors:   make(map[string]int),
		BigramErrors: make(map[string]int),
	}

	// Compare character by character
	minLen := len(input)
	if minLen > len(target) {
		minLen = len(target)
	}

	for i := 0; i < minLen; i++ {
		targetChar := string(target[i])
		inputChar := string(input[i])

		if targetChar != inputChar {
			// Record character error
			analysis.CharErrors[targetChar]++

			// Record bigram error if possible
			if i > 0 {
				bigram := string(target[i-1]) + targetChar
				analysis.BigramErrors[bigram]++
			}
		}
	}

	return analysis
}

// UpdateKeyStats updates the database with errors and successes for each character
func (db *DB) UpdateKeyStats(target, input string) error {
	minLen := len(input)
	if minLen > len(target) {
		minLen = len(target)
	}

	// Track all characters typed
	charStats := make(map[string]struct {
		errors    int
		successes int
	})

	for i := 0; i < minLen; i++ {
		key := string(target[i])
		stats := charStats[key]

		if target[i] == input[i] {
			stats.successes++
		} else {
			stats.errors++
		}

		charStats[key] = stats
	}

	// Update database
	for key, stats := range charStats {
		if err := db.upsertKeyStats(key, stats.errors, stats.successes); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) upsertKeyStats(key string, errors, successes int) error {
	query := `
	INSERT INTO key_stats (key, errors, successes, last_practiced)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		errors = errors + ?,
		successes = successes + ?,
		last_practiced = ?
	`
	now := time.Now()
	_, err := db.conn.Exec(query, key, errors, successes, now, errors, successes, now)
	return err
}

// GetWeakestKeys returns the keys with the highest error rate
func (db *DB) GetWeakestKeys(limit int) ([]KeyStat, error) {
	query := `
	SELECT key, errors, successes, last_practiced
	FROM key_stats
	WHERE (errors + successes) >= 5
	ORDER BY CAST(errors AS REAL) / (errors + successes) DESC
	LIMIT ?
	`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []KeyStat
	for rows.Next() {
		var s KeyStat
		if err := rows.Scan(&s.Key, &s.Errors, &s.Successes, &s.LastPracticed); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetAllKeyStats returns all key statistics
func (db *DB) GetAllKeyStats() ([]KeyStat, error) {
	query := `
	SELECT key, errors, successes, last_practiced
	FROM key_stats
	ORDER BY (errors + successes) DESC
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []KeyStat
	for rows.Next() {
		var s KeyStat
		if err := rows.Scan(&s.Key, &s.Errors, &s.Successes, &s.LastPracticed); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetSessionsForGraph returns sessions for graphing (limited and ordered)
func (db *DB) GetSessionsForGraph(limit int) ([]Session, error) {
	query := `
	SELECT id, text, wpm, accuracy, duration, error_count, timestamp
	FROM sessions
	ORDER BY timestamp ASC
	LIMIT ?
	`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Text, &s.WPM, &s.Accuracy, &s.Duration, &s.ErrorCount, &s.Timestamp); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}
