package site

import (
    "net/http"
    "github.com/gorilla/context"
    "github.com/gorilla/sessions"
)

type contextKeyType int
const (
    ContextKey contextKeyType = iota
)


// Variables stored as the context for a request used by the view in each page
type ContextVariables struct {
    *sessions.Session
    user *User
}

func GetContextVariables(r *http.Request) (*ContextVariables, error) {
    vars, ok := context.GetOk(r, ContextKey)
    if !ok {
        session, err := store.Get(r, MainSessionName)
        if err != nil {
            return nil, err
        }
        vars = &ContextVariables{session, nil}
        context.Set(r, ContextKey, vars)
    }

    return vars.(*ContextVariables), nil
}

func (c *ContextVariables) User() interface{} {
    return c.user
}

func (c *ContextVariables) SetUser(u *User) {
    c.user = u
}
