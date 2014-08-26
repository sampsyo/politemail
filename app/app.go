package app

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path"
)

import (
	"github.com/sampsyo/politemail/tmplpool"
)

type App struct {
	basedir   string
	templates *tmplpool.Pool
	DB        *bolt.DB
}

func (a *App) handleCompose(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Foo string
	}{
		"bar",
	}
	a.templates.Render(w, "compose", data)
}

type Message struct {
	To      string
	Subject string
	Body    string
	Options []string
}

func (a *App) handleMessage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msg := Message{
		r.FormValue("to"),
		r.FormValue("subject"),
		r.FormValue("body"),
		r.Form["option"],
	}
	a.addMessage(&msg)
	a.templates.Render(w, "confirm", msg)
}

func (a *App) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", a.handleCompose)
	r.HandleFunc("/message", a.handleMessage)
	staticdir := path.Join(a.basedir, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticdir)))
	return r
}

func New(basedir string, debug bool) *App {
	app := new(App)
	app.basedir = basedir

	app.templates = tmplpool.New(path.Join(basedir, "template"))
	app.templates.Debug = debug
	app.templates.Common = []string{"base"}
	app.templates.BaseDef = "base"

	db, err := bolt.Open(path.Join(basedir, "politemail.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("messages"))
		return err
	})
	app.DB = db

	return app
}

func (a *App) Teardown() {
	a.DB.Close()
}
