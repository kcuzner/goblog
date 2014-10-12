package auth

import (
	"github.com/kcuzner/goblog/site"
	"github.com/kcuzner/goblog/site/templates"
	"net/http"
)

const (
	AdministrateUsersRole Role = "Administrate Users"
)

func adminPageGet(w http.ResponseWriter, r *http.Request) {
	println("asdf")
	site.RenderTemplate(w, r, "users/admin", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		return d, nil
	})
}

func adminSearchGet(w http.ResponseWriter, r *http.Request) {
}

func adminUserSavePut(w http.ResponseWriter, r *http.Request) {
}

func init() {
	RegisterRole(AdministrateUsersRole)
	s := site.GetSite()

	sr := s.Router().PathPrefix("/users").Subrouter()
	
	s.Router().Handle("/users", Authorize(adminPageGet).HasRole(AdministrateUsersRole)).
		Methods("GET")

	sr.Handle("/search", Authorize(adminSearchGet).HasRole(AdministrateUsersRole)).
		Methods("GET").
		Headers("X-Requested-With", "XMLHttpRequest",
		"Content-Type", "application/json")
	sr.Handle("/save", Authorize(adminUserSavePut).HasRole(AdministrateUsersRole)).
		Methods("PUT").
		Headers("X-Requested-With", "XMLHttpRequest",
		"Content-Type", "application/json")
}
