package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tummychow/goose/document"
	_ "github.com/tummychow/goose/document/file"
	_ "github.com/tummychow/goose/document/sql"
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

	renderer := render.New(render.Options{
		IsDevelopment: len(os.Getenv("GOOSE_DEV")) != 0,
	})

	r := mux.NewRouter()
	r.StrictSlash(false)

	r.Methods("GET").Path("/public{_:/.*|$}").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))

	wcon := WikiController{
		Store:  masterStore,
		Render: renderer,
	}
	r.Methods("GET", "POST").Path("/w{_:/.+}").HandlerFunc(wcon.Show)
	r.Methods("GET").Path("/e{_:/.+}").HandlerFunc(wcon.Edit)
	r.Methods("POST").Path("/e{_:/.+}").HandlerFunc(wcon.Save)

	http.ListenAndServe(os.Getenv("GOOSE_PORT"), r)
}
