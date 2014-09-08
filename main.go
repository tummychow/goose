package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tummychow/goose/document"
	_ "github.com/tummychow/goose/document/file"
	"net/http"
	"net/url"
	"os"
)

// Initializes the DocumentStore instance that the application will use. If the
// initialization is unsuccessful, the program will exit from this function.
// The backend data must be a URI stored in the GOOSE_BACKEND environment
// variable.
func initializeStore() document.DocumentStore {
	backendURI := os.Getenv("GOOSE_BACKEND")
	if len(backendURI) == 0 {
		fmt.Println("GOOSE_BACKEND not defined")
		os.Exit(1)
	}

	ret, err := document.NewStore(backendURI)
	if err != nil {
		fmt.Printf("Error while initializing GOOSE_BACKEND=%q\n%v\n", backendURI, err)
		os.Exit(1)
	}
	return ret
}

func main() {
	masterStore := initializeStore()
	defer masterStore.Close()
	_, err := masterStore.Get("/foo/bar")
	if _, ok := err.(document.DocumentNotFoundError); ok {
		masterStore.Update("/foo/bar", "Welcome to the page foo bar")
	}

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

		targetName, err := url.QueryUnescape(r.URL.Path[2:])
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		store, err := masterStore.Copy()
		if err != nil {
			http.Error(w, "Could not copy DocumentStore", 500)
			return
		}
		defer store.Close()

		doc, err := store.Get(targetName)
		if _, ok := err.(document.DocumentNotFoundError); ok {
			w.WriteHeader(404)
			fmt.Fprintf(w, "You requested page %q, but it doesn't exist", targetName)
			return
		} else if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "Requested\n%q\n\nName\n%v\n\nTimestamp\n%v\n\nContents\n%v", targetName, doc.Name, doc.Timestamp.Local().Format("Jan 2 2006 15:04:05"), doc.Content)
	})

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
