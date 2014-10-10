package main

import (
	_ "github.com/kcuzner/goblog/site"
	_ "github.com/kcuzner/goblog/site/auth"
	"github.com/kcuzner/goblog/site/config"
	"log"
	"net/http"
)

func main() {
	c := config.Config

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(c.PublicDir))))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
