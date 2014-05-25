package main

import (
    "log"
    "net/http"
    _ "bitbucket.org/kcuzner/goblog/site"
)

func main() {
    log.Fatal(http.ListenAndServe(":3000", nil))
}
