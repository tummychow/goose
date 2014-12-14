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

func (c WikiController) Show(w http.ResponseWriter, r *http.Request) {
	doc, unknownErr := c.handleDocument(r)

	switch err := unknownErr.(type) {
	case nil:
		c.Render.HTML(w, http.StatusOK, "wikipage", doc)
	case document.NotFoundError:
		c.Render.HTML(w, http.StatusNotFound, "wiki404", err.Name)
	default:
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", err.Error())
	}
}

func (c WikiController) Edit(w http.ResponseWriter, r *http.Request) {
	doc, unknownErr := c.handleDocument(r)

	switch err := unknownErr.(type) {
	case nil:
		c.Render.HTML(w, http.StatusOK, "wikiedit", doc)
	case document.NotFoundError:
		c.Render.HTML(w, http.StatusNotFound, "wikiedit", map[string]string{
			"Name":    err.Name,
			"Content": "",
		})
	default:
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", err.Error())
	}
}

func (c WikiController) Save(w http.ResponseWriter, r *http.Request) {
	doc, unknownErr := c.handleDocument(r)

	switch err := unknownErr.(type) {
	case nil:
		http.Redirect(w, r, "/w"+doc.Name, http.StatusMovedPermanently)
	default:
		c.Render.HTML(w, http.StatusInternalServerError, "wiki500", err.Error())
	}
}

func (c WikiController) handleDocument(r *http.Request) (document.Document, error) {
	store, targetName, err := c.pre(r)
	if err != nil {
		return document.Document{}, err
	}
	defer store.Close()

	if newContent := r.PostFormValue("content"); len(newContent) > 0 {
		err = store.Update(targetName, newContent)
		if err != nil {
			return document.Document{}, err
		}
	}

	return store.Get(targetName)
}

func (c WikiController) pre(r *http.Request) (document.DocumentStore, string, error) {
	store, err := c.Store.Copy()
	if err != nil {
		return nil, "", err
	}

	targetName, err := url.QueryUnescape(r.URL.Path[2:])
	if err != nil {
		return nil, "", err
	}
	// gorilla invokes path.Clean already but it restores trailing slashes,
	// and we need to remove those
	// https://github.com/gorilla/mux/blob/master/mux.go#L69
	targetName = path.Clean(targetName)
	if targetName == "." || targetName == "/" {
		targetName = ""
	}

	return store, targetName, nil
}
