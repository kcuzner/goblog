package site

import (
    "net/http"
    "github.com/gorilla/mux"
    "labix.org/v2/mgo"
    "bitbucket.org/kcuzner/goblog/site/config"
)

type Site struct {
    r *mux.Router
}

var site = Site{mux.NewRouter()}

func GetSite() *Site {
    return &site
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

func requireSession() *mgo.Session {
    c := config.GetConfiguration()

    session, err := mgo.Dial(c.ConnectionString)
    if err != nil {
        panic(err)
    }

    return session
}

var baseSession = requireSession()

// Gets the Mgo session for the passed request
// TODO: Change this to have per-session requests
func GetMgoSession() *mgo.Session {
    return baseSession
}

func init() {
    http.HandleFunc("/", processHooks)
}
