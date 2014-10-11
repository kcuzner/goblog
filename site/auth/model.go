package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"github.com/kcuzner/goblog/site/db"
	"labix.org/v2/mgo/bson"
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
	Role  string
	Users []User
	// User type used by mgo and also for json data encoding
	User struct {
		Id          bson.ObjectId `json:"id" bson:"_id"`
		Username    string        `json:"username" bson:"username"`
		Password    []byte        `json:"password" bson:"password"`
		Salt        []byte        `json:"salt" bson:"salt"`
		DisplayName string        `json:"display_name" bson:"display_name"`
		Roles       []Role        `json:"roles" bson:"roles"`
	}
)

// Creates a new user
// TODO; Change this so that it doesn't actually save the user
func NewUser(username, password, displayName string) (*User, error) {
	user := new(User)
	user.Id = bson.NewObjectId()
	user.Username = username
	user.DisplayName = displayName
	user.SetPassword(password)
	user.Roles = make([]Role, 0)

	if db.Current.Exists(user) {
		return nil, errors.New("User already exists")
	}

	_, err := db.Current.Upsert(user)

	return user, err
}

func GetUser(username string) (*User, error) {
	user := new(User)
	err := db.Current.Find(user, bson.M{"username": username}).One(&user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) Collection() string  { return "users" }
func (u *User) Indexes() [][]string { return [][]string{[]string{"username"}} }
func (u *User) Unique() bson.M      { return bson.M{"username": u.Username} }
func (u *User) PreSave()            {}

// Creates a key from a plaintext string using this User's salt
func (u *User) getKey(plaintext string) []byte {
	return pbkdf2.Key([]byte(plaintext), u.Salt, PasswordIterations, KeyLengthBits/8, sha512.New)
}

// Sets the password for the user, generating a new salt in the process
func (u *User) SetPassword(plaintext string) error {
	if plaintext == "" || len(plaintext) < 6 {
		return errors.New("Password not sufficiently long")
	}

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

func (u *User) HasRole(role Role) bool {
	for i := range u.Roles {
		if u.Roles[i] == role {
			return true
		}
	}

	return false
}

func (u *User) AddRole(role Role) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

func (u *User) RemoveRole(role Role) {
	var idx int
	for i := range u.Roles {
		if u.Roles[i] == role {
			idx = i
		}
	}

	u.Roles = append(u.Roles[:idx], u.Roles[idx+1:]...)
}

func init() {
	db.Register(&User{})
}
