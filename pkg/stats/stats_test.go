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
