package blog

import (
	"encoding/json"
	"github.com/kcuzner/goblog/site"
	"github.com/kcuzner/goblog/site/auth"
	"github.com/kcuzner/goblog/site/db"
	"github.com/kcuzner/goblog/site/templates"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	post := new(Post)
	err := db.Current.Find(post, bson.M{"path": path}).One(&post)
	if err != nil || post == nil {
		return false
	}

	site.RenderTemplate(w, r, "blog/post", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		d["Post"] = post
		return d, nil
	})

	return true
}

// Handles creating a new post form from nothing
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
	err := db.Current.Find(feed, bson.M{"_id": bson.ObjectIdHex(r.URL.Query().Get("feed"))}).One(&feed)
	if err != nil || feed == nil {
		//no feed?
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := NewPost("", "", "", "", user.Id)

	doPostEditor(post, []Feed{*feed}, w, r)
}

//Handles editing an existing post
func editPostGet(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFor(r)
	if !user.HasRole(NewPostRole) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	id, ok := mux.Vars(r)["id"]
	if !ok || len(id) != 24 {
		//no id or an invalid id length?
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := new(Post)
	err := db.Current.Find(post, bson.M{"_id": bson.ObjectIdHex(id)}).One(&post)
	if err != nil || post == nil {
		//no post?
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feeds, err := post.Feeds()
	if err != nil {
		//feed getting error?
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	doPostEditor(post, feeds, w, r)
}

// Renders the post editor for the passed post with the passed feed hints
// Note that the feeds array is merged client-side with the feeds already attached to the post
func doPostEditor(post *Post, feeds []Feed, w http.ResponseWriter, r *http.Request) {
	site.RenderTemplate(w, r, "blog/edit-post", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		feedIds := make([]string, len(feeds))
		for i := range feeds {
			feedIds[i] = feeds[i].Id.Hex()
		}

		allFeeds, err := GetAllFeeds()
		if err != nil {
			return nil, err
		}

		//TODO: The post ids returned needs to be fixed

		d["AllFeeds"] = allFeeds
		d["Feeds"] = feedIds
		d["Post"] = post

		return d, nil
	})
}

type editDTO struct {
	Id      string   `json:id`
	Feeds   []string `json:feeds`
	Title   string   `json:title`
	Path    string   `json:path`
	Parser  string   `json:parser`
	Content string   `json:content`
	Tags    string   `json:tags`
}

type editResponse struct {
	Error string `json:error`
	Id    string `json:id`
}

// Handles submission of the edit post form created by doPostEditor
func editPostPost(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFor(r)
	if !user.HasRole(NewPostRole) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	d := json.NewDecoder(r.Body)
	e := json.NewEncoder(w)

	var req editDTO
	if d.Decode(&req) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//update post
	var post Post
	if err := db.Current.Find(post, bson.M{"_id": bson.ObjectIdHex(req.Id)}).One(&post); err != nil {
		if err != mgo.ErrNotFound {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			post.Created = time.Now()
		}
	}

	post.Id = bson.ObjectIdHex(req.Id)
	post.Path = req.Path
	post.Title = req.Title
	post.Content = req.Content
	post.Parser = req.Parser
	post.Modified = time.Now()
	post.Author = user.Id
	post.Tags = strings.Fields(req.Tags)

	//update feeds
	current := make(map[string]bool)
	for i := range req.Feeds {
		current[req.Feeds[i]] = true
	}
	existing, err := post.Feeds()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for i := range existing {
		if v, ok := current[existing[i].Id.Hex()]; !ok || !v {
			current[existing[i].Id.Hex()] = false
		}
	}
	feedIds := make([]bson.ObjectId, 0, len(current))
	for k, _ := range current {
		println(k)
		feedIds = append(feedIds, bson.ObjectIdHex(k))
	}
	feed := new(Feed)
	var feeds Feeds
	if err := db.Current.Find(feed, bson.M{"_id": bson.M{"$in": feedIds}}).All(&feeds); err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		e.Encode(editResponse{"Unable to load feeds", post.Id.Hex()})
		return
	}
	for i := range feeds {
		if current[feeds[i].Id.Hex()] {
			(&feeds[i]).AddPost(post.Id)
		} else {
			(&feeds[i]).RemovePost(post.Id)
		}
		if _, err = db.Current.Upsert(feeds[i]); err != nil {
			println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			e.Encode(editResponse{"Unable to save feeds", post.Id.Hex()})
			return
		}
	}

	if _, err := db.Current.Upsert(post); err != nil {
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		e.Encode(editResponse{"Unable to save post", post.Id.Hex()})
		return
	} else if err := e.Encode(editResponse{"", post.Id.Hex()}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// gets a page of posts with a specific tag
func tagGet(w http.ResponseWriter, r *http.Request) {
	tag, ok := mux.Vars(r)["tag"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	posts, err := GetPostsByTag(tag, 1, 20)
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	site.RenderTemplate(w, r, "blog/feed", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
		d["Page"] = posts
		if err != nil {
			d["Page"] = make([]interface{}, 0)
		}
		d["FeedTitle"] = "Tag: " + tag
		return d, nil
	})
}

// adds tag data to the template variables
func addPostTags(r *http.Request, t *templates.Vars) {
	tags, err := GetTagCount()
	if err == nil {
		(*t)["Tags"] = tags
	}
}

func init() {
	auth.RegisterRole(NewPostRole)
	
	site.HandlePathFunc(feedGet)
	site.HandlePathFunc(postGet)

	templates.Register(addPostTags)

	s := site.GetSite()
	pr := s.Router().PathPrefix("/posts").Subrouter()
	pr.HandleFunc("/new", newPostGet).
		Methods("GET")
	pr.HandleFunc("/edit/{id}", editPostGet).
		Methods("GET")
	pr.HandleFunc("/edit", editPostPost).
		Methods("POST").
		Headers("X-Requested-With", "XMLHttpRequest")
	pr.HandleFunc("/tag/{tag}", tagGet).
		Methods("GET")
}
