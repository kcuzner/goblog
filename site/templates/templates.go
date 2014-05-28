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

    //request events
    go func() {
        defer close(t.requests)
        templates := make(map[string]*template.Template)

        watched := make(map[string]bool)

        for {
            request, ok := <- t.requests
            if !ok {
                return
            }

            if request.Type == getRequest {
                template, ok := templates[request.Name]
                if !ok {
                    log.Println("Template cache: not found ", request.Name)
                    //start watching this file
                    if isWatched, ok := watched[request.Name]; !ok || !isWatched {
                        watcher.Watch(request.Name)
                        watched[request.Name] = true
                    }
                    //requeue this as a refresh request
                    t.requests <- templateRequest{refreshRequest, request.Name, request.Response}
                } else {
                    request.Response <- templateResponse{template, nil}
                }
            } else if request.Type == refreshRequest {
                log.Println("Template cache: loading ", request.Name)
                template, err := amber.CompileFile(request.Name, defaultOptions)
                if err == nil {
                    templates[request.Name] = template
                }

                request.Respond(templateResponse{template, err})
            } else {
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
