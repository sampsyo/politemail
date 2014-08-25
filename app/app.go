package app

import (
	"github.com/gorilla/mux"
	"net/http"
	"path"
)

import (
	"github.com/sampsyo/politemail/tmplpool"
)

type App struct {
	basedir   string
	templates *tmplpool.Pool
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

	return app
}
