package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sampsyo/politemail/tmplpool"
	"net/http"
	"path"
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

func (a *App) handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", a.handleCompose)
	r.HandleFunc("/message", a.handleMessage)
	staticdir := path.Join(a.basedir, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticdir)))
	return r
}

func NewApp(basedir string, debug bool) *App {
	app := new(App)
	app.basedir = basedir

	app.templates = tmplpool.New(path.Join(basedir, "template"))
	app.templates.Debug = debug
	app.templates.Common = []string{"base"}
	app.templates.BaseDef = "base"

	return app
}

func main() {
	debug := flag.Bool("debug", false, "always reload templates")
	flag.Parse()

	app := NewApp(".", *debug)
	fmt.Println("http://0.0.0.0:8080")
	http.Handle("/", app.handler())
	http.ListenAndServe(":8080", nil)
}
