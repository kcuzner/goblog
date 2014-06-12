package site

import (
	"bitbucket.org/kcuzner/goblog/site/config"
	"bitbucket.org/kcuzner/goblog/site/templates"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

type Site struct {
	r *mux.Router
}

func (s *Site) Router() *mux.Router {
	return s.r
}

var site = Site{mux.NewRouter()}

func GetSite() *Site {
	return &site
}

const mainSessionName = "session"

var store = sessions.NewCookieStore([]byte("Jh!$xPnz6=YeR+N"))

func Session(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, mainSessionName)
}

// Hook to be executed before each request
type BeforeHook interface {
	Before(*http.Request)
}

// Struct for use creating a closure for a BeforeHook
type BeforeHookImpl struct {
	BeforeFunc func(*http.Request)
}

func (h BeforeHookImpl) Before(r *http.Request) { h.BeforeFunc(r) }

// Hook to be executed after each request
type AfterHook interface {
	After(*http.Request)
}

// Struct for use creating a closure for a AfterHook
type AfterHookImpl struct {
	AfterFunc func(*http.Request)
}

func (h AfterHookImpl) After(r *http.Request) { h.AfterFunc(r) }

// Combination type of a before and after hook for ease of use with closures
type BeforeAfterHookImpl struct {
	BeforeHookImpl
	AfterHookImpl
}

var beforeHooks = make([]BeforeHook, 0)
var afterHooks = make([]AfterHook, 0)

// Registers the passed hook to be ran before the request is processed
func RegisterHookBefore(hook BeforeHook) {
	beforeHooks = append(beforeHooks, hook)
}

// Registers the passed hook to be ran after the request is processed
func RegisterHookAfter(hook AfterHook) {
	afterHooks = append(afterHooks, hook)
}

// Processes the hooks for the passed writer/request combination
func processHooks(w http.ResponseWriter, request *http.Request) {
	for i := range beforeHooks {
		beforeHooks[i].Before(request)
	}

	s := GetSite()
	s.r.ServeHTTP(w, request)

	for i := range afterHooks {
		afterHooks[i].After(request)
	}
}

// Handles rendering of a specific template and session saving
// This takes the response writer, request, template name, and a function to
// execute when the template and session are successfully loaded. The function
// should return something to be fed into the template's Execute method.
func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, f func(http.ResponseWriter, *http.Request, templates.Vars) (templates.Vars, error)) {
	tmpl, err := templates.Cache.Get(name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("renderTemplate:", err)
		return
	}

	//the very last thing we do is execute the template
	//This works because if there is an error, we already write the response
	//before this deferred execution function is executed
	var data interface{}
	defer func(d interface{}) {
		err := tmpl.Execute(w, d)
		if err != nil {
			log.Println(err)
		}
	}(&data)

	session, err := Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("renderTemplate:", err)
		return
	}
	defer session.Save(r, w)

	data, err = f(w, r, templates.GetGlobalVars(r))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("renderTemplate:", err)
		return
	}
}

type FlashKey string

const (
	ErrorFlashKey   FlashKey = "error"
	WarningFlashKey          = "warning"
	InfoFlashKey             = "info"
	SuccessFlashKey          = "success"
)

func Flash(s *sessions.Session, key FlashKey, message string) {
	s.AddFlash(message, string(key))
}

func getFlashes(r *http.Request, key FlashKey) []string {
	s, err := Session(r)
	if err != nil {
		return nil
	}

	inFlashes := s.Flashes(string(key))
	outFlashes := make([]string, 0)
	for i := range inFlashes {
		outFlashes = append(outFlashes, inFlashes[i].(string))
	}

	println(key, outFlashes)

	return outFlashes
}

func addFlashes(r *http.Request, d *templates.Vars) {
	vars := *d
	vars["Errors"] = getFlashes(r, ErrorFlashKey)
	vars["Warnings"] = getFlashes(r, WarningFlashKey)
	vars["Infos"] = getFlashes(r, InfoFlashKey)
	vars["Successes"] = getFlashes(r, SuccessFlashKey)
}

func addConfig(r *http.Request, d *templates.Vars) {
	c := config.Config
	(*d)["SiteTitle"] = c.GlobalVars.SiteTitle
}

func init() {
	templates.Register(addConfig)
	templates.Register(addFlashes)

	http.HandleFunc("/", processHooks)
}
