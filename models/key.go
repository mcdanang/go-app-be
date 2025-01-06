package models

type Key struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StaffID     int    `json:"staff_id"`
}
