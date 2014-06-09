package db

import (
    "labix.org/v2/mgo/bson"
    "labix.org/v2/mgo"
    "bitbucket.org/kcuzner/goblog/site/config"
)

type Model interface {
    Collection() string
    Indexes() [][]string
}

type OrderedModel interface {
    Model
    Sorting() string
}

type Updatable interface {
    Model
    Unique() bson.M
    PreSave()
}

type repository struct {
    session *mgo.Session
    db *mgo.Database
}

func requireSession() *mgo.Session {
    c := config.Config

    session, err := mgo.Dial(c.ConnectionString)
    if err != nil {
        panic(err)
    }

    return session
}

var Current = newRepository()

// Creates a new repository
func newRepository() *repository {
    session := requireSession()
    db := session.DB("")

    repo := repository{session, db}

    return &repo
}

func (r *repository) registerIndexes(m Model) {
    collection := r.Cursor(m)
    indexes := m.Indexes()
    for _, v := range indexes {
        err := collection.EnsureIndex(mgo.Index{Key: v})
        if err != nil {
            panic(err)
        }
    }
}

func (r *repository) Cursor(m Model) *mgo.Collection {
    return r.db.C(m.Collection())
}

func (r *repository) Find(m Model, query interface{}) *mgo.Query {
    return r.Cursor(m).Find(query)
}

func (r *repository) Latest(o OrderedModel, query interface{}) *mgo.Query {
    return r.Find(o, query).Sort(o.Sorting())
}

func (r *repository) Exists(u Updatable) bool {
    var data interface{}
    err := r.Find(u, u.Unique()).One(&data)
    if err != nil {
        return false
    }
    return true
}

func (r *repository) Upsert(u Updatable) (info *mgo.ChangeInfo, err error) {
    u.PreSave()
    return r.Cursor(u).Upsert(u.Unique(), u)
}

var Models = []Model{}

func Register(m Model) {
    Models = append(Models, m)
    Current.registerIndexes(m)
}
