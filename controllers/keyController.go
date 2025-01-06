package controllers

import (
	"database/sql"
	"encoding/json"
	"go-app-be/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PaginatedResponse struct {
	Data       []models.Key `json:"data"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}

// Get all keys with pagination and name filter
func GetKeys(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil || pageSize <= 0 {
			pageSize = 3
		}

		nameFilter := r.URL.Query().Get("name")
		offset := (page - 1) * pageSize

		whereClause := "WHERE 1=1"
		var queryParams []interface{}

		if nameFilter != "" {
			whereClause += " AND name ILIKE $1"
			nameParam := "%" + nameFilter + "%"
			queryParams = append(queryParams, nameParam)
		}

		countQuery := "SELECT COUNT(*) FROM keys " + whereClause
		var total int
		err = db.QueryRow(countQuery, queryParams...).Scan(&total)
		if err != nil {
			log.Printf("Error counting records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		selectQuery := "SELECT id, name, description, staff_id FROM keys " + whereClause
		selectQuery += " ORDER BY id LIMIT $" + strconv.Itoa(len(queryParams)+1) + " OFFSET $" + strconv.Itoa(len(queryParams)+2)
		queryParams = append(queryParams, pageSize, offset)

		rows, err := db.Query(selectQuery, queryParams...)
		if err != nil {
			log.Printf("Error querying records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var keys []models.Key
		for rows.Next() {
			var k models.Key
			if err := rows.Scan(&k.ID, &k.Name, &k.Description, &k.StaffID); err != nil {
				log.Printf("Error scanning row: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			keys = append(keys, k)
		}

		response := PaginatedResponse{
			Data:       keys,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// Get a specific key by ID
func GetKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var k models.Key
		err := db.QueryRow("SELECT id, name, description, staff_id FROM keys WHERE id = $1", id).Scan(&k.ID, &k.Name, &k.Description, &k.StaffID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Key not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving key: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		json.NewEncoder(w).Encode(k)
	}
}

// Create a new key
func CreateKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var k models.Key
		if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify staff exists if staff_id is provided
		if k.StaffID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM staff WHERE id = $1)", k.StaffID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking staff existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Staff ID does not exist", http.StatusBadRequest)
				return
			}
		}

		err := db.QueryRow(
			"INSERT INTO keys (name, description, staff_id) VALUES ($1, $2, $3) RETURNING id",
			k.Name, k.Description, k.StaffID,
		).Scan(&k.ID)

		if err != nil {
			log.Printf("Error creating key: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(k)
	}
}

// Update a key
func UpdateKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var k models.Key
		if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify key exists
		var existingKey models.Key
		err := db.QueryRow(
			"SELECT id, name, description, staff_id FROM keys WHERE id = $1",
			id,
		).Scan(&existingKey.ID, &existingKey.Name, &existingKey.Description, &existingKey.StaffID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Key not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving key: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Verify staff exists if staff_id is provided
		if k.StaffID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM staff WHERE id = $1)", k.StaffID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking staff existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Staff ID does not exist", http.StatusBadRequest)
				return
			}
		}

		_, err = db.Exec(
			"UPDATE keys SET name = $1, description = $2, staff_id = $3 WHERE id = $4",
			k.Name, k.Description, k.StaffID, id,
		)

		if err != nil {
			log.Printf("Error updating key: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		k.ID = existingKey.ID
		json.NewEncoder(w).Encode(k)
	}
}

// Delete a key
func DeleteKey(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var existingKey models.Key
		err := db.QueryRow(
			"SELECT id, name, description, staff_id FROM keys WHERE id = $1",
			id,
		).Scan(&existingKey.ID, &existingKey.Name, &existingKey.Description, &existingKey.StaffID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Key not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving key: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.Exec("DELETE FROM keys WHERE id = $1", id)
		if err != nil {
			log.Printf("Error deleting key: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Key deleted successfully"})
	}
}
