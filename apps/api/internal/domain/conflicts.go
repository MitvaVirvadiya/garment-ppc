package domain

import (
	"fmt"
	"sort"
	"time"
)

type lineDayKey struct {
	LineID string
	Day    string
}

type occupy struct {
	BlockID  string
	Start    time.Time
	End      time.Time
	Minutes  int
}

func DetectConflicts(lines []Line, plants []Plant, blocks []Block) []Conflict {
	lineByID := map[string]Line{}
	for _, l := range lines {
		lineByID[l.ID] = l
	}
	plantByID := map[string]Plant{}
	for _, p := range plants {
		plantByID[p.ID] = p
	}

	var out []Conflict

	byLine := map[string][]Block{}
	for _, b := range blocks {
		byLine[b.LineID] = append(byLine[b.LineID], b)
	}
	for lineID, set := range byLine {
		sort.Slice(set, func(i, j int) bool { return set[i].StartsAt.Before(set[j].StartsAt) })
		for i := 0; i < len(set); i++ {
			for j := i + 1; j < len(set); j++ {
				a, b := set[i], set[j]
				if !b.StartsAt.Before(a.EndsAt) {
					break
				}
				if overlaps(a.StartsAt, a.EndsAt, b.StartsAt, b.EndsAt) {
					start := maxTime(a.StartsAt, b.StartsAt)
					end := minTime(a.EndsAt, b.EndsAt)
					line := lineByID[lineID]
					out = append(out, Conflict{
						ID:        fmt.Sprintf("overlap:%s:%s", short(a.ID), short(b.ID)),
						Kind:      "overlap",
						LineID:    lineID,
						LineCode:  line.Code,
						PlantID:   line.PlantID,
						PlantCode: line.PlantCode,
						BlockIDs:  []string{a.ID, b.ID},
						StartsAt:  start,
						EndsAt:    end,
						Message:   fmt.Sprintf("%s warp snag: %s and %s share %s", line.Code, a.POCode, b.POCode, line.Code),
					})
				}
			}
		}
	}

	// Minutes inside the plant shift window vs line shift capacity.
	// Multi-day blocks do not bill the night; a full shift equals ShiftMinutes.
	load := map[lineDayKey][]occupy{}
	for _, b := range blocks {
		line, ok := lineByID[b.LineID]
		if !ok {
			continue
		}
		plant := plantByID[line.PlantID]
		loc := locationOrUTC(plant.Timezone)
		cur := b.StartsAt.In(loc)
		end := b.EndsAt.In(loc)
		for day := truncateDay(cur); !day.After(end); day = day.AddDate(0, 0, 1) {
			shiftStart, shiftEnd := shiftWindow(day, plant)
			clock := int(shiftEnd.Sub(shiftStart).Minutes())
			if clock <= 0 {
				clock = line.ShiftMinutes
			}
			segStart := maxTime(b.StartsAt.In(loc), shiftStart)
			segEnd := minTime(b.EndsAt.In(loc), shiftEnd)
			clipped := int(segEnd.Sub(segStart).Minutes())
			if clipped <= 0 {
				continue
			}
			booked := clipped * line.ShiftMinutes / clock
			if booked < 1 {
				booked = 1
			}
			key := lineDayKey{LineID: b.LineID, Day: day.Format("2006-01-02")}
			load[key] = append(load[key], occupy{BlockID: b.ID, Start: segStart, End: segEnd, Minutes: booked})
		}
	}
	for key, segs := range load {
		line := lineByID[key.LineID]
		total := 0
		ids := make([]string, 0, len(segs))
		var start, end time.Time
		for i, s := range segs {
			total += s.Minutes
			ids = append(ids, s.BlockID)
			if i == 0 || s.Start.Before(start) {
				start = s.Start
			}
			if i == 0 || s.End.After(end) {
				end = s.End
			}
		}
		if total > line.ShiftMinutes {
			out = append(out, Conflict{
				ID:          fmt.Sprintf("shift:%s:%s", short(key.LineID), key.Day),
				Kind:        "over_shift",
				LineID:      key.LineID,
				LineCode:    line.Code,
				PlantID:     line.PlantID,
				PlantCode:   line.PlantCode,
				BlockIDs:    unique(ids),
				StartsAt:    start,
				EndsAt:      end,
				MinutesOver: total - line.ShiftMinutes,
				Message:     fmt.Sprintf("%s pick overflow: %d min booked vs %d min shift on %s", line.Code, total, line.ShiftMinutes, key.Day),
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].StartsAt.Equal(out[j].StartsAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].StartsAt.Before(out[j].StartsAt)
	})
	return out
}

func ApplyConflicts(blocks []Block, conflicts []Conflict) []Block {
	hit := map[string][]string{}
	for _, c := range conflicts {
		for _, id := range c.BlockIDs {
			hit[id] = append(hit[id], c.ID)
		}
	}
	out := make([]Block, len(blocks))
	copy(out, blocks)
	for i := range out {
		ids := hit[out[i].ID]
		out[i].ConflictIDs = ids
		out[i].HasConflict = len(ids) > 0
	}
	return out
}

func overlaps(a0, a1, b0, b1 time.Time) bool {
	return a0.Before(b1) && b0.Before(a1)
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func shiftWindow(day time.Time, plant Plant) (time.Time, time.Time) {
	sh, sm := parseHM(plant.ShiftStart, 8, 0)
	eh, em := parseHM(plant.ShiftEnd, 17, 0)
	start := time.Date(day.Year(), day.Month(), day.Day(), sh, sm, 0, 0, day.Location())
	end := time.Date(day.Year(), day.Month(), day.Day(), eh, em, 0, 0, day.Location())
	if !end.After(start) {
		end = start.Add(8 * time.Hour)
	}
	return start, end
}

func parseHM(raw string, dh, dm int) (int, int) {
	var h, m int
	if _, err := fmt.Sscanf(raw, "%d:%d", &h, &m); err != nil {
		return dh, dm
	}
	return h, m
}

func locationOrUTC(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func short(id string) string {
	if len(id) < 8 {
		return id
	}
	return id[:8]
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
