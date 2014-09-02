package document

import (
	"fmt"
	"net/url"
)

var implMap = map[string]func(*url.URL) (DocumentStore, error){}

// RegisterStore maps a given URI scheme to a function that creates
// DocumentStores using that URI scheme. It should be called in the init()
// function of a package containing a DocumentStore implementation.
//
// For example, the MongoDocumentStore package calls RegisterStore with the
// scheme "mongodb". It passes a function that takes a MongoDB connection URI
// and returns a new DocumentStore wrapping around that MongoDB instance.
func RegisterStore(scheme string, factory func(*url.URL) (DocumentStore, error)) {
	if factory == nil {
		panic("goose/document: Register function is nil")
	}
	if _, exists := implMap[scheme]; exists {
		panic("goose/document: Register called twice for scheme " + scheme)
	}
	implMap[scheme] = factory
}

// NewStore takes a URI, representing a connection string for a storage system
// of some kind, and returns a DocumentStore that uses that storage system.
func NewStore(uri string) (DocumentStore, error) {
	parsedUri, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, fmt.Errorf("goose/document: %q is not a URI", uri)
	}

	factory, ok := implMap[parsedUri.Scheme]
	if !ok {
		return nil, fmt.Errorf("goose/document: unknown scheme %q (forgotten import?)", parsedUri.Scheme)
	}

	return factory(parsedUri)
}
