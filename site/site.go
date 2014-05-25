package site

import (
    "net/http"
    "github.com/gorilla/mux"
)

type Site struct {
    initialized bool
    r *mux.Router
}

func (s *Site) initialize() {
    if !s.initialized {
        s.r = mux.NewRouter()
    }
}

var site Site

func GetSite() *Site {
    site.initialize()
    return &site
}

func init() {
    s := GetSite()
    http.Handle("/", s.r)
}
