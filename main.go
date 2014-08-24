package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sampsyo/politemail/tmplpool"
	"net/http"
)

var cache = tmplpool.New("./template")

func handleCompose(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Foo string
	}{
		"bar",
	}
	cache.Render(w, "compose", data)
}

type Message struct {
	To      string
	Subject string
	Body    string
	Options []string
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msg := Message{
		r.FormValue("to"),
		r.FormValue("subject"),
		r.FormValue("body"),
		r.Form["option"],
	}
	cache.Render(w, "confirm", msg)
}

func main() {
	debug := flag.Bool("debug", false, "always reload templates")
	flag.Parse()

	cache.Debug = *debug
	cache.Common = []string{"base"}
	cache.BaseDef = "base"

	r := mux.NewRouter()
	r.HandleFunc("/", handleCompose)
	r.HandleFunc("/message", handleMessage)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	fmt.Println("http://0.0.0.0:8080")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
