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
	// Route{
	// 	"Index",
	// 	"POST",
	// 	"/",
	// 	RootHandler,
	// },
	Route{
		"Case",
		"GET",
		"/cases/{id}",
		CaseHandler,
	},
	Route{
		"Case",
		"POST",
		"/cases/{id}",
		CaseHandler,
	},
	Route{
		"Challenge",
		"GET",
		"/challenges/{id}",
		ChallengeHandler,
	},
	Route{
		"Challenge",
		"POST",
		"/challenges/{id}",
		ChallengeHandler,
	},
	Route{
		"Login",
		"GET",
		"/login",
		LoginHandler,
	},
	Route{
		"Login",
		"POST",
		"/login",
		LoginHandler,
	},
	Route{
		"Logout",
		"GET",
		"/logout",
		LogoutHandler,
	},
	Route{
		"My Account",
		"GET",
		"/account",
		AccountHandler,
	},
	Route{
		"My Account",
		"POST",
		"/account",
		AccountHandler,
	},
	Route{
		"Register",
		"GET",
		"/register",
		RegisterHandler,
	},
	Route{
		"Register",
		"POST",
		"/register",
		RegisterHandler,
	},
	Route{
		"Search",
		"GET",
		"/search",
		SearchHandler,
	},
	Route{
		"FacebookAuthCallback",
		"GET",
		"/facebookCallback",
		HandleFacebookCallback,
	},
	Route{
		"FacebookAuth",
		"GET",
		"/facebookAuth",
		HandleFacebookLogin,
	},
	Route{
		"Dashboard",
		"GET",
		"/dashboard",
		DashboardHandler,
	},
	Route{
		"Dashboard",
		"POST",
		"/dashboard",
		DashboardHandler,
	},
	Route{
		"Dashboard",
		"POST",
		"/dashboard/challenges",
		DashboardChallengesHandler,
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
