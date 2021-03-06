package blog

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kcuzner/goblog/site/db"
	"github.com/russross/blackfriday"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"math"
	"sort"
	"time"
)

type (
	Posts []Post
	Post  struct {
		Id       bson.ObjectId `json:"id" bson:"_id"`
		Path     string        `json:"path" bson:"path"`         //current path to the post
		Tags     []string      `json:"tags" bson:"tags"`         //current tags on the post
		Versions []PostVersion `json:"versions" bson:"versions"` //sequential versions of this post (last is most recent)
		Created  time.Time     `json:"created" bson:"created"`   //time this post was originally created
		Modified time.Time     `json:"modified" bson:"modified"` //last modified time of this post
		Author   bson.ObjectId `json:"author" bson:"_author"`    //author of this post
		revision *PostVersion
	}
	PostVersion struct {
		Created time.Time `json:"created" bson:"created"` //time this version was created
		Path    string    `json:"path" bson:"path"`
		Title   string    `json:"title" bson:"title"`
		Content string    `json:"content" bson:"content"`
		Parser  string    `json:"parser" bson:"parser"`
		Tags    []string  `json:"tags" bson:"tags"`
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
	post.Path = path
	post.Versions = append(post.Versions, PostVersion{
		Created: time.Now(),
		Path:    path,
		Title:   title,
		Content: content,
		Parser:  parser,
	})
	post.Id = bson.NewObjectId()
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
	Size  float64
}

func (t tagData) Style() template.CSS {
	return template.CSS(fmt.Sprintf("font-size: %.2fem;", t.Size))
}

type tagCount []tagData

func (t tagCount) Len() int      { return len(t) }
func (t tagCount) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t tagCount) Less(i, j int) bool {
	first := t[i]
	second := t[j]
	if first.Count < second.Count {
		return true
	} else if first.Count > second.Count {
		return false
	} else {
		//since this is going to be reverse sorted, but we still want
		//alphabetical order, we need to reverse the alphabetical part of
		//the comparison
		return bytes.Compare([]byte(first.Tag), []byte(second.Tag)) > 0
	}
}

// map reduce query for counting post tags
// This, my friends, is why we are using mongo here
var tagCountMapReduce = mgo.MapReduce{
	Map: `function () {
		if (!this.tags) {
			return
		}
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

	sort.Sort(sort.Reverse(count))

	//we will now create the relative sizes
	if len(count) > 0 {
		max := float64(count[0].Count)
		for i := range count {
			c := float64(count[i].Count)
			max = math.Max(c, max)
		}
		for i := range count {
			count[i].Size = math.Max(float64(count[i].Count)/max, 0.5)
		}
	}

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
		Sort("-created").
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
func (p *Post) PreSave() {
	if p.Versions == nil {
		p.Versions = make([]PostVersion, 0)
	}

	if p.revision != nil {
		//we are being revised!
		p.Versions = append(p.Versions, *p.revision)
		p.Tags = p.revision.Tags
		p.Path = p.revision.Path
		p.revision = nil
	}
}

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

// Gets the most recent version of this post
func (p Post) Version() PostVersion {
	if len(p.Versions) == 0 {
		return PostVersion{}
	}

	return p.Versions[len(p.Versions)-1]
}

// Gets the compiled HTML for this post
func (p Post) Compiled() template.HTML {
	if len(p.Versions) == 0 {
		return template.HTML("")
	}

	v := p.Version()

	if v.Parser == "Markdown" {
		return template.HTML(blackfriday.MarkdownCommon([]byte(v.Content)))
	}

	return template.HTML(v.Content)
}

func (p Post) Title() string {
	return p.Version().Title
}

func (p Post) CreatedString() string {
	return p.Created.Format(time.RFC1123)
}

// Sets the current revision of this post
// Returns true if the revision will be added
func (p *Post) SetRevision(v *PostVersion) bool {
	if (v.SameAs(p.Version())) {
		return false
	}
	
	p.revision = v
	return true
}

// Compares this post version to another post version
func (v PostVersion) SameAs(other PostVersion) bool {
	//compare tags first
	myTags := make(map[string]bool)
	for i := range v.Tags {
		myTags[v.Tags[i]] = true
	}
	
	theirTags := make(map[string]bool)
	for i := range other.Tags {
		theirTags[other.Tags[i]] = true
		if _, ok := myTags[other.Tags[i]]; !ok {
			return false //they have a tag we don't
		}
	}
	
	for i := range v.Tags {
		if _, ok := theirTags[v.Tags[i]]; !ok {
			return false //we have a tag they don't
		}
	}
	
	//if we make it this far, compare the other properties
	return v.Path == other.Path &&
		v.Title == other.Title &&
		v.Content == other.Content &&
		v.Parser == other.Parser
}

func (p Posts) Len() int           { return len(p) }
func (p Posts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
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
	postIds := f.visiblePostIds()

	length := len(postIds)
	if number < 1 || number > int(math.Ceil(float64(length)/float64(size))) {
		return nil, errors.New("Page number is out of range")
	}

	index0 := (number - 1) * size
	index1 := int(math.Min(float64(length), float64(index0+size)))

	pageIds := postIds[index0:index1]

	post := new(Post)
	var results Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": pageIds}}).All(&results)
	if err != nil {
		return nil, err
	}

	sResults := make(map[bson.ObjectId]Post)
	for i := range results {
		sResults[results[i].Id] = results[i]
	}

	page := make(Posts, 0, len(pageIds))
	for i := range pageIds {
		p, ok := sResults[pageIds[i]]
		if ok {
			page = append(page, p)
		}
	}

	return page, nil
}

// Returns true if the feed has no visible posts
func (f Feed) Empty() bool {
	return len(f.visiblePostIds()) == 0
}

// Gets a "preview" of the most recent posts for this feed.
// "most recent" is the titles of the last three posts possibly with a ...
// appended on the end if there are more posts
func (f Feed) Preview() []string {
	postIds := f.visiblePostIds()

	previews := make([]string, 0, 4)

	topIndex := int(math.Min(float64(len(postIds)), 3))
	previewIds := postIds[:topIndex] //we get the first 3 postids

	post := new(Post)
	var posts Posts
	err := db.Current.Find(post, bson.M{"_id": bson.M{"$in": previewIds}}).All(&posts)
	if err != nil {
		log.Println(err)
	}

	for i := range posts {
		previews = append(previews, posts[i].Title())
	}

	if len(previewIds) < len(postIds) {
		previews = append(previews, "...")
	}

	return previews
}

// Adds a post or re-activates it
func (f *Feed) AddPost(id bson.ObjectId) {
	for i := range f.FeedPosts {
		if f.FeedPosts[i].Id == id {
			f.FeedPosts[i].Visible = true
			return
		}
	}

	f.FeedPosts = append([]FeedPost{FeedPost{id, true}}, f.FeedPosts...)
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
