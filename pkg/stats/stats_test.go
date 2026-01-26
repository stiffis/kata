package stats

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestKeyStatStructure(t *testing.T) {
	tmpDB := "/tmp/kata_test_srs.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	stat := KeyStat{
		Key:           "a",
		Errors:        2,
		Successes:     8,
		LastPracticed: time.Now(),
		Interval:      1,
		Repetitions:   1,
		EaseFactor:    2.5,
	}

	if stat.Interval != 1 {
		t.Errorf("Expected Interval 1, got %d", stat.Interval)
	}
	if stat.EaseFactor != 2.5 {
		t.Errorf("Expected EaseFactor 2.5, got %f", stat.EaseFactor)
	}
}

func TestSM2Algorithm(t *testing.T) {
	cases := []struct {
		name                string
		initialReps         int
		initialInterval     int
		initialEase         float64
		quality             int
		expectedReps        int
		expectedMinInterval int
		expectedMaxInterval int
		expectedMinEase     float64
	}{
		{
			name:                "first review correct",
			initialReps:         0,
			initialInterval:     0,
			initialEase:         2.5,
			quality:             4,
			expectedReps:        1,
			expectedMinInterval: 1,
			expectedMaxInterval: 1,
			expectedMinEase:     2.4,
		},
		{
			name:                "second review correct",
			initialReps:         1,
			initialInterval:     1,
			initialEase:         2.5,
			quality:             4,
			expectedReps:        2,
			expectedMinInterval: 6,
			expectedMaxInterval: 6,
			expectedMinEase:     2.4,
		},
		{
			name:                "third review correct",
			initialReps:         2,
			initialInterval:     6,
			initialEase:         2.5,
			quality:             4,
			expectedReps:        3,
			expectedMinInterval: 15,
			expectedMaxInterval: 15,
			expectedMinEase:     2.4,
		},
		{
			name:                "failed review resets",
			initialReps:         3,
			initialInterval:     15,
			initialEase:         2.5,
			quality:             2,
			expectedReps:        0,
			expectedMinInterval: 1,
			expectedMaxInterval: 1,
			expectedMinEase:     2.1,
		},
		{
			name:                "ease factor minimum floor",
			initialReps:         1,
			initialInterval:     1,
			initialEase:         1.4,
			quality:             0,
			expectedReps:        0,
			expectedMinInterval: 1,
			expectedMaxInterval: 1,
			expectedMinEase:     1.3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stat := &KeyStat{
				Key:         "a",
				Repetitions: tc.initialReps,
				Interval:    tc.initialInterval,
				EaseFactor:  tc.initialEase,
			}

			stat.UpdateSM2(tc.quality)

			if stat.Repetitions != tc.expectedReps {
				t.Errorf("Expected %d repetitions, got %d", tc.expectedReps, stat.Repetitions)
			}

			if stat.Interval < tc.expectedMinInterval || stat.Interval > tc.expectedMaxInterval {
				t.Errorf("Expected interval between %d and %d, got %d",
					tc.expectedMinInterval, tc.expectedMaxInterval, stat.Interval)
			}

			if stat.EaseFactor < tc.expectedMinEase {
				t.Errorf("Expected ease factor >= %.2f, got %.2f", tc.expectedMinEase, stat.EaseFactor)
			}

			if stat.EaseFactor < 1.3 {
				t.Errorf("Ease factor below minimum 1.3: %.2f", stat.EaseFactor)
			}
		})
	}
}

func TestGetDueKeys(t *testing.T) {
	tmpDB := "/tmp/kata_test_due_keys.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	// Insert test data
	now := time.Now()
	testKeys := []struct {
		key           string
		errors        int
		successes     int
		lastPracticed time.Time
		interval      int
	}{
		{"a", 2, 8, now.Add(-2 * 24 * time.Hour), 1},   // due (1 day interval, practiced 2 days ago)
		{"b", 1, 9, now.Add(-7 * 24 * time.Hour), 6},   // due (6 day interval, practiced 7 days ago)
		{"c", 0, 10, now.Add(-1 * time.Hour), 1},       // not due (1 day interval, practiced 1 hour ago)
		{"d", 5, 5, now.Add(-10 * 24 * time.Hour), 15}, // not due yet (15 day interval, practiced 10 days ago)
		{"e", 1, 1, now.Add(-3 * 24 * time.Hour), 1},   // only 2 attempts total, should be excluded
	}

	for _, tk := range testKeys {
		_, err := db.conn.Exec(`
			INSERT INTO key_stats (key, errors, successes, last_practiced, interval, repetitions, ease_factor)
			VALUES (?, ?, ?, ?, ?, 1, 2.5)
		`, tk.key, tk.errors, tk.successes, tk.lastPracticed, tk.interval)
		if err != nil {
			t.Fatalf("Failed to insert test key %s: %v", tk.key, err)
		}
	}

	// Get due keys with limit 5
	dueKeys, err := db.GetDueKeys(5)
	if err != nil {
		t.Fatalf("GetDueKeys failed: %v", err)
	}

	if len(dueKeys) != 2 {
		t.Errorf("Expected 2 due keys, got %d", len(dueKeys))
	}

	// Check that we got the correct keys
	foundA, foundB := false, false
	for _, ks := range dueKeys {
		if ks.Key == "a" {
			foundA = true
		}
		if ks.Key == "b" {
			foundB = true
		}
		if ks.Key == "c" || ks.Key == "d" || ks.Key == "e" {
			t.Errorf("Got unexpected key: %s", ks.Key)
		}
	}

	if !foundA {
		t.Error("Expected to find key 'a' in due keys")
	}
	if !foundB {
		t.Error("Expected to find key 'b' in due keys")
	}

	// Test with limit 1
	dueKeys, err = db.GetDueKeys(1)
	if err != nil {
		t.Fatalf("GetDueKeys with limit 1 failed: %v", err)
	}

	if len(dueKeys) != 1 {
		t.Errorf("Expected 1 due key with limit, got %d", len(dueKeys))
	}

	// The oldest one should be "b" (practiced 7 days ago)
	if len(dueKeys) > 0 && dueKeys[0].Key != "b" {
		t.Errorf("Expected oldest due key to be 'b', got '%s'", dueKeys[0].Key)
	}
}

func TestUpdateKeyStatsWithSRS(t *testing.T) {
	tmpDB := "/tmp/kata_test_update_srs.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	target := "aaabbb"
	input := "aaabbb"

	err = db.UpdateKeyStats(target, input)
	if err != nil {
		t.Fatalf("UpdateKeyStats failed: %v", err)
	}

	allStats, err := db.GetAllKeyStats()
	if err != nil {
		t.Fatalf("GetAllKeyStats failed: %v", err)
	}

	if len(allStats) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(allStats))
	}

	for _, stat := range allStats {
		if stat.Key == "a" || stat.Key == "b" {
			if stat.Errors != 0 {
				t.Errorf("Key %s: expected 0 errors, got %d", stat.Key, stat.Errors)
			}
			if stat.Successes != 3 {
				t.Errorf("Key %s: expected 3 successes, got %d", stat.Key, stat.Successes)
			}
			if stat.Interval != 1 {
				t.Errorf("Key %s: expected interval 1 (first correct review), got %d", stat.Key, stat.Interval)
			}
			if stat.Repetitions != 1 {
				t.Errorf("Key %s: expected 1 repetition, got %d", stat.Key, stat.Repetitions)
			}
		}
	}

	err = db.UpdateKeyStats("aaa", "aab")
	if err != nil {
		t.Fatalf("UpdateKeyStats second call failed: %v", err)
	}

	statA, err := db.GetAllKeyStats()
	if err != nil {
		t.Fatalf("GetAllKeyStats after second update failed: %v", err)
	}

	var updatedA *KeyStat
	for i := range statA {
		if statA[i].Key == "a" {
			updatedA = &statA[i]
			break
		}
	}

	if updatedA == nil {
		t.Fatal("Key 'a' not found after second update")
	}

	if updatedA.Errors != 1 {
		t.Errorf("Expected 1 error for 'a', got %d", updatedA.Errors)
	}
	if updatedA.Successes != 5 {
		t.Errorf("Expected 5 successes for 'a', got %d", updatedA.Successes)
	}

	accuracy := float64(updatedA.Successes) / float64(updatedA.Errors+updatedA.Successes)
	if accuracy < 0.8 || accuracy > 0.9 {
		t.Logf("Accuracy for 'a': %.2f (5/6 = 0.83)", accuracy)
	}

	if updatedA.Interval < 1 {
		t.Errorf("Expected interval >= 1, got %d", updatedA.Interval)
	}
}

func TestSaveSession(t *testing.T) {
	tmpDB := "/tmp/kata_test_save_session.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	session := Session{
		Text:       "hello world",
		WPM:        45.5,
		Accuracy:   92.3,
		Duration:   60.0,
		ErrorCount: 3,
		Timestamp:  time.Now(),
	}

	err = db.SaveSession(session)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	sessions, err := db.GetRecentSessions(1)
	if err != nil {
		t.Fatalf("GetRecentSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.Text != "hello world" {
		t.Errorf("Expected text 'hello world', got '%s'", s.Text)
	}
	if s.WPM != 45.5 {
		t.Errorf("Expected WPM 45.5, got %.1f", s.WPM)
	}
	if s.Accuracy != 92.3 {
		t.Errorf("Expected accuracy 92.3, got %.1f", s.Accuracy)
	}
	if s.Duration != 60.0 {
		t.Errorf("Expected duration 60.0, got %.1f", s.Duration)
	}
	if s.ErrorCount != 3 {
		t.Errorf("Expected error count 3, got %d", s.ErrorCount)
	}
}

func TestGetRecentSessions(t *testing.T) {
	tmpDB := "/tmp/kata_test_recent_sessions.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	for i := 0; i < 5; i++ {
		session := Session{
			Text:       "test",
			WPM:        float64(40 + i*5),
			Accuracy:   90.0,
			Duration:   60.0,
			ErrorCount: i,
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		}
		if err := db.SaveSession(session); err != nil {
			t.Fatalf("SaveSession %d failed: %v", i, err)
		}
	}

	sessions, err := db.GetRecentSessions(3)
	if err != nil {
		t.Fatalf("GetRecentSessions failed: %v", err)
	}

	if len(sessions) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(sessions))
	}

	if sessions[0].WPM < sessions[1].WPM {
		t.Error("Sessions should be ordered DESC by timestamp (most recent first)")
	}

	if sessions[0].ErrorCount != 4 {
		t.Errorf("Most recent session should have 4 errors, got %d", sessions[0].ErrorCount)
	}
}

func TestGetSessionsForGraph(t *testing.T) {
	tmpDB := "/tmp/kata_test_graph_sessions.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	for i := 0; i < 10; i++ {
		session := Session{
			Text:       "test",
			WPM:        float64(30 + i*2),
			Accuracy:   85.0 + float64(i),
			Duration:   60.0,
			ErrorCount: i,
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		}
		if err := db.SaveSession(session); err != nil {
			t.Fatalf("SaveSession %d failed: %v", i, err)
		}
	}

	sessions, err := db.GetSessionsForGraph(5)
	if err != nil {
		t.Fatalf("GetSessionsForGraph failed: %v", err)
	}

	if len(sessions) != 5 {
		t.Fatalf("Expected 5 sessions, got %d", len(sessions))
	}

	if sessions[0].WPM > sessions[1].WPM {
		t.Error("Sessions should be ordered ASC by timestamp (oldest first)")
	}

	if sessions[0].WPM != 30.0 {
		t.Errorf("First session should have WPM 30.0, got %.1f", sessions[0].WPM)
	}
}

func TestGetAverageWPM(t *testing.T) {
	tmpDB := "/tmp/kata_test_avg_wpm.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	avgBefore, _ := db.GetAverageWPM()
	if avgBefore != 0 {
		t.Errorf("Expected 0 WPM before any sessions, got %.1f", avgBefore)
	}

	wpmValues := []float64{40.0, 50.0, 60.0}
	expectedAvg := 50.0

	now := time.Now()
	for i, wpm := range wpmValues {
		session := Session{
			Text:       "test",
			WPM:        wpm,
			Accuracy:   90.0,
			Duration:   60.0,
			ErrorCount: 0,
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
		}
		if err := db.SaveSession(session); err != nil {
			t.Fatalf("SaveSession failed: %v", err)
		}
	}

	avg, err := db.GetAverageWPM()
	if err != nil {
		t.Fatalf("GetAverageWPM failed: %v", err)
	}

	if avg != expectedAvg {
		t.Errorf("Expected average WPM %.1f, got %.1f", expectedAvg, avg)
	}
}

func TestGetWeakestKeys(t *testing.T) {
	tmpDB := "/tmp/kata_test_weakest.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	testData := []struct {
		key       string
		errors    int
		successes int
	}{
		{"a", 5, 5},
		{"b", 1, 9},
		{"c", 8, 2},
		{"d", 0, 10},
		{"e", 6, 2},
	}

	now := time.Now()
	for _, td := range testData {
		_, err = db.conn.Exec(`
			INSERT INTO key_stats (key, errors, successes, last_practiced, interval, repetitions, ease_factor)
			VALUES (?, ?, ?, ?, 0, 0, 2.5)
		`, td.key, td.errors, td.successes, now)
		if err != nil {
			t.Fatalf("Insert failed for key '%s': %v", td.key, err)
		}
	}

	weakKeys, err := db.GetWeakestKeys(3)
	if err != nil {
		t.Fatalf("GetWeakestKeys failed: %v", err)
	}

	if len(weakKeys) != 3 {
		t.Fatalf("Expected 3 keys, got %d", len(weakKeys))
	}

	if weakKeys[0].Key != "c" {
		t.Errorf("Expected 'c' as weakest (80%% error rate), got '%s'", weakKeys[0].Key)
	}

	if weakKeys[1].Key != "e" {
		t.Errorf("Expected 'e' as second weakest (75%% error rate), got '%s'", weakKeys[1].Key)
	}

	if weakKeys[2].Key != "a" {
		t.Errorf("Expected 'a' as third weakest (50%% error rate), got '%s'", weakKeys[2].Key)
	}
}

func TestGetAllKeyStats(t *testing.T) {
	tmpDB := "/tmp/kata_test_all_stats.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	allBefore, _ := db.GetAllKeyStats()
	if len(allBefore) != 0 {
		t.Errorf("Expected 0 stats initially, got %d", len(allBefore))
	}

	err = db.UpdateKeyStats("abc", "abc")
	if err != nil {
		t.Fatalf("UpdateKeyStats failed: %v", err)
	}

	allAfter, err := db.GetAllKeyStats()
	if err != nil {
		t.Fatalf("GetAllKeyStats failed: %v", err)
	}

	if len(allAfter) != 3 {
		t.Fatalf("Expected 3 key stats, got %d", len(allAfter))
	}

	keys := make(map[string]bool)
	for _, stat := range allAfter {
		keys[stat.Key] = true
		if stat.Errors != 0 {
			t.Errorf("Key '%s': expected 0 errors, got %d", stat.Key, stat.Errors)
		}
		if stat.Successes != 1 {
			t.Errorf("Key '%s': expected 1 success, got %d", stat.Key, stat.Successes)
		}
	}

	if !keys["a"] || !keys["b"] || !keys["c"] {
		t.Error("Expected to find keys 'a', 'b', 'c'")
	}
}

func TestAnalyzeErrors(t *testing.T) {
	cases := []struct {
		name            string
		target          string
		input           string
		expectedErrors  map[string]int
		expectedBigrams map[string]int
	}{
		{
			name:            "perfect match",
			target:          "hello",
			input:           "hello",
			expectedErrors:  map[string]int{},
			expectedBigrams: map[string]int{},
		},
		{
			name:            "single error",
			target:          "hello",
			input:           "hallo",
			expectedErrors:  map[string]int{"e": 1},
			expectedBigrams: map[string]int{"he": 1},
		},
		{
			name:            "multiple errors",
			target:          "test",
			input:           "txst",
			expectedErrors:  map[string]int{"e": 1},
			expectedBigrams: map[string]int{"te": 1},
		},
		{
			name:            "input shorter than target",
			target:          "hello",
			input:           "hel",
			expectedErrors:  map[string]int{},
			expectedBigrams: map[string]int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzeErrors(tc.target, tc.input)

			if len(analysis.CharErrors) != len(tc.expectedErrors) {
				t.Errorf("Expected %d char errors, got %d", len(tc.expectedErrors), len(analysis.CharErrors))
			}

			for char, count := range tc.expectedErrors {
				if analysis.CharErrors[char] != count {
					t.Errorf("Expected %d errors for '%s', got %d", count, char, analysis.CharErrors[char])
				}
			}

			for bigram, count := range tc.expectedBigrams {
				if analysis.BigramErrors[bigram] != count {
					t.Errorf("Expected %d errors for bigram '%s', got %d", count, bigram, analysis.BigramErrors[bigram])
				}
			}
		})
	}
}

func TestAccuracyToQuality(t *testing.T) {
	cases := []struct {
		accuracy        float64
		expectedQuality int
	}{
		{1.0, 5},
		{0.95, 5},
		{0.94, 4},
		{0.85, 4},
		{0.84, 3},
		{0.70, 3},
		{0.69, 2},
		{0.50, 2},
		{0.49, 1},
		{0.30, 1},
		{0.29, 0},
		{0.0, 0},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("accuracy_%.2f", tc.accuracy), func(t *testing.T) {
			quality := accuracyToQuality(tc.accuracy)
			if quality != tc.expectedQuality {
				t.Errorf("Accuracy %.2f: expected quality %d, got %d",
					tc.accuracy, tc.expectedQuality, quality)
			}
		})
	}
}

func TestMigrateSchema(t *testing.T) {
	tmpDB := "/tmp/kata_test_migration.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	conn, err := sql.Open("sqlite", tmpDB)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer conn.Close()

	_, err = conn.Exec(`
		CREATE TABLE key_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL,
			errors INTEGER DEFAULT 0,
			successes INTEGER DEFAULT 0,
			last_practiced DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(key)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create old schema: %v", err)
	}

	var count int
	err = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('key_stats') WHERE name='interval'`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check columns: %v", err)
	}
	if count != 0 {
		t.Fatal("Old schema should not have 'interval' column")
	}

	conn.Close()

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	err = db.conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('key_stats') WHERE name='interval'`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check migrated columns: %v", err)
	}
	if count != 1 {
		t.Error("Migration should have added 'interval' column")
	}

	err = db.conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('key_stats') WHERE name='repetitions'`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check repetitions column: %v", err)
	}
	if count != 1 {
		t.Error("Migration should have added 'repetitions' column")
	}

	err = db.conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('key_stats') WHERE name='ease_factor'`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check ease_factor column: %v", err)
	}
	if count != 1 {
		t.Error("Migration should have added 'ease_factor' column")
	}
}

func TestIndexesCreated(t *testing.T) {
	tmpDB := "/tmp/kata_test_indexes.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	rows, err := db.conn.Query(`SELECT name FROM sqlite_master WHERE type='index' AND sql IS NOT NULL`)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	indexes := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan index name: %v", err)
		}
		indexes[name] = true
	}

	expectedIndexes := []string{
		"idx_sessions_timestamp",
		"idx_key_stats_last_practiced",
		"idx_key_stats_attempts",
	}

	for _, idx := range expectedIndexes {
		if !indexes[idx] {
			t.Errorf("Expected index '%s' not found", idx)
		}
	}

	if len(indexes) < len(expectedIndexes) {
		t.Logf("Found indexes: %v", indexes)
	}
}

func TestIndexPerformance(t *testing.T) {
	tmpDB := "/tmp/kata_test_perf.db"
	os.Remove(tmpDB)
	defer os.Remove(tmpDB)

	db, err := NewDB(tmpDB)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now()
	for i := 0; i < 1000; i++ {
		session := Session{
			Text:       "test",
			WPM:        float64(30 + i%50),
			Accuracy:   85.0,
			Duration:   60.0,
			ErrorCount: i % 10,
			Timestamp:  now.Add(time.Duration(i) * time.Second),
		}
		if err := db.SaveSession(session); err != nil {
			t.Fatalf("SaveSession failed: %v", err)
		}
	}

	for i := 0; i < 100; i++ {
		key := string(rune('a' + (i % 26)))
		_, err := db.conn.Exec(`
			INSERT OR REPLACE INTO key_stats (key, errors, successes, last_practiced, interval, repetitions, ease_factor)
			VALUES (?, ?, ?, ?, ?, ?, 2.5)
		`, key, i%10, 10-i%10, now.Add(time.Duration(-i)*24*time.Hour), i%7, i%3)
		if err != nil {
			t.Fatalf("Insert key_stats failed: %v", err)
		}
	}

	start := time.Now()
	_, err = db.GetRecentSessions(20)
	if err != nil {
		t.Fatalf("GetRecentSessions failed: %v", err)
	}
	durationRecent := time.Since(start)

	start = time.Now()
	_, err = db.GetDueKeys(50)
	if err != nil {
		t.Fatalf("GetDueKeys failed: %v", err)
	}
	durationDue := time.Since(start)

	start = time.Now()
	_, err = db.GetSessionsForGraph(50)
	if err != nil {
		t.Fatalf("GetSessionsForGraph failed: %v", err)
	}
	durationGraph := time.Since(start)

	if durationRecent > 50*time.Millisecond {
		t.Errorf("GetRecentSessions took %v (expected < 50ms with index)", durationRecent)
	}

	if durationDue > 50*time.Millisecond {
		t.Errorf("GetDueKeys took %v (expected < 50ms with index)", durationDue)
	}

	if durationGraph > 50*time.Millisecond {
		t.Errorf("GetSessionsForGraph took %v (expected < 50ms with index)", durationGraph)
	}

	t.Logf("Performance: GetRecentSessions=%v, GetDueKeys=%v, GetSessionsForGraph=%v",
		durationRecent, durationDue, durationGraph)
}
