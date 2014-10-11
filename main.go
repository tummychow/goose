package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tummychow/goose/document"
	_ "github.com/tummychow/goose/document/file"
	"gopkg.in/unrolled/render.v1"
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

	t := render.New(render.Options{Layout: "layout"})

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
			t.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
				"Title": "Error",
				"Error": err.Error(),
			})
			return
		}

		store, err := masterStore.Copy()
		if err != nil {
			t.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
				"Title": "Error",
				"Error": err.Error(),
			})
			return
		}
		defer store.Close()

		doc, err := store.Get(targetName)
		if _, ok := err.(document.DocumentNotFoundError); ok {
			t.HTML(w, http.StatusNotFound, "wiki404", map[string]interface{}{
				"Title": targetName,
				"Name":  targetName,
			})
			return
		} else if err != nil {
			t.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
				"Title": "Error",
				"Error": err.Error(),
			})
			return
		}

		t.HTML(w, http.StatusOK, "wikipage", map[string]interface{}{
			"Title": doc.Name,
			"Doc":   doc,
		})
	})

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
