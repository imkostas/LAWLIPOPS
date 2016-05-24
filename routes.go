package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Route struct holds information needed to route traffic
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes type is a helper type for a slice of Route structs
type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		RootHandler,
	},
	Route{
		"Case",
		"GET",
		"/cases/{id}",
		CaseHandler,
	},
	Route{
		"Upload",
		"GET",
		"/upload",
		UploadHandler,
	},
	Route{
		"Upload",
		"POST",
		"/upload",
		UploadHandler,
	},
}

// NewRouter function creates a new router and returns a pointer to it
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}
