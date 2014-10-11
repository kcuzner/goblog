package main

import (
	_ "github.com/kcuzner/goblog/site"
	_ "github.com/kcuzner/goblog/site/auth"
	"github.com/kcuzner/goblog/site/config"
	"log"
	"net/http"
	"os"
)

func main() {
	c := config.Config

	host := os.Getenv("IP")
	port := os.Getenv("PORT")

	if host == "" {
		host = "0.0.0.0"
	}

	if port == "" {
		port = "3000"
	}

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(c.PublicDir))))

	log.Println("Starting server on " + host + ":" + port)
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}
