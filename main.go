package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	r.StrictSlash(false)

	r.Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
