package auth

import (
	"github.com/gorilla/context"
	"github.com/kcuzner/goblog/site"
	"log"
	"net/http"
	"net/url"
)

// Gets the current user for a request
func UserFor(r *http.Request) *User {
	user, ok := context.GetOk(r, UserKey)
	if !ok {
		return nil
	}
	return user.(*User)
}

// Authorizer
type authorizer struct {
	fn        func(w http.ResponseWriter, r *http.Request)
	roles     []Role
	unauthUrl string
}

// Creates a new authorizer for the passed handler function
func Authorize(fn func(w http.ResponseWriter, r *http.Request)) authorizer {
	return authorizer{fn, make([]Role, 0), "/"}
}

// Creates a clone of the pointer portions of this authorizer
func (a authorizer) clone() authorizer {
	roles := a.roles
	a.roles = make([]Role, len(roles))
	copy(a.roles, roles)

	return a
}

// Adds a role to the list of roles that can access the handler.
// Roles are or'd together
func (a authorizer) HasRole(role Role) authorizer {
	clone := a.clone()

	clone.roles = append(clone.roles, role)

	return clone
}

// Changes the forbidden redirect URL to something other than "/"
func (a authorizer) RedirectTo(url string) authorizer {
	clone := a.clone()

	clone.unauthUrl = url

	return clone
}

// Determines the authorization status of a request
func (a authorizer) isAuthorized(r *http.Request) bool {
	user := UserFor(r)

	if user == nil {
		return false
	}

	for i := range a.roles {
		if user.HasRole(a.roles[i]) {
			return true
		}
	}

	return len(a.roles) == 0
}

// Executes the authorization for this authorizer
// Implementation of http.Handler
func (a authorizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorized := a.isAuthorized(r)
	user := UserFor(r)

	if !authorized && r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		//AJAX unauthorized request
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	} else if !authorized {
		session, err := site.Session(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		var flash string
		var redirect string
		//Non-ajax unauthorized request
		if user == nil {
			//redirect to login
			flash = "You must log in to access this page"
			redirect = "/user/login?next=" + url.QueryEscape(r.URL.String())
		} else {
			//redirect to home or error page
			flash = "You do not have sufficient privileges to access this page"
			redirect = a.unauthUrl
		}

		site.Flash(session, site.ErrorFlashKey, flash)
		session.Save(r, w)
		http.Redirect(w, r, redirect, http.StatusFound)
	} else {
		//authorized request
		a.fn(w, r)
	}
}
