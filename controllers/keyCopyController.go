package controllers

import (
	"database/sql"
	"encoding/json"
	"go-app-be/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Get all key copies
func GetKeyCopies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM key_copies")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var keyCopies []models.KeyCopy
		for rows.Next() {
			var kc models.KeyCopy
			if err := rows.Scan(&kc.ID, &kc.KeyID, &kc.StaffID); err != nil {
				log.Fatal(err)
			}
			keyCopies = append(keyCopies, kc)
		}

		json.NewEncoder(w).Encode(keyCopies)
	}
}

// Create key copy
func CreateKeyCopy(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var kc models.KeyCopy
		json.NewDecoder(r.Body).Decode(&kc)

		err := db.QueryRow("INSERT INTO key_copies (key_id, staff_id) VALUES ($1, $2) RETURNING id", kc.KeyID, kc.StaffID).Scan(&kc.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(kc)
	}
}

// Delete a key copy
func DeleteKeyCopy(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var existingKey models.KeyCopy
		err := db.QueryRow("SELECT * FROM key_copies WHERE id = $1", id).Scan(&existingKey.ID, &existingKey.KeyID, &existingKey.StaffID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM key_copies WHERE id = $1", id)
			if err != nil {
				log.Fatal(err)
			}
		}

		json.NewEncoder(w).Encode("Key Copy deleted")

	}
}
