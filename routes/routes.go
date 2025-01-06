package routes

import (
	"database/sql"
	"go-app-be/controllers"

	"github.com/gorilla/mux"
)

// SetupRoutes sets up all the routes for the application
func SetupRoutes(router *mux.Router, db *sql.DB) {
	// Key Routes
	router.HandleFunc("/keys", controllers.GetKeys(db)).Methods("GET")
	router.HandleFunc("/keys/{id}", controllers.GetKey(db)).Methods("GET")
	router.HandleFunc("/keys", controllers.CreateKey(db)).Methods("POST")
	router.HandleFunc("/keys/{id}", controllers.UpdateKey(db)).Methods("PUT")
	router.HandleFunc("/keys/{id}", controllers.DeleteKey(db)).Methods("DELETE")

	// Key Copy Routes
	router.HandleFunc("/key_copies", controllers.GetKeyCopies(db)).Methods("GET")
	router.HandleFunc("/key_copies", controllers.CreateKeyCopy(db)).Methods("POST")
	router.HandleFunc("/key_copies/{id}", controllers.DeleteKeyCopy(db)).Methods("DELETE")

	// Staff Routes
	router.HandleFunc("/staff", controllers.GetStaff(db)).Methods("GET")
	router.HandleFunc("/staff", controllers.CreateStaff(db)).Methods("POST")
}
