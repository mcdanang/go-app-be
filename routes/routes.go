package routes

import (
	"database/sql"
	"go-app-be/controllers"

	"github.com/gorilla/mux"
)

// SetupRoutes sets up all the routes for the application
func SetupRoutes(router *mux.Router, db *sql.DB) {
	// Key Routes
	router.HandleFunc("/keys", controllers.GetKeys(db)).Methods("GET", "OPTIONS")
	router.HandleFunc("/keys/{id}", controllers.GetKey(db)).Methods("GET", "OPTIONS")
	router.HandleFunc("/keys", controllers.CreateKey(db)).Methods("POST", "OPTIONS")
	router.HandleFunc("/keys/{id}", controllers.UpdateKey(db)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/keys/{id}", controllers.DeleteKey(db)).Methods("DELETE", "OPTIONS")

	// Key Copy Routes
	router.HandleFunc("/key-copies", controllers.GetKeyCopies(db)).Methods("GET", "OPTIONS")
	router.HandleFunc("/key-copies", controllers.CreateKeyCopy(db)).Methods("POST", "OPTIONS")
	router.HandleFunc("/key-copies/{id}", controllers.UpdateKeyCopy(db)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/key-copies/{id}", controllers.DeleteKeyCopy(db)).Methods("DELETE", "OPTIONS")

	// Staff Routes
	router.HandleFunc("/staffs", controllers.GetStaffs(db)).Methods("GET", "OPTIONS")
	router.HandleFunc("/staffs", controllers.CreateStaff(db)).Methods("POST", "OPTIONS")
	router.HandleFunc("/staffs/{id}", controllers.UpdateStaff(db)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/staffs/{id}", controllers.DeleteStaff(db)).Methods("DELETE", "OPTIONS")
}
