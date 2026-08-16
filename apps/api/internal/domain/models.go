package domain

import "time"

const (
	RolePlantHead    = "plant_head"
	RoleMerchandiser = "merchandiser"
	RoleViewer       = "viewer"
)

var SOPSequence = []string{
	"released",
	"marker",
	"cut",
	"loaded",
	"sewing",
	"mid_qc",
	"finishing",
	"packed",
	"ex_factory",
}

type Plant struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Region        string `json:"region"`
	Timezone      string `json:"timezone"`
	ShiftStart    string `json:"shift_start"`
	ShiftEnd      string `json:"shift_end"`
	ShiftMinutes  int    `json:"shift_minutes"`
}

type Unit struct {
	ID      string `json:"id"`
	PlantID string `json:"plant_id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
}

type Line struct {
	ID            string `json:"id"`
	UnitID        string `json:"unit_id"`
	PlantID       string `json:"plant_id"`
	PlantCode     string `json:"plant_code"`
	UnitCode      string `json:"unit_code"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Operators     int    `json:"operators"`
	SAMMinutes    float64 `json:"sam_minutes"`
	ShiftMinutes  int    `json:"shift_minutes"`
}

type User struct {
	ID       string  `json:"id"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	PlantID  *string `json:"plant_id"`
	Title    string  `json:"title"`
}

type Order struct {
	ID               string  `json:"id"`
	POCode           string  `json:"po_code"`
	StyleCode        string  `json:"style_code"`
	StyleName        string  `json:"style_name"`
	Buyer            string  `json:"buyer"`
	PlantID          string  `json:"plant_id"`
	PlantCode        string  `json:"plant_code"`
	Qty              int     `json:"qty"`
	DueDate          string  `json:"due_date"`
	Status           string  `json:"status"`
	MinutesPerPiece  float64 `json:"minutes_per_piece"`
	Colorway         string  `json:"colorway"`
	MerchOwner       string  `json:"merch_owner"`
	PlacedQty        int     `json:"placed_qty"`
	UnplacedQty      int     `json:"unplaced_qty"`
}

type Block struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	LineID        string    `json:"line_id"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	PlannedQty    int       `json:"planned_qty"`
	SOPState      string    `json:"sop_state"`
	Pace          float64   `json:"pace"`
	ActualPct     float64   `json:"actual_pct"`
	PlannedPct    float64   `json:"planned_pct"`
	PiecesDone    int       `json:"pieces_done"`
	DelayMinutes  int       `json:"delay_minutes"`
	Notes         string    `json:"notes"`
	POCode        string    `json:"po_code"`
	StyleName     string    `json:"style_name"`
	Buyer         string    `json:"buyer"`
	Colorway      string    `json:"colorway"`
	MerchOwner    string    `json:"merch_owner"`
	LineCode      string    `json:"line_code"`
	LineKind      string    `json:"line_kind"`
	PlantID       string    `json:"plant_id"`
	PlantCode     string    `json:"plant_code"`
	ConflictIDs   []string  `json:"conflict_ids"`
	HasConflict   bool      `json:"has_conflict"`
}

type SOPEntry struct {
	ID        string    `json:"id"`
	BlockID   string    `json:"block_id"`
	State     string    `json:"state"`
	EnteredAt time.Time `json:"entered_at"`
	Actor     string    `json:"actor"`
}

type Tick struct {
	ID            string    `json:"id"`
	BlockID       string    `json:"block_id"`
	ActualPct     float64   `json:"actual_pct"`
	PlannedPct    float64   `json:"planned_pct"`
	PiecesDone    int       `json:"pieces_done"`
	DelayMinutes  int       `json:"delay_minutes"`
	TickedAt      time.Time `json:"ticked_at"`
}

type Conflict struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	LineID      string    `json:"line_id"`
	LineCode    string    `json:"line_code"`
	PlantID     string    `json:"plant_id"`
	PlantCode   string    `json:"plant_code"`
	BlockIDs    []string  `json:"block_ids"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	MinutesOver int       `json:"minutes_over"`
	Message     string    `json:"message"`
}

func PlannedPct(now, start, end time.Time) float64 {
	if !end.After(start) {
		return 0
	}
	if now.Before(start) {
		return 0
	}
	if !now.Before(end) {
		return 100
	}
	return (float64(now.Sub(start)) / float64(end.Sub(start))) * 100
}

func DelayMinutes(planned, actual float64, start, end time.Time) int {
	if planned <= actual {
		return 0
	}
	dur := end.Sub(start).Minutes()
	if dur <= 0 {
		return 0
	}
	return int((planned - actual) / 100 * dur)
}

func NextSOP(current string) (string, bool) {
	for i, s := range SOPSequence {
		if s == current && i+1 < len(SOPSequence) {
			return SOPSequence[i+1], true
		}
	}
	return current, false
}

func SOPIndex(state string) int {
	for i, s := range SOPSequence {
		if s == state {
			return i
		}
	}
	return -1
}

func IsActiveSOP(state string) bool {
	switch state {
	case "loaded", "sewing", "mid_qc", "finishing":
		return true
	default:
		return false
	}
}
