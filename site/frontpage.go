package site

import (
	"github.com/kcuzner/goblog/site/templates"
	"net/http"
	"sort"
)

// Block for something to appear on the front page
type FrontPageBlock struct {
	Order                     int
	Heading, Subheading, Body string
	MediaObject               string
}

// Sorter for the front page blocks
type FrontPageByOrder []FrontPageBlock

func (f FrontPageByOrder) Len() int           { return len(f) }
func (f FrontPageByOrder) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FrontPageByOrder) Less(i, j int) bool { return f[i].Order < f[j].Order }

type FrontPageHandler interface {
	GetFrontPage(request *http.Request) []FrontPageBlock
}

var frontPageHandlers []FrontPageHandler

func getHandlers() []FrontPageHandler {
	if frontPageHandlers == nil {
		frontPageHandlers = make([]FrontPageHandler, 0)
	}

	return frontPageHandlers
}

func AppendHandler(handler FrontPageHandler) {
	frontPageHandlers = append(getHandlers(), handler)
}

func init() {
	s := GetSite()

	s.r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		blocks := make([]FrontPageBlock, 0)
		for i := range frontPageHandlers {
			blocks = append(blocks, frontPageHandlers[i].GetFrontPage(r)...)
		}
		sort.Sort(FrontPageByOrder(blocks))

		RenderTemplate(w, r, "frontpage", func(w http.ResponseWriter, r *http.Request, d templates.Vars) (templates.Vars, error) {
			return d, nil
		})
	}).Name("index")
}
