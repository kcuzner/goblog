package auth

import (
	"bytes"
	"encoding/json"
	"github.com/kcuzner/goblog/site"
	"github.com/kcuzner/goblog/site/db"
	"github.com/kcuzner/goblog/site/templates"
	"html/template"
	"net/http"
	"log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	AdministrateUsersRole Role = "Administrate Users"
)

func adminPageGet(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	e := json.NewEncoder(buf)
	
	if err := e.Encode(AllRoles()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	
	site.RenderTemplate(w, r, "users/admin", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		d["ModuleStyle"] = template.CSS("display: none;")
		d["AllRoles"] = template.JS(buf.String())
		return d, nil
	})
}

func adminSearchGet(w http.ResponseWriter, r *http.Request) {
	phrase := r.FormValue("phrase")
	
	e := json.NewEncoder(w)
	
	var results Users
	user := new(User)
	if err := db.Current.Find(user, bson.M{"username": bson.RegEx{Pattern:".*"+phrase+".*"}}).All(&results); err != nil {
		//error searching
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := e.Encode(results); err != nil {
		//error encoding
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

type userSaveDTO struct {
	Id string `json:id`
	Username string `json:id`
	DisplayName string `json:displayName`
	Roles []Role `json:roles`
}

func adminUserSavePut(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	e := json.NewEncoder(w)
	
	var req userSaveDTO;
	d.Decode(&req)
	
	var user User
	err := db.Current.Find(&user, bson.M{"_id": bson.ObjectIdHex(req.Id)}).One(&user)
	if err != nil && err != mgo.ErrNotFound {
		//error loading
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	} else if err == mgo.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	user.Username = req.Username
	user.DisplayName = req.DisplayName
	user.Roles = req.Roles
	
	_, err = db.Current.Upsert(&user)
	if err != nil {
		//error saving
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	err = e.Encode(user)
	if err != nil {
		//error encoding
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func init() {
	RegisterRole(AdministrateUsersRole)
	s := site.GetSite()

	sr := s.Router().PathPrefix("/users").Subrouter()
	
	s.Router().Handle("/users", Authorize(adminPageGet).HasRole(AdministrateUsersRole)).
		Methods("GET")

	sr.Handle("/search", Authorize(adminSearchGet).HasRole(AdministrateUsersRole)).
		Methods("GET").
		Headers("X-Requested-With", "XMLHttpRequest")
	sr.Handle("/save", Authorize(adminUserSavePut).HasRole(AdministrateUsersRole)).
		Methods("PUT").
		Headers("X-Requested-With", "XMLHttpRequest",
		"Content-Type", "application/json")
}
