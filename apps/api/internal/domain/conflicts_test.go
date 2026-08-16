package domain

import (
	"testing"
	"time"
)

func TestDetectConflictsOverlapAndShift(t *testing.T) {
	loc := time.FixedZone("IST", 5*3600+30*60)
	day := time.Date(2026, 8, 18, 0, 0, 0, 0, loc)
	plants := []Plant{{ID: "p1", Code: "TK", Timezone: "Asia/Kolkata", ShiftStart: "08:00", ShiftEnd: "17:00", ShiftMinutes: 480}}
	lines := []Line{{
		ID: "l1", Code: "TK-S3", PlantID: "p1", PlantCode: "TK", ShiftMinutes: 480,
	}}
	blocks := []Block{
		{ID: "b1", LineID: "l1", POCode: "TK-PO-1", StartsAt: day.Add(8 * time.Hour), EndsAt: day.Add(17 * time.Hour)},
		{ID: "b2", LineID: "l1", POCode: "TK-PO-2", StartsAt: day.Add(12 * time.Hour), EndsAt: day.Add(20 * time.Hour)},
	}
	cs := DetectConflicts(lines, plants, blocks)
	if len(cs) < 2 {
		t.Fatalf("expected overlap + over_shift, got %#v", cs)
	}
	kinds := map[string]bool{}
	for _, c := range cs {
		kinds[c.Kind] = true
	}
	if !kinds["overlap"] || !kinds["over_shift"] {
		t.Fatalf("missing kinds: %#v", cs)
	}
}

func TestSingleShiftBlockIsNotOverflow(t *testing.T) {
	loc := time.FixedZone("IST", 5*3600+30*60)
	day := time.Date(2026, 8, 18, 0, 0, 0, 0, loc)
	plants := []Plant{{ID: "p1", Code: "TK", Timezone: "Asia/Kolkata", ShiftStart: "08:00", ShiftEnd: "17:00", ShiftMinutes: 480}}
	lines := []Line{{ID: "l1", Code: "TK-S1", PlantID: "p1", PlantCode: "TK", ShiftMinutes: 480}}
	blocks := []Block{
		{ID: "b1", LineID: "l1", POCode: "TK-PO-1", StartsAt: day.Add(8 * time.Hour), EndsAt: day.AddDate(0, 0, 3).Add(17 * time.Hour)},
	}
	cs := DetectConflicts(lines, plants, blocks)
	for _, c := range cs {
		if c.Kind == "over_shift" {
			t.Fatalf("full-shift multi-day block should not overflow: %#v", c)
		}
	}
}
