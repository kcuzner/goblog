package blog

import (
	"github.com/kcuzner/goblog/site/db"
	"errors"
	"github.com/russross/blackfriday"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math"
	"sort"
	"time"
)

type (
	Posts []Post
	Post  struct {
		Id       bson.ObjectId `json:"id" bson:"_id"`
		Path     string        `json:"path" bson:"path"`
		Title    string        `json:"title" bson:"title"`
		Content  string        `json:"content" bson:"content"`
		Parser   string        `json:"parser" bson:"parser"`
		Tags     []string      `json:"tags" bson:"tags"`
		Created  time.Time     `json:"created" bson:"created"`
		Modified time.Time     `json:"modified" bson:"modified"`
		Author   bson.ObjectId `json:"author" bson:"_author"`
	}
	FeedPost struct {
		Id      bson.ObjectId `json:"id" bson:"_id"`
		Visible bool          `json:"visible" bson:"visible"`
	}
	Feeds []Feed
	Feed  struct {
		Id        bson.ObjectId `json:"id" bson:"_id"`
		Path      string        `json:"path" bson:"path"`
		Title     string        `json:"title" bson:"title"`
		FeedPosts []FeedPost    `json:"posts" bson:"posts"`
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

type tagData struct {
	Tag   string `bson:"_id"`
	Count int    `bson:"value"`
}
type tagCount []tagData

func (t tagCount) Len() int           { return len(t) }
func (t tagCount) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tagCount) Less(i, j int) bool { return t[i].Count < t[j].Count }

// map reduce query for counting post tags
// This, my friends, is why we are using mongo here
var tagCountMapReduce = mgo.MapReduce{
	Map: `function () {
		if (!this.tags) { return; }
		for(index in this.tags) {
			emit(this.tags[index], 1);
		}
	}`,
	Reduce: `function (previous, current) {
		var count = 0;

		for (index in current) {
			count += current[index];
		}

		return count;
	}`,
}

// Gets the tag count for all posts
func GetTagCount() (tagCount, error) {
	post := new(Post)
	var count tagCount
	_, err := db.Current.Find(post, nil).MapReduce(&tagCountMapReduce, &count)
	if err != nil {
		return nil, err
	}

	sort.Sort(count)

	return count, nil
}

// Gets all posts by the tag, sorted in order by creatino date
// tag: Tag to find posts for
// page: Page number to grab posts for (NOTE: This uses skip/take and thus is O(N) for the page number)
func GetPostsByTag(tag string, page, size int) (Posts, error) {
	if page < 1 || size < 1 {
		return nil, errors.New("Page and size must be greater than 0")
	}

	post := new(Post)
	results := make(Posts, 0, 20)
	iter := db.Current.
		Find(post, bson.M{"tags": tag}).
		Sort("created").
		Skip((page - 1) * size).
		Iter()
	for i := 0; i < size; i++ {
		ok := iter.Next(&post)
		if !ok {
			if err := iter.Err(); err != nil {
				return nil, err
			} else {
				break
			}
		} else {
			results = append(results, *post)
		}
	}

	return results, nil
}

func (p Post) Collection() string  { return "posts" }
func (p Post) Indexes() [][]string { return [][]string{[]string{"path"}} }
func (p Post) Unique() bson.M      { return bson.M{"_id": p.Id} }
func (p Post) PreSave()            {}

// Gets all feeds that have this post attached
func (p Post) Feeds() (Feeds, error) {
	feed := new(Feed)

	var results Feeds
	err := db.Current.Find(feed, bson.M{"posts": bson.M{"_id": p.Id, "visible": true}}).All(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Gets the compiled HTML for this post
func (p Post) Compiled() template.HTML {
	if p.Parser == "Markdown" {
		return template.HTML(blackfriday.MarkdownCommon([]byte(p.Content)))
	}

	return template.HTML(p.Content)
}

func (p Post) CreatedString() string {
	return p.Created.Format(time.RFC1123)
}

func (p Posts) Len() int { return len(p) }
func (p Posts) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Posts) Less(i, j int) bool { return p[i].Created.Before(p[j].Created) }

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

// Gets post ids of posts that are currently visible in the feed
func (f Feed) visiblePostIds() []bson.ObjectId {
	posts := make([]bson.ObjectId, 0, len(f.FeedPosts))
	for i := range f.FeedPosts {
		if f.FeedPosts[i].Visible {
			posts = append(posts, f.FeedPosts[i].Id)
		}
	}

	return posts
}

// Gets all visible posts attached to this feed
func (f Feed) Posts() (Posts, error) {
	post := new(Post)
	var results Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": f.visiblePostIds()}}).All(&results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Gets a "page" of posts using the feed's visible posts
func (f Feed) PostPage(number, size int) (Posts, error) {
	posts := f.visiblePostIds()

	length := len(posts)
	if number < 1 || number > int(math.Ceil(float64(length)/float64(size))) {
		return nil, errors.New("Page number is out of range")
	}

	index0 := (number - 1) * size
	index1 := int(math.Min(float64(length), float64(index0+size)))

	post := new(Post)
	var results Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": posts[index0:index1]}}).All(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Adds a post or re-activates it
func (f *Feed) AddPost(id bson.ObjectId) {
	for i := range f.FeedPosts {
		if f.FeedPosts[i].Id == id {
			f.FeedPosts[i].Visible = true
			return
		}
	}

	f.FeedPosts = append(f.FeedPosts, FeedPost{id, true})
}

// De-activates a post
func (f *Feed) RemovePost(id bson.ObjectId) {
	for i := range f.FeedPosts {
		if f.FeedPosts[i].Id == id {
			f.FeedPosts[i].Visible = false
			return
		}
	}
}

func init() {
	db.Register(&Post{})
	db.Register(&Feed{})
}
