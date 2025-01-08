package entities

type EventGoogle struct {
	Summary     string `json:"summary"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Start       string `json:"start"` // ISO 8601 format, e.g., "2025-01-03T10:00:00+07:00"
	End         string `json:"end"`   // ISO 8601 format, e.g., "2025-01-03T11:00:00+07:00"
}
