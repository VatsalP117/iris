package api

import (
	"testing"
	"time"
)

func TestPreviousPeriodMatchesCurrentDuration(t *testing.T) {
	previousFrom, previousTo, ok := previousPeriod("2026-07-01", "2026-07-07")
	if !ok {
		t.Fatal("previousPeriod returned ok=false")
	}

	start, err := time.Parse(time.RFC3339Nano, previousFrom)
	if err != nil {
		t.Fatalf("failed to parse previous start: %v", err)
	}
	end, err := time.Parse(time.RFC3339Nano, previousTo)
	if err != nil {
		t.Fatalf("failed to parse previous end: %v", err)
	}

	if start.Format("2006-01-02") != "2026-06-24" {
		t.Fatalf("unexpected previous start: %s", start)
	}
	if end.Format("2006-01-02") != "2026-06-30" {
		t.Fatalf("unexpected previous end: %s", end)
	}
}

func TestPercentChange(t *testing.T) {
	if got := percentChange(125, 100); got != 25 {
		t.Fatalf("percentChange returned %v, want 25", got)
	}
	if got := percentChange(10, 0); got != 100 {
		t.Fatalf("percentChange with zero baseline returned %v, want 100", got)
	}
	if got := percentChange(0, 0); got != 0 {
		t.Fatalf("percentChange with empty periods returned %v, want 0", got)
	}
}
