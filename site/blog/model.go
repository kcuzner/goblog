package blog

import (
	"github.com/kcuzner/goblog/site/db"
	"labix.org/v2/mgo/bson"
	"time"
	"errors"
	"math"
)

type (
	Posts []Post
	Post  struct {
		Id       bson.ObjectId   `json:"id" bson:"_id"`
		Path     string          `json:"path" bson:"path"`
		Title    string          `json:"title" bson:"title"`
		Content  string          `json:"content" bson:"content"`
		Parser   string          `json:"parser" bson:"parser"`
		Created  time.Time       `json:"created" bson:"created"`
		Modified time.Time       `json:"modified" bson:"modified"`
		Author   bson.ObjectId `json:"author" bson:"_author"`
	}
	Feeds []Feed
	Feed  struct {
		Id      bson.ObjectId   `json:"id" bson:"_id"`
		Path    string          `json:"path" bson:"path"`
		Title   string          `json:"title" bson:"title"`
		PostIds []bson.ObjectId `json:"posts" bson:"_posts"`
	}
)

func NewPost(path, title, content, parser string, author bson.ObjectId) *Post {
	post := new(Post)
	post.Id = bson.NewObjectId()
	post.Path = path
	post.Title = title
	post.Content = content
	post.Parser = parser
	post.Created = time.Now()
	post.Modified = post.Created
	post.Author = author

	return post
}

func GetPost(path string) (*Post, error) {
	post := new(Post)
	err := db.Current.Find(post, bson.M{"path": path}).One(&post)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (p Post) Collection() string  { return "posts" }
func (p Post) Indexes() [][]string { return [][]string{[]string{"path"}} }
func (p Post) Unique() bson.M      { return bson.M{"path": p.Path} }
func (p Post) PreSave()            {}

func NewFeed(path, title string) *Feed {
	feed := new(Feed)
	feed.Id = bson.NewObjectId()
	feed.Path = path
	feed.Title = title

	return feed
}

func GetAllFeeds() (Feeds, error) {
	feed := new(Feed)

	var results Feeds
	err := db.Current.Find(feed, nil).All(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetFeed(path string) (*Feed, error) {
	feed := new(Feed)
	err := db.Current.Find(feed, bson.M{"path": path}).One(&feed)

	if err != nil {
		return nil, err
	}

	return feed, nil
}

func (f Feed) Collection() string  { return "feeds" }
func (f Feed) Indexes() [][]string { return [][]string{[]string{"path"}} }
func (f Feed) Unique() bson.M      { return bson.M{"path": f.Path} }
func (f Feed) PreSave()            {}

func (f Feed) Posts() (Posts, error) {
	post := new(Post)
	var results Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": f.PostIds}}).All(&results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (f Feed) PostPage(number, size int) (Posts, error) {
	length := len(f.PostIds)
	if number < 1 || number > int(math.Ceil(float64(length) / float64(size))) {
		return nil, errors.New("Page number is out of range")
	}

	index := (number - 1) * size

	post := new(Post)
	var results Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": f.PostIds[index:index + size]}}).All(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (f *Feed) AddPost(p *Post) {
	f.PostIds = append(f.PostIds, p.Id)
}

func init() {
	db.Register(&Post{})
	db.Register(&Feed{})
}
