package blog

import (
	"net/http"
	//"html/template"
	"github.com/kcuzner/goblog/site"
	"github.com/kcuzner/goblog/site/auth"
	"github.com/kcuzner/goblog/site/db"
	"github.com/kcuzner/goblog/site/templates"
	"labix.org/v2/mgo/bson"
	"strconv"
)

const (
	NewPostRole auth.Role = "PostNew"
)

func feedGet(path string, w http.ResponseWriter, r *http.Request) bool {
	feed := new(Feed)
	err := db.Current.Find(feed, bson.M{"path": path}).One(&feed)
	if err != nil || feed == nil {
		return false
	}

	site.RenderTemplate(w, r, "blog/feed", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		number, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || number < 1 {
			number = 1
		}
		d["Page"], err = feed.PostPage(number, 20)
		if err != nil {
			d["Page"] = make([]interface{}, 0)
		}
		d["FeedId"] = feed.Id.Hex()
		d["FeedTitle"] = feed.Title
		return d, nil
	})
	
	return true
}

func postGet(path string, w http.ResponseWriter, r *http.Request) bool {
	println("post?", path)
	return false
}

func newPostGet(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFor(r)
	if !user.HasRole(NewPostRole) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	feedId := r.URL.Query().Get("feed")
	if feedId == "" || len(feedId) != 24 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feed := new(Feed)
	err := db.Current.Find(feed, bson.M{ "_id": bson.ObjectIdHex(r.URL.Query().Get("feed")) }).One(&feed)
	if err != nil || feed == nil {
		//no feed?
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := NewPost("", "", "", "", user.Id)

	doPostEditor(post, []Feed{ *feed }, w, r)
}

func doPostEditor(post *Post, feeds []Feed, w http.ResponseWriter, r *http.Request) {
	site.RenderTemplate(w, r, "blog/edit-post", func (w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		feedIds := make([]string, len(feeds))
		for i := range feeds {
			feedIds[i] = feeds[i].Id.Hex()
		}

		allFeeds, err := GetAllFeeds()
		if err != nil{
			return nil, err
		}

		//TODO: The post ids returned needs to be fixed

		d["AllFeeds"] = allFeeds
		d["Feeds"] = feedIds
		d["Post"] = post

		return d, nil
	})
}

func newPostPost(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFor(r)
	if !user.HasRole(NewPostRole) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func init() {
	auth.RegisterRole(NewPostRole)
	
	site.HandlePathFunc(feedGet)
	site.HandlePathFunc(postGet)

	s := site.GetSite()
	pr := s.Router().PathPrefix("/posts").Subrouter()
	pr.HandleFunc("/new", newPostGet).
		Methods("GET")
	pr.HandleFunc("/new", newPostPost).
		Methods("POST").
		Headers("X-Requested-With", "XMLHttpRequest")
}
