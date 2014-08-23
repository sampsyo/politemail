package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func ComposeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "a test")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", ComposeHandler)

	fmt.Println("http://0.0.0.0:8080")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
