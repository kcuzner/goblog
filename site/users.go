package site

import (
    "net/http"
    "bitbucket.org/kcuzner/goblog/site/templates"
)

func userLoginGet(w http.ResponseWriter, r *http.Request) {
    tmpl, err := templates.Cache.Get("user/login")

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, r)
}

func userLoginPost(w http.ResponseWriter, r *http.Request) {

}

func init() {
    s := GetSite()

    sr := s.r.PathPrefix("/user").Subrouter()

    sr.HandleFunc("/login", userLoginGet).
        Methods("GET")
    sr.HandleFunc("/login", userLoginPost).
        Methods("POST")
}
