package site

import (
    "log"
    "net/http"
    "errors"
    "bytes"
    "crypto/rand"
    "crypto/sha512"
    "labix.org/v2/mgo/bson"
    "labix.org/v2/mgo"
    "code.google.com/p/go.crypto/pbkdf2"
    "bitbucket.org/kcuzner/goblog/site/templates"
)

const (
    // Length in bytes of user salts
    SaltLength = 24
    // Number of iterations to use for generating password keys via pbkdf2
    PasswordIterations = 32767
    // Key length in bits to be generated via pbkdf2
    KeyLengthBits = 256
)

type (
    // User slice type
    Users []User
    // User type used by mgo and also for json data encoding
    User struct {
        Id bson.ObjectId `json:"id" bson:"_id"`
        Username string `json:"username" bson:"username"`
        Password []byte `json:"password" bson:"password"`
        Salt []byte `json:"salt" bson:"salt"`
        DisplayName string `json:"display_name" bson:"display_name"`
    }
)

// Creates a key from a plaintext string using this User's salt 
func (u *User) getKey(plaintext string) []byte {
    return pbkdf2.Key([]byte(plaintext), u.Salt, PasswordIterations, KeyLengthBits / 8, sha512.New)
}

// Sets the password for the user, generating a new salt in the process
func (u *User) SetPassword(plaintext string) error {
    salt := make([]byte, SaltLength)
    n, err := rand.Read(salt)
    if err != nil {
        return err
    }
    if n != SaltLength {
        return errors.New("Unable to generate salt of sufficient length")
    }

    u.Salt = salt
    u.Password = u.getKey(plaintext)
    return nil
}

// Validates the passed plaintext password against this user's stored password
func (u *User) ValidatePassword(plaintext string) bool {
    test := u.getKey(plaintext)

    return bytes.Equal(u.Password, test)
}

func getUserCollection() *mgo.Collection {
    db := GetMgoSession().DB("")
    return db.C("users")
}

func GetUser(username string) (*User, error) {
    var results Users
    err := getUserCollection().Find(map[string]interface{}{
        "username": username,
        }).Limit(1).Iter().All(&results)

    if err != nil {
        return nil, err
    }

    if len(results) > 0 {
        return &results[0], nil
    }

    return nil, nil
}


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
    db := GetMgoSession().DB("")
    users := db.C("users")

    var results Users
    err := users.Find(map[string]interface{}{
        "username": "kcuzner",
        }).Limit(1).Iter().All(&results)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println(err)
        return
    }

    log.Println("Found ", len(results), " users")


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
