package site

import (
    "net/http"
    "log"
    "bitbucket.org/kcuzner/goblog/site/templates"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
)

type Site struct {
    r *mux.Router
}

var site = Site{mux.NewRouter()}

func GetSite() *Site {
    return &site
}

const MainSessionName = "session"

var Store = sessions.NewCookieStore([]byte("Jh!$xPnz6=YeR+N"))

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
func RenderTemplate(w http.ResponseWriter, r *http.Request, name string, f func(http.ResponseWriter, *http.Request, templates.GlobalVars) (interface{}, error)) {
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
    defer func(d interface{}) { tmpl.Execute(w, d) }(&data)

    session, err := Store.Get(r, MainSessionName)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("renderTemplate:", err)
        return
    }
    defer session.Save(r, w)

    vars, err := GetContextVariables(r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("renderTemplate:", err)
        return
    }

    data, err = f(w, r, templates.GetGlobalVars(vars))
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("renderTemplate:", err)
        return
    }
}

func init() {
    http.HandleFunc("/", processHooks)
}
