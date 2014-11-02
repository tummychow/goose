package main

import (
	"github.com/tummychow/goose/document"
	"gopkg.in/unrolled/render.v1"
	"net/http"
	"net/url"
)

type WikiController struct {
	Store  document.DocumentStore
	Render *render.Render
}

func (c WikiController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/w/" {
		http.Redirect(w, r, "..", 301)
		return
	}

	targetName, err := url.QueryUnescape(r.URL.Path[2:])
	if err != nil {
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
			"Title": "Error",
			"Error": err.Error(),
		})
		return
	}

	store, err := c.Store.Copy()
	if err != nil {
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
			"Title": "Error",
			"Error": err.Error(),
		})
		return
	}
	defer store.Close()

	doc, err := store.Get(targetName)
	if _, ok := err.(document.DocumentNotFoundError); ok {
		c.Render.HTML(w, http.StatusNotFound, "wiki404", map[string]interface{}{
			"Title": targetName,
			"Name":  targetName,
		})
		return
	} else if err != nil {
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", map[string]interface{}{
			"Title": "Error",
			"Error": err.Error(),
		})
		return
	}

	c.Render.HTML(w, http.StatusOK, "wikipage", map[string]interface{}{
		"Title": doc.Name,
		"Doc":   doc,
	})
}
