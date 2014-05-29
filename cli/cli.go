package main

import (
    "fmt"
    "bufio"
    "os"
    //"labix.org/v2/mgo"
    //"github.com/howeyc/gopass"
    //"bitbucket.org/kcuzner/goblog/site/config"
    //"bitbucket.org/kcuzner/goblog/site"
)

func main() {
    //c := config.GetConfiguration()

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Printf("Enter admin username: ")
    scanner.Scan()
    username := scanner.Text()
    println(username)
}
