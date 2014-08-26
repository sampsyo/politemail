package main

import (
	"flag"
	"fmt"
	"net/http"
)

import (
	"github.com/sampsyo/politemail/app"
)

func main() {
	debug := flag.Bool("debug", false, "always reload templates")
	flag.Parse()

	app := app.New(".", *debug)
	defer app.Teardown()

	fmt.Println("http://0.0.0.0:8080")
	http.Handle("/", app.Handler())
	http.ListenAndServe(":8080", nil)
}
