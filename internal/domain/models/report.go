package models

type UserReport struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Category    string `json:"category"`
}
