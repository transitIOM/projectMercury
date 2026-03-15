package models

type ServiceAlert struct {
	ID               string   `json:"id"`
	Header           string   `json:"header"`
	Description      string   `json:"description"`
	AffectedEntities []string `json:"affected_entities"`
}
