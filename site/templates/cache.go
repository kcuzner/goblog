package templates

import (
    "log"
    "path"
    "errors"
    "html/template"
    "bitbucket.org/kcuzner/goblog/site/config"
    "github.com/eknkc/amber"
    "github.com/howeyc/fsnotify"
)

type templateRequestType int

const (
    getRequest templateRequestType = iota
    refreshRequest
)

type templateResponse struct {
    Template *template.Template
    Error error
}

type templateRequest struct {
    Type templateRequestType
    Name string
    Response chan templateResponse
}

func (r *templateRequest) Respond(s templateResponse) {
    if r.Response != nil {
        r.Response <- s
    }
}

type TemplateCache struct {
    templateDir string
    requests chan templateRequest
}

var defaultOptions = amber.Options{true, false}

func NewTemplateCache() (*TemplateCache, error) {
    c := config.GetConfiguration()

    t := TemplateCache{c.TemplateDir, make(chan templateRequest, 50)}

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    //watcher events
    go func() {
        for {
            select {
            case ev := <- watcher.Event:
                if ev.IsModify() {
                    log.Println("Template cache: ", ev.Name, " modified")
                    t.requests <- templateRequest{refreshRequest, ev.Name, nil}
                }
            case err := <- watcher.Error:
                log.Println("Template cache: error: ", err)
            }
        }
    }()

    watcher.Watch(c.TemplateDir)

    //request events
    go func() {
        defer close(t.requests)
        var templates = make(map[string]*template.Template)

        watchedDirs := make(map[string]bool)

        for {
            request, ok := <- t.requests
            if !ok {
                return
            }

            switch request.Type {
            case getRequest:
                dir := path.Dir(request.Name)
                template, ok := templates[request.Name]
                if !ok || template == nil {
                    //fill in the blank so that it has a space to refresh into
                    templates[request.Name] = nil
                    //start watching this file
                    if isWatched, ok := watchedDirs[dir]; !ok || !isWatched {
                        watcher.Watch(dir)
                        watchedDirs[dir] = true
                    }
                    //load the template
                    log.Println("Template cache: loading ", request.Name)
                    template, err := amber.CompileFile(request.Name, defaultOptions)
                    if err == nil {
                        templates[request.Name] = template
                    }

                    request.Respond(templateResponse{template, err})    
                } else {
                    request.Respond(templateResponse{template, nil})
                }
            case refreshRequest:
                //see if the template existed
                _, ok := templates[request.Name]
                //clear the cache
                log.Println("Template cache: clearing")
                templates = make(map[string]*template.Template)
                if !ok {
                    request.Respond(templateResponse{nil, errors.New("Refresh for non-used template")})
                    break
                } else {
                    //requeue this as a get request
                    t.requests <- templateRequest{getRequest, request.Name, request.Response}
                }
            default:
                request.Respond(templateResponse{nil, errors.New("Unknown template request type")})
            }
        }
    }()

    return &t, nil
}

func RequireTemplateCache() *TemplateCache {
    c, err := NewTemplateCache()
    if err != nil {
        panic(err)
    }

    return c
}

func (t *TemplateCache) Get(name string) (*template.Template, error) {
    p := path.Join(t.templateDir, name + ".amber")
    request := templateRequest{getRequest, p, make(chan templateResponse, 1)}
    defer close(request.Response)

    t.requests <- request

    response := <- request.Response

    return response.Template, response.Error
}

var Cache = *RequireTemplateCache()

func init() {

}
