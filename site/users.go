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
        log.Println(err)
        return
    }

    //the very last thing we do is execute the template
    var data interface{}
    defer func(d interface{}) { tmpl.Execute(w, d) }(&data)

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer session.Save(r, w)

    data = templates.GetGlobalVars(session)
}

// Handles POST /user/login.
// Validates the user and possibly sets the session user if everything is valid
func userLoginPost(w http.ResponseWriter, r *http.Request) {
    redirectAddr := "/user/login"
    defer func(addr *string) { http.Redirect(w, r, *addr, http.StatusFound) }(&redirectAddr)

    repo := NewRepository()
    defer repo.Close()

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer session.Save(r, w)

    username := r.FormValue("username")
    password := r.FormValue("password")

    if username == "" || password == "" {
        templates.Flash(session, templates.ErrorFlashKey, "Username and password are required")

    } else {
        user, err := repo.Users().User(username)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            log.Println(err)
            return
        }

        if user == nil || !user.ValidatePassword(password) {
            templates.Flash(session, templates.ErrorFlashKey, "Username or password incorrect")
        } else {
            templates.Flash(session, templates.SuccessFlashKey, "You have been logged in")
            redirectAddr = "/"
            return
        }
    }
}

func init() {
    s := GetSite()

    sr := s.r.PathPrefix("/user").Subrouter()

    sr.HandleFunc("/login", userLoginGet).
        Methods("GET")
    sr.HandleFunc("/login", userLoginPost).
        Methods("POST")
}
