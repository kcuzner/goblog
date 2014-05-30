package site

import (
    "log"
    "net/http"
    "bitbucket.org/kcuzner/goblog/site/templates"
)


// Handles GET /user/login.
// Simply displays a form
func userLoginGet(w http.ResponseWriter, r *http.Request) {
    tmpl, err := templates.Cache.Get("user/login")

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, templates.GetGlobalVars())
}

// Handles POST /user/login.
// Validates the user and possibly sets the session user if everything is valid
func userLoginPost(w http.ResponseWriter, r *http.Request) {
    repo := NewRepository()
    defer repo.Close()

    _, err := repo.GetUserRepository().GetUser("kcuzner")

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }


    http.Redirect(w, r, "/user/login", http.StatusFound)
}

func init() {
    s := GetSite()

    sr := s.r.PathPrefix("/user").Subrouter()

    sr.HandleFunc("/login", userLoginGet).
        Methods("GET")
    sr.HandleFunc("/login", userLoginPost).
        Methods("POST")
}
