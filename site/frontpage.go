package site

import(
    "sort"
    "net/http"
    "bitbucket.org/kcuzner/goblog/site/templates"
)

// Block for something to appear on the front page
type FrontPageBlock struct {
    Order int
    Heading, Subheading, Body string
    MediaObject string
}

// Sorter for the front page blocks
type FrontPageByOrder []FrontPageBlock
func (f FrontPageByOrder) Len() int { return len(f) }
func (f FrontPageByOrder) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
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

//var defaultOptions = amber.Options{true, false}

func init() {
    s := GetSite()

    s.r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
        blocks := make([]FrontPageBlock, 0)
        for i := range frontPageHandlers {
            blocks = append(blocks, frontPageHandlers[i].GetFrontPage(request)...)
        }
        sort.Sort(FrontPageByOrder(blocks))

        tmpl, err := templates.Cache.Get("frontpage")

        if err != nil {
            writer.WriteHeader(http.StatusInternalServerError)
            return
        }

        tmpl.Execute(writer, request)
    }).Name("index")
}
