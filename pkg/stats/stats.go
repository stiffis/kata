package stats

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Session struct {
	ID         int
	Text       string
	WPM        float64
	Accuracy   float64
	Duration   float64
	ErrorCount int
	Timestamp  time.Time
}

type KeyStat struct {
	Key           string
	Errors        int
	Successes     int
	LastPracticed time.Time
	Interval      int
	Repetitions   int
	EaseFactor    float64
}

func (k *KeyStat) UpdateSM2(quality int) {
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}

	if quality >= 3 {
		if k.Repetitions == 0 {
			k.Interval = 1
		} else if k.Repetitions == 1 {
			k.Interval = 6
		} else {
			k.Interval = int(float64(k.Interval) * k.EaseFactor)
		}
		k.Repetitions++
	} else {
		k.Repetitions = 0
		k.Interval = 1
	}

	k.EaseFactor = k.EaseFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if k.EaseFactor < 1.3 {
		k.EaseFactor = 1.3
	}

	k.LastPracticed = time.Now()
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
		interval INTEGER DEFAULT 0,
		repetitions INTEGER DEFAULT 0,
		ease_factor REAL DEFAULT 2.5,
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

func AnalyzeErrors(target, input string) ErrorAnalysis {
	analysis := ErrorAnalysis{
		CharErrors:   make(map[string]int),
		BigramErrors: make(map[string]int),
	}

	minLen := len(input)
	if minLen > len(target) {
		minLen = len(target)
	}

	for i := 0; i < minLen; i++ {
		targetChar := string(target[i])
		inputChar := string(input[i])

		if targetChar != inputChar {
			analysis.CharErrors[targetChar]++

			if i > 0 {
				bigram := string(target[i-1]) + targetChar
				analysis.BigramErrors[bigram]++
			}
		}
	}

	return analysis
}

func (db *DB) UpdateKeyStats(target, input string) error {
	minLen := len([]rune(input))
	targetRunes := []rune(target)
	inputRunes := []rune(input)

	if minLen > len(targetRunes) {
		minLen = len(targetRunes)
	}

	charStats := make(map[string]struct {
		errors    int
		successes int
	})

	for i := 0; i < minLen; i++ {
		key := string(targetRunes[i])
		stats := charStats[key]

		if targetRunes[i] == inputRunes[i] {
			stats.successes++
		} else {
			stats.errors++
		}

		charStats[key] = stats
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO key_stats (key, errors, successes, last_practiced)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		errors = errors + ?,
		successes = successes + ?,
		last_practiced = ?
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for key, stats := range charStats {
		_, err := stmt.Exec(key, stats.errors, stats.successes, now, stats.errors, stats.successes, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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

func (db *DB) GetWeakestKeys(limit int) ([]KeyStat, error) {
	query := `
	SELECT key, errors, successes, last_practiced, interval, repetitions, ease_factor
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
		if err := rows.Scan(&s.Key, &s.Errors, &s.Successes, &s.LastPracticed, &s.Interval, &s.Repetitions, &s.EaseFactor); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (db *DB) GetDueKeys(limit int) ([]KeyStat, error) {
	query := `
	SELECT key, errors, successes, last_practiced, interval, repetitions, ease_factor
	FROM key_stats
	WHERE (errors + successes) >= 3
	ORDER BY last_practiced ASC
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var stats []KeyStat
	for rows.Next() {
		var s KeyStat
		if err := rows.Scan(&s.Key, &s.Errors, &s.Successes, &s.LastPracticed, &s.Interval, &s.Repetitions, &s.EaseFactor); err != nil {
			return nil, err
		}

		daysSince := now.Sub(s.LastPracticed).Hours() / 24
		if daysSince >= float64(s.Interval) {
			stats = append(stats, s)
			if len(stats) >= limit {
				break
			}
		}
	}

	return stats, nil
}

func (db *DB) GetAllKeyStats() ([]KeyStat, error) {
	query := `
	SELECT key, errors, successes, last_practiced, interval, repetitions, ease_factor
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
		if err := rows.Scan(&s.Key, &s.Errors, &s.Successes, &s.LastPracticed, &s.Interval, &s.Repetitions, &s.EaseFactor); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

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
