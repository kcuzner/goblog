package site

import (
    "log"
    "net/http"
    "github.com/gorilla/context"
    "bitbucket.org/kcuzner/goblog/site/templates"
)

type UserKeyType int

const (
    UserKey UserKeyType = iota
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
    //the very last thing we do is redirect to somewhere (set by redirectAddr)
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
            session.Values["username"] = username
            redirectAddr = "/"
            return
        }
    }
}

func userLogoutGet(w http.ResponseWriter, r *http.Request) {
    defer http.Redirect(w, r, "/", http.StatusFound)

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer session.Save(r, w)

    delete(session.Values, "username")
    templates.Flash(session, templates.SuccessFlashKey, "You have been logged out")
}

// to be executed before each request to set some global variables that will be helpful in templates
func userOnBeforeRequest(r *http.Request) {
    repo := NewRepository()
    defer repo.Close()

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        return
    }

    val, ok := session.Values["username"]
    if !ok {
        return
    }

    user, err := repo.Users().User(val.(string))
    if err != nil || user == nil {
        return
    }

    context.Set(r, UserKey, user)
    println("User is set")
}

func init() {
    RegisterHookBefore(BeforeHookImpl{userOnBeforeRequest})

    s := GetSite()

    sr := s.r.PathPrefix("/user").Subrouter()

    sr.HandleFunc("/login", userLoginGet).
        Methods("GET")
    sr.HandleFunc("/login", userLoginPost).
        Methods("POST")
    sr.HandleFunc("/logout", userLogoutGet).
        Methods("GET")
}
