package blog

import (
	"net/http"
	"github.com/kcuzner/goblog/site"
	//"github.com/kcuzner/goblog/site/db"
)

func feedGet(path string, w http.ResponseWriter, r *http.Request) bool {
	println("feed?", path)
	return false
}

func postGet(path string, w http.ResponseWriter, r *http.Request) bool {
	println("post?", path)
	return false
}

func init() {
	site.HandlePathFunc(feedGet)
	site.HandlePathFunc(postGet)
}
