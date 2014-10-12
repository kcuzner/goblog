package auth

import (
	"github.com/gorilla/context"
	"github.com/kcuzner/goblog/site"
	"github.com/kcuzner/goblog/site/db"
	"github.com/kcuzner/goblog/site/templates"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"net/url"
)

type UserKeyType int

const (
	UserKey UserKeyType = iota
)

// Handles GET /user/login.
// Simply displays a form
func userLoginGet(w http.ResponseWriter, r *http.Request) {
	site.RenderTemplate(w, r, "user/login", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		d["Next"] = r.URL.Query().Get("next")
		return d, nil
	})
}

// Handles POST /user/login.
// Validates the user and possibly sets the session user if everything is valid
func userLoginPost(w http.ResponseWriter, r *http.Request) {
	next := r.FormValue("next")

	//the very last thing we do is redirect to somewhere (set by redirectAddr)
	redirectAddr := "/user/login"
	if next != "" {
		redirectAddr += "?next=" + url.QueryEscape(next)
	}
	defer func(addr *string) { http.Redirect(w, r, *addr, http.StatusFound) }(&redirectAddr)

	session, err := site.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer session.Save(r, w)

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		site.Flash(session, site.ErrorFlashKey, "Username and password are required")

	} else {
		user, err := GetUser(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if user == nil || !user.ValidatePassword(password) {
			site.Flash(session, site.ErrorFlashKey, "Username or password incorrect")
		} else {
			site.Flash(session, site.SuccessFlashKey, "You have been logged in")
			session.Values["username"] = username
			if next != "" {
				redirectAddr = next
			} else {
				redirectAddr = "/"
			}
			return
		}
	}
}

func userLogoutGet(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusFound)

	session, err := site.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer session.Save(r, w)

	delete(session.Values, "username")
	site.Flash(session, site.SuccessFlashKey, "You have been logged out")
}

func userProfileGet(w http.ResponseWriter, r *http.Request) {
	site.RenderTemplate(w, r, "user/profile", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		return d, nil
	})
}

func userProfilePost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/user/profile", http.StatusFound)

	session, err := site.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer session.Save(r, w)

	user := UserFor(r)

	displayName := r.FormValue("displayName")

	if displayName == "" {
		site.Flash(session, site.ErrorFlashKey, "Display Name is required")
		return
	}

	user.DisplayName = displayName
	db.Current.Upsert(user)
	site.Flash(session, site.SuccessFlashKey, "Profile has been saved")
}

func userPasswordGet(w http.ResponseWriter, r *http.Request) {
	site.RenderTemplate(w, r, "user/password", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		return d, nil
	})
}

func userPasswordPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/user/password", http.StatusFound)

	session, err := site.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer session.Save(r, w)

	user := UserFor(r)

	oldPassword := r.FormValue("oldPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	if !user.ValidatePassword(oldPassword) {
		site.Flash(session, site.ErrorFlashKey, "Incorrect password")
		return
	}

	if newPassword != confirmPassword {
		site.Flash(session, site.ErrorFlashKey, "Passwords don't match")
		return
	}

	err = user.SetPassword(newPassword)
	if err != nil {
		//TODO: Potentially dangerious here...
		site.Flash(session, site.ErrorFlashKey, err.Error())
		return
	}

	db.Current.Upsert(user)
	site.Flash(session, site.SuccessFlashKey, "Password has been changed")
}

// to be executed before each request to set some global variables that will be helpful in templates
func userOnBeforeRequest(r *http.Request) {
	session, err := site.Session(r)
	if err != nil {
		return
	}

	val, ok := session.Values["username"]
	if !ok {
		return
	}

	user := new(User)
	err = db.Current.Find(user, bson.M{"username": val.(string)}).One(&user)
	if err != nil || user == nil {
		return
	}
	context.Set(r, UserKey, user)
}

func addUser(r *http.Request, d *templates.Vars) {
	user := UserFor(r)
	(*d)["User"] = user
	if user != nil {
		for i := range user.Roles {
			(*d)["Role"+user.Roles[i]] = true
		}
	}
}

func init() {
	templates.Register(addUser)
	site.RegisterHookBefore(site.BeforeHookImpl{userOnBeforeRequest})

	s := site.GetSite()

	sr := s.Router().PathPrefix("/user").Subrouter()

	sr.HandleFunc("/login", userLoginGet).
		Methods("GET")
	sr.HandleFunc("/login", userLoginPost).
		Methods("POST")
	sr.HandleFunc("/logout", userLogoutGet).
		Methods("GET")
	sr.Handle("/profile", Authorize(userProfileGet)).
		Methods("GET")
	sr.Handle("/profile", Authorize(userProfilePost)).
		Methods("POST")
	sr.Handle("/password", Authorize(userPasswordGet)).
		Methods("GET")
	sr.Handle("/password", Authorize(userPasswordPost)).
		Methods("POST")
}
