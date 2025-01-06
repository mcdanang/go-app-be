package controllers

import (
	"database/sql"
	"encoding/json"
	"go-app-be/models"
	"log"
	"net/http"
)

// Get all staff
func GetStaff(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM staff")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var staffMembers []models.Staff
		for rows.Next() {
			var s models.Staff
			if err := rows.Scan(&s.ID, &s.Name, &s.Role); err != nil {
				log.Fatal(err)
			}
			staffMembers = append(staffMembers, s)
		}

		json.NewEncoder(w).Encode(staffMembers)
	}
}

// Create a staff member
func CreateStaff(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.Staff
		json.NewDecoder(r.Body).Decode(&s)

		err := db.QueryRow("INSERT INTO staff (name, role) VALUES ($1, $2) RETURNING id", s.Name, s.Role).Scan(&s.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(s)
	}
}
