package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	r.StrictSlash(false)

	// TODO: route "/public" to "/public/" instead of 404
	r.Methods("GET").PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// TODO: route "/w" to "/w/" (both of which should route to "/") instead of 404
	r.Methods("GET").PathPrefix("/w/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/w/" {
			http.Redirect(w, r, "..", 301)
			return
		}
		fmt.Fprintln(w, "Welcome to the wiki")
	})

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
