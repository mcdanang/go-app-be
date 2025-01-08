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

type PaginatedResponseKeyCopy struct {
	Data []struct {
		models.KeyCopy
		KeyName string `json:"key_name"`
	} `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

// Get all key copies with pagination and key_name filter
func GetKeyCopies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Pagination parameters
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

		// Construct WHERE clause and parameters
		whereClause := "WHERE 1=1"
		var queryParams []interface{}

		if nameFilter != "" {
			whereClause += " AND LOWER(k.name) LIKE $1"
			queryParams = append(queryParams, "%"+nameFilter+"%")
		}

		// Count query
		countQuery := `
			SELECT COUNT(*)
			FROM key_copies kc
			JOIN keys k ON kc.key_id = k.id
			JOIN staffs s ON kc.staff_id = s.id
			` + whereClause

		var total int
		err = db.QueryRow(countQuery, queryParams...).Scan(&total)
		if err != nil {
			log.Printf("Error counting records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		// Data query with JOINs
		selectQuery := `
			SELECT kc.id, kc.key_id, k.name AS key_name, kc.staff_id, s.name AS staff_name
			FROM key_copies kc
			JOIN keys k ON kc.key_id = k.id
			JOIN staffs s ON kc.staff_id = s.id
			` + whereClause + `
			ORDER BY kc.id
			LIMIT $` + strconv.Itoa(len(queryParams)+1) + ` OFFSET $` + strconv.Itoa(len(queryParams)+2)

		queryParams = append(queryParams, pageSize, offset)

		rows, err := db.Query(selectQuery, queryParams...)
		if err != nil {
			log.Printf("Error querying records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Scan results
		type KeyCopyWithNames struct {
			models.KeyCopy
			KeyName   string `json:"key_name"`
			StaffName string `json:"staff_name"`
		}

		var keyCopies []KeyCopyWithNames

		for rows.Next() {
			var kCopy models.KeyCopy
			var keyName, staffName string
			if err := rows.Scan(&kCopy.ID, &kCopy.KeyID, &keyName, &kCopy.StaffID, &staffName); err != nil {
				log.Printf("Error scanning row: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			keyCopies = append(keyCopies, KeyCopyWithNames{
				KeyCopy:   kCopy,
				KeyName:   keyName,
				StaffName: staffName,
			})
		}

		response := struct {
			Data       interface{} `json:"data"`
			Total      int         `json:"total"`
			Page       int         `json:"page"`
			PageSize   int         `json:"pageSize"`
			TotalPages int         `json:"totalPages"`
		}{
			Data:       keyCopies,
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

// // Get a specific key by ID
// func GetKey(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		id := vars["id"]

// 		var k models.Key
// 		err := db.QueryRow("SELECT id, name, description, staff_id FROM keys WHERE id = $1", id).Scan(&k.ID, &k.Name, &k.Description, &k.StaffID)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				http.Error(w, "Key not found", http.StatusNotFound)
// 			} else {
// 				log.Printf("Error retrieving key: %v", err)
// 				http.Error(w, "Internal server error", http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		json.NewEncoder(w).Encode(k)
// 	}
// }

// Create a new key copy
func CreateKeyCopy(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var k models.KeyCopy
		if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify staff exists if staff_id is provided
		if k.StaffID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM staffs WHERE id = $1)", k.StaffID).Scan(&exists)
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

		// Verify key exists if key_id is provided
		if k.KeyID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM keys WHERE id = $1)", k.KeyID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking key existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Key ID does not exist", http.StatusBadRequest)
				return
			}
		}

		err := db.QueryRow(
			"INSERT INTO key_copies (key_id, staff_id) VALUES ($1, $2) RETURNING id",
			k.KeyID, k.StaffID,
		).Scan(&k.ID)

		if err != nil {
			log.Printf("Error creating key copy: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(k)
	}
}

// Update a key copy
func UpdateKeyCopy(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var k models.KeyCopy
		if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify key exists
		var existingKeyCopy models.KeyCopy
		err := db.QueryRow(
			"SELECT id, key_id, staff_id FROM key_copies WHERE id = $1",
			id,
		).Scan(&existingKeyCopy.ID, &existingKeyCopy.KeyID, &existingKeyCopy.StaffID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Key Copy not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving key copy: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Verify staff exists if staff_id is provided
		if k.StaffID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM staffs WHERE id = $1)", k.StaffID).Scan(&exists)
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

		// Verify key exists if key_id is provided
		if k.KeyID != 0 {
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM keys WHERE id = $1)", k.KeyID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking key existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Key ID does not exist", http.StatusBadRequest)
				return
			}
		}

		_, err = db.Exec(
			"UPDATE key_copies SET key_id = $1, staff_id = $2 WHERE id = $3",
			k.KeyID, k.StaffID, id,
		)

		if err != nil {
			log.Printf("Error updating key copy: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		k.ID = existingKeyCopy.ID
		json.NewEncoder(w).Encode(k)
	}
}

// Delete a key copy
func DeleteKeyCopy(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var existingKeyCopy models.KeyCopy
		err := db.QueryRow(
			"SELECT id, key_id, staff_id FROM key_copies WHERE id = $1",
			id,
		).Scan(&existingKeyCopy.ID, &existingKeyCopy.KeyID, &existingKeyCopy.StaffID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Key copy not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving key copy: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.Exec("DELETE FROM key_copies WHERE id = $1", id)
		if err != nil {
			log.Printf("Error deleting key copy: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Key copy deleted successfully"})
	}
}
