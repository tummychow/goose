package main

import (
	"github.com/tummychow/goose/document"
	"gopkg.in/unrolled/render.v1"
	"net/http"
	"net/url"
	"path"
)

type WikiController struct {
	Store  document.DocumentStore
	Render *render.Render
}

func (c WikiController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	doc, err := c.getDocument(r)
	if docErr, ok := err.(document.NotFoundError); ok {
		c.Render.HTML(w, http.StatusNotFound, "wiki404", map[string]interface{}{
			"Title": docErr.Name,
			"Name":  docErr.Name,
		})
	} else if err != nil {
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
			"Title": "Error",
			"Error": err.Error(),
		})
	} else {
		c.Render.HTML(w, http.StatusOK, "wikipage", map[string]interface{}{
			"Title": doc.Name,
			"Doc":   doc,
		})
	}
}

func (c WikiController) getDocument(r *http.Request) (document.Document, error) {
	store, err := c.Store.Copy()
	if err != nil {
		return document.Document{}, err
	}
	defer store.Close()

	targetName, err := url.QueryUnescape(r.URL.Path[2:])
	if err != nil {
		return document.Document{}, err
	}
	// gorilla invokes path.Clean already but it restores trailing slashes,
	// and we need to remove those
	// https://github.com/gorilla/mux/blob/master/mux.go#L69
	targetName = path.Clean(targetName)

	return store.Get(targetName)
}
