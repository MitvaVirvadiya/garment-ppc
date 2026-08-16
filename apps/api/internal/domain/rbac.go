package domain

import "strings"

type Session struct {
	User        User     `json:"user"`
	CanPlace    bool     `json:"can_place"`
	CanMove     bool     `json:"can_move"`
	CanAdvance  bool     `json:"can_advance"`
	LockedPlant *string  `json:"locked_plant_id"`
	FocusBuyer  string   `json:"focus_buyer"`
	KPISet      string   `json:"kpi_set"`
}

func SessionFor(user User) Session {
	s := Session{User: user}
	switch user.Role {
	case RolePlantHead:
		s.CanPlace = true
		s.CanMove = true
		s.CanAdvance = true
		s.LockedPlant = user.PlantID
		s.KPISet = "plant"
	case RoleMerchandiser:
		s.CanPlace = true
		s.CanMove = true
		s.CanAdvance = true
		s.FocusBuyer = merchFocus(user.Slug)
		s.KPISet = "merch"
	case RoleViewer:
		s.CanPlace = false
		s.CanMove = false
		s.CanAdvance = false
		s.LockedPlant = user.PlantID
		s.KPISet = "viewer"
	}
	return s
}

func merchFocus(slug string) string {
	if strings.EqualFold(slug, "rohan") {
		return "Rohan Desai"
	}
	return ""
}

func CanAccessPlant(s Session, plantID string) bool {
	if s.LockedPlant == nil || *s.LockedPlant == "" {
		return true
	}
	return *s.LockedPlant == plantID
}

func CanMutateLine(s Session, linePlantID string) bool {
	if s.User.Role == RoleViewer {
		return false
	}
	if s.User.Role == RolePlantHead {
		return CanAccessPlant(s, linePlantID)
	}
	return true
}
