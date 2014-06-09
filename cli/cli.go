package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	//"labix.org/v2/mgo"
	"github.com/howeyc/gopass"
	//"bitbucket.org/kcuzner/goblog/site/config"
	"bitbucket.org/kcuzner/goblog/site"
)

func main() {
	//c := config.GetConfiguration()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Enter admin username: ")
	scanner.Scan()
	username := scanner.Text()

	repo := site.NewRepository()
	defer repo.Close()

	user, err := repo.Users().User(username)
	if err != nil {
		panic(err)
	}

	if user == nil {
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
		user, err = repo.Users().Create(username, password, displayName)
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

}
