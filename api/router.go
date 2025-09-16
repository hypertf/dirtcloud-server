package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hypertf/dirtcloud-server/web"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Web console routes
	webHandler := web.NewHandler(handler.service)
	webRouter := router.PathPrefix("/web").Subrouter()
	
	// Dashboard
	webRouter.HandleFunc("", webHandler.Dashboard).Methods("GET")
	webRouter.HandleFunc("/", webHandler.Dashboard).Methods("GET")
	
	// Project routes
	webRouter.HandleFunc("/projects", webHandler.ListProjects).Methods("GET")
	webRouter.HandleFunc("/projects", webHandler.CreateProject).Methods("POST")
	webRouter.HandleFunc("/projects/new", webHandler.NewProjectForm).Methods("GET")
	webRouter.HandleFunc("/projects/{id}/edit", webHandler.EditProjectForm).Methods("GET")
	webRouter.HandleFunc("/projects/{id}", webHandler.UpdateProject).Methods("PUT")
	webRouter.HandleFunc("/projects/{id}", webHandler.DeleteProject).Methods("DELETE")
	
	// Instance routes
	webRouter.HandleFunc("/instances", webHandler.ListInstances).Methods("GET")
	webRouter.HandleFunc("/instances", webHandler.CreateInstance).Methods("POST")
	webRouter.HandleFunc("/instances/new", webHandler.NewInstanceForm).Methods("GET")
	webRouter.HandleFunc("/instances/{id}/edit", webHandler.EditInstanceForm).Methods("GET")
	webRouter.HandleFunc("/instances/{id}", webHandler.UpdateInstance).Methods("PUT")
	webRouter.HandleFunc("/instances/{id}", webHandler.DeleteInstance).Methods("DELETE")
	
	// Metadata routes
	webRouter.HandleFunc("/metadata", webHandler.ListMetadata).Methods("GET")
	webRouter.HandleFunc("/metadata", webHandler.CreateMetadata).Methods("POST")
	webRouter.HandleFunc("/metadata/new", webHandler.NewMetadataForm).Methods("GET")
	webRouter.HandleFunc("/metadata/edit", webHandler.EditMetadataForm).Methods("GET")
	webRouter.HandleFunc("/metadata/update", webHandler.UpdateMetadata).Methods("PUT")
	webRouter.HandleFunc("/metadata/delete", webHandler.DeleteMetadata).Methods("DELETE")

	// API prefix
	api := router.PathPrefix("/v1").Subrouter()

	// Project routes
	api.HandleFunc("/projects", handler.CreateProject).Methods("POST")
	api.HandleFunc("/projects", handler.ListProjects).Methods("GET")
	api.HandleFunc("/projects/{id}", handler.GetProject).Methods("GET")
	api.HandleFunc("/projects/{id}", handler.UpdateProject).Methods("PATCH")
	api.HandleFunc("/projects/{id}", handler.DeleteProject).Methods("DELETE")

	// Instance routes
	api.HandleFunc("/instances", handler.CreateInstance).Methods("POST")
	api.HandleFunc("/instances", handler.ListInstances).Methods("GET")
	api.HandleFunc("/instances/{id}", handler.GetInstance).Methods("GET")
	api.HandleFunc("/instances/{id}", handler.UpdateInstance).Methods("PATCH")
	api.HandleFunc("/instances/{id}", handler.DeleteInstance).Methods("DELETE")

	// Metadata routes
	api.HandleFunc("/metadata", handler.CreateMetadata).Methods("POST")
	api.HandleFunc("/metadata", handler.ListMetadata).Methods("GET").Queries("prefix", "")
	api.HandleFunc("/metadata", handler.ListMetadata).Methods("GET")
	api.HandleFunc("/metadata/{id}", handler.GetMetadata).Methods("GET")
	api.HandleFunc("/metadata/{id}", handler.UpdateMetadata).Methods("PATCH")
	api.HandleFunc("/metadata/{id}", handler.DeleteMetadata).Methods("DELETE")

	// Add CORS middleware for development
	router.Use(corsMiddleware)

	// Add logging middleware
	router.Use(loggingMiddleware)

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Dirt-No-Chaos, X-Dirt-Latency")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware adds basic request logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add proper structured logging here
		// For now, we'll let the main server handle logging
		next.ServeHTTP(w, r)
	})
}