package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/kcuzner/goblog/site/auth"
	"github.com/kcuzner/goblog/site/db"
	"labix.org/v2/mgo"
	"os"
)

func main() {
	//c := config.GetConfiguration()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Enter admin username: ")
	scanner.Scan()
	username := scanner.Text()

	user, err := auth.GetUser(username)
	if err != nil && err != mgo.ErrNotFound {
		panic(err)
	}

	if user == nil || err == mgo.ErrNotFound {
		//create a new user
		var password string
		for {
			fmt.Printf("Enter admin password: ")
			password = string(gopass.GetPasswd())
			fmt.Printf("Enter password again: ")
			if password == string(gopass.GetPasswd()) {
				break
			}
			println("Passwords don't match")
		}

		fmt.Printf("Enter admin display name: ")
		scanner.Scan()
		displayName := scanner.Text()

		user, err = auth.NewUser(username, password, displayName)
		if err != nil {
			panic(err)
		}
	} else {
		//confirm that the user is the same
		fmt.Printf("Enter admin password: ")
		if !user.ValidatePassword(string(gopass.GetPasswd())) {
			panic(errors.New("Incorrect password"))
		}
	}
	
	user.AddRole(auth.AdministrateUsersRole)
	db.Current.Upsert(user)

}
