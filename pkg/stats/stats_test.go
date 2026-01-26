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
