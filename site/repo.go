package site

import (
    "errors"
    "bytes"
    "crypto/rand"
    "crypto/sha512"
    "labix.org/v2/mgo/bson"
    "labix.org/v2/mgo"
    "code.google.com/p/go.crypto/pbkdf2"
    "bitbucket.org/kcuzner/goblog/site/config"
)

const (
    // Length in bytes of user salts
    SaltLength = 24
    // Number of iterations to use for generating password keys via pbkdf2
    PasswordIterations = 32767
    // Key length in bits to be generated via pbkdf2
    KeyLengthBits = 256
)

type Repository struct {
    session *mgo.Session
    db *mgo.Database
    users *UserRepository
}

func requireSession() *mgo.Session {
    c := config.GetConfiguration()

    session, err := mgo.Dial(c.ConnectionString)
    if err != nil {
        panic(err)
    }

    return session
}

var baseSession = requireSession()

// Creates a new repository
func NewRepository() *Repository {
    session := baseSession.Copy()
    db := session.DB("")

    repo := Repository{session, db, nil}

    return &repo
}

func (r *Repository) GetUserRepository() *UserRepository {
    if r.users == nil {
        r.users = r.newUserRepository()
    }
    return r.users
}

func (r *Repository) Close() {
    r.session.Close()
}

type UserRepository struct {
    // parent repository which created this
    r *Repository
    // user collection
    c *mgo.Collection
}

// Creates a new user repository for a repository
func (r *Repository) newUserRepository() *UserRepository {
    return &UserRepository{r, r.db.C("users")}
}

// Creates a new user
func (u *UserRepository) CreateUser(username, password, displayName string) (*User, error) {
    user := User{}
    user.Id = bson.NewObjectId()
    user.Username = username
    user.DisplayName = displayName
    user.SetPassword(password)

    _, err := u.c.UpsertId(user.Id, user)

    return &user, err
}

// Gets an existing user
func (u *UserRepository) GetUser(username string) (*User, error) {
    var results Users
    err := u.c.Find(map[string]interface{}{
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

// Updates an existing user
func (u *UserRepository) Update(user *User) error {
    return nil
}

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
