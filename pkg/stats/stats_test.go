package stats

import (
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
