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

// Get all staffs with pagination and name filter
func GetStaffs(db *sql.DB) http.HandlerFunc {
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
			whereClause += " AND staffs.name ILIKE $1"
			nameParam := "%" + nameFilter + "%"
			queryParams = append(queryParams, nameParam)
		}

		countQuery := "SELECT COUNT(*) FROM staffs " + whereClause
		var total int
		err = db.QueryRow(countQuery, queryParams...).Scan(&total)
		if err != nil {
			log.Printf("Error counting records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		// Join with the staffs table to get the staff_name
		selectQuery := `
			SELECT staffs.id, staffs.name, staffs.role
			FROM staffs
		` + whereClause + `
			ORDER BY staffs.id 
			LIMIT $` + strconv.Itoa(len(queryParams)+1) + ` OFFSET $` + strconv.Itoa(len(queryParams)+2)
		queryParams = append(queryParams, pageSize, offset)

		rows, err := db.Query(selectQuery, queryParams...)
		if err != nil {
			log.Printf("Error querying records: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var staffs []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Role string `json:"role"`
		}

		for rows.Next() {
			var s struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Role string `json:"role"`
			}
			if err := rows.Scan(&s.ID, &s.Name, &s.Role); err != nil {
				log.Printf("Error scanning row: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			staffs = append(staffs, s)
		}

		response := struct {
			Data       interface{} `json:"data"`
			Total      int         `json:"total"`
			Page       int         `json:"page"`
			PageSize   int         `json:"pageSize"`
			TotalPages int         `json:"totalPages"`
		}{
			Data:       staffs,
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

// Create a staff member
func CreateStaff(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.Staff
		json.NewDecoder(r.Body).Decode(&s)

		err := db.QueryRow("INSERT INTO staffs (name, role) VALUES ($1, $2) RETURNING id", s.Name, s.Role).Scan(&s.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(s)
	}
}

// Update a staff member
func UpdateStaff(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var s models.Staff
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify staff exists
		var existingStaff models.Staff
		err := db.QueryRow(
			"SELECT id, name, role FROM staffs WHERE id = $1",
			id,
		).Scan(&existingStaff.ID, &existingStaff.Name, &existingStaff.Role)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Staff not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving staff: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.Exec(
			"UPDATE staffs SET name = $1, role = $2 WHERE id = $3",
			s.Name, s.Role, id,
		)

		if err != nil {
			log.Printf("Error updating key: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		s.ID = existingStaff.ID
		json.NewEncoder(w).Encode(s)
	}
}

// Delete a staff member
func DeleteStaff(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var existingStaff models.Staff
		err := db.QueryRow(
			"SELECT id, name, role FROM staffs WHERE id = $1",
			id,
		).Scan(&existingStaff.ID, &existingStaff.Name, &existingStaff.Role)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Staff not found", http.StatusNotFound)
			} else {
				log.Printf("Error retrieving staff: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.Exec("DELETE FROM staffs WHERE id = $1", id)
		if err != nil {
			log.Printf("Error deleting staff: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Staff deleted successfully"})
	}
}
