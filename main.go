package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tummychow/goose/document"
	_ "github.com/tummychow/goose/document/file"
	"gopkg.in/unrolled/render.v1"
	"net/http"
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
	if _, ok := err.(document.NotFoundError); ok {
		masterStore.Update("/foo/bar", "#supdawg\nWelcome to the page **foo bar**\n```javascript\nvar foo = require('bar');\n```")
	}

	renderer := render.New(render.Options{
		IsDevelopment: len(os.Getenv("GOOSE_DEV")) != 0,
	})

	r := mux.NewRouter()
	r.StrictSlash(false)

	r.Methods("GET").Path("/public{_:/.*|$}").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	r.Methods("GET", "POST").Path("/w{_:/.+}").Handler(WikiController{masterStore, renderer})

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
