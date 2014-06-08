package templates

import (
    "net/http"
)

// Source for flashes compatible with gorrila sessions
type FlashSource interface {
    Flashes(...string) []interface{}
}

// Source for a user
type UserSource interface {
    User() interface{}
}

type GlobalVarSource interface {
    FlashSource
    UserSource
}

// Store for flashes compatible with gorrilla sessions
type FlashStore interface {
    AddFlash(value interface{}, vars ...string)
}

// Global template data
type GlobalVars struct {
    SiteTitle string
    Errors, Warnings, Infos, Successes []string
    User interface{}
}

type Vars interface{}

var handlers = make([]func(*http.Request, Vars) Vars, 0)

// Registers a function which will modify the global vars in some way, generally
// by adding new members
func Register(handler func(*http.Request, Vars) Vars) {
    handlers = append(handlers, handler)
}

// Gets the global variables for templates
// This runs all handlers which have been registered to create global variables.
// The handlers could run in arbitrary order.
func GetGlobalVars(r *http.Request) interface{} {
    var data Vars = struct{}{}
    for i := range handlers {
        data = handlers[i](r, data)
    }

    return data
}

