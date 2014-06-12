package templates

import (
	"net/http"
)

type Vars map[interface{}]interface{}

var handlers = make([]func(*http.Request, *Vars), 0)

// Registers a function which will modify the global vars in some way, generally
// by adding new members
func Register(handler func(*http.Request, *Vars)) {
	handlers = append(handlers, handler)
}

// Gets the global variables for templates
// This runs all handlers which have been registered to create global variables.
// The handlers could run in arbitrary order.
func GetGlobalVars(r *http.Request) Vars {
	data := make(Vars)
	for i := range handlers {
		handlers[i](r, &data)
	}

	return data
}
