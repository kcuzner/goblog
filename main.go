package main

import (
	_ "bitbucket.org/kcuzner/goblog/site"
	_ "bitbucket.org/kcuzner/goblog/site/auth"
	"bitbucket.org/kcuzner/goblog/site/config"
	"log"
	"net/http"
)

func main() {
	c := config.Config

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(c.PublicDir))))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
