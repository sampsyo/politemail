package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

var cache = NewCache(".")

func handleCompose(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Foo string
	}{
		"bar",
	}
	cache.Render(w, "compose", data)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleCompose)

	fmt.Println("http://0.0.0.0:8080")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
