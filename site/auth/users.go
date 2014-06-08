package auth

import (
    "log"
    "net/http"
    "github.com/gorilla/context"
    "bitbucket.org/kcuzner/goblog/site"
    "bitbucket.org/kcuzner/goblog/site/templates"
)

type UserKeyType int

const (
    UserKey UserKeyType = iota
)

// Handles GET /user/login.
// Simply displays a form
func userLoginGet(w http.ResponseWriter, r *http.Request) {
    site.RenderTemplate(w, r, "user/login", func(w http.ResponseWriter, r *http.Request, d templates.GlobalVars) (interface{}, error) {
        return d, nil})
}

// Handles POST /user/login.
// Validates the user and possibly sets the session user if everything is valid
func userLoginPost(w http.ResponseWriter, r *http.Request) {
    //the very last thing we do is redirect to somewhere (set by redirectAddr)
    redirectAddr := "/user/login"
    defer func(addr *string) { http.Redirect(w, r, *addr, http.StatusFound) }(&redirectAddr)

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

func userProfileGet(w http.ResponseWriter, r *http.Request) {
    site.RenderTemplate(w, r, "user/profile", func(w http.ResponseWriter, r *http.Request, d templates.GlobalVars) (interface{}, error) {
        return d, nil})
}

func userProfilePost(w http.ResponseWriter, r *http.Request) {
    defer http.Redirect(w, r, "/user/profile", http.StatusFound)

    repo := NewRepository()
    defer repo.Close()

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer session.Save(r, w)

    vars, err := GetContextVariables(r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }

    user := vars.User().(*User)

    displayName := r.FormValue("displayName")

    if displayName == "" {
        templates.Flash(session, templates.ErrorFlashKey, "Display Name is required")
        return
    }

    user.DisplayName = displayName
    repo.Users().Update(user)
    templates.Flash(session, templates.SuccessFlashKey, "Profile has been saved")
}

func userPasswordGet(w http.ResponseWriter, r *http.Request) {
    site.RenderTemplate(w, r, "user/password", func(w http.ResponseWriter, r *http.Request, d templates.GlobalVars) (interface{}, error) {
        return d, nil})
}

func userPasswordPost(w http.ResponseWriter, r *http.Request) {
    defer http.Redirect(w, r, "/user/password", http.StatusFound)

    repo := NewRepository()
    defer repo.Close()

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }
    defer session.Save(r, w)

    vars, err := GetContextVariables(r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }

    user := vars.User().(*User)

    oldPassword := r.FormValue("oldPassword")
    newPassword := r.FormValue("newPassword")
    confirmPassword := r.FormValue("confirmPassword")

    if !user.ValidatePassword(oldPassword) {
        templates.Flash(session, templates.ErrorFlashKey, "Incorrect password")
        return
    }

    if newPassword != confirmPassword {
        templates.Flash(session, templates.ErrorFlashKey, "Passwords don't match")
        return
    }

    err = user.SetPassword(newPassword)
    if err != nil {
        //TODO: Potentially dangerious here...
        templates.Flash(session, templates.ErrorFlashKey, err.Error())
        return
    }

    repo.Users().Update(user)
    templates.Flash(session, templates.SuccessFlashKey, "Password has been changed")
}

func UserFor(r *http.Request) *User {
    user, ok := context.GetOk(r, UserKey)
    return user.(*User)
}

// to be executed before each request to set some global variables that will be helpful in templates
func userOnBeforeRequest(r *http.Request) {
    repo := NewRepository()
    defer repo.Close()

    session, err := store.Get(r, MainSessionName)
    if err != nil {
        return
    }

    vars, err := GetContextVariables(r)
    if err != nil {
        return
    }

    val, ok := session.Values["username"]
    if !ok {
        return
    }

    user := new(User)
    err := db.Current.Find(user, bson.M{"username": val.(string)}).One(&user)
    if err != nil || user == nil {
        return
    }
    context.Set(r, UserKey, user)
}

func init() {
    db.Register(User{})
    site.RegisterHookBefore(BeforeHookImpl{userOnBeforeRequest})

    s := site.GetSite()

    sr := s.r.PathPrefix("/user").Subrouter()

    sr.HandleFunc("/login", userLoginGet).
        Methods("GET")
    sr.HandleFunc("/login", userLoginPost).
        Methods("POST")
    sr.HandleFunc("/logout", userLogoutGet).
        Methods("GET")
    sr.HandleFunc("/profile", userProfileGet).
        Methods("GET")
    sr.HandleFunc("/profile", userProfilePost).
        Methods("POST")
    sr.HandleFunc("/password", userPasswordGet).
        Methods("GET")
    sr.HandleFunc("/password", userPasswordPost).
        Methods("POST")
}
