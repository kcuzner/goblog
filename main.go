package main

import (
    "log"
    "net/http"
    _ "bitbucket.org/kcuzner/goblog/site/auth"
    _ "bitbucket.org/kcuzner/goblog/site"
    "bitbucket.org/kcuzner/goblog/site/config"
)

func main() {
    c := config.Config

    http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(c.PublicDir))))

    log.Fatal(http.ListenAndServe(":3000", nil))
}
