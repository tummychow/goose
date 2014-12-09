// Package file provides an implementation of DocumentStore using flat files.
package file

import (
	"fmt"
	"github.com/tummychow/goose/document"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func init() {
	document.RegisterStore("file", func(target *url.URL) (document.DocumentStore, error) {
		if len(target.Host) != 0 {
			return nil, fmt.Errorf("goose/document/file: unexpected URI host %q", target.Host)
		}
		return &FileDocumentStore{root: filepath.Clean(target.Path), mutex: &sync.RWMutex{}}, nil
	})
}

// similar to RFC3339Nano, but with trailing nanosecond zeroes preserved
var fileTimeFormat = "2006-01-02T15:04:05.000000000Z07:00"

// FileDocumentStore is an implementation of DocumentStore, using a standard
// UNIX filesystem. A Document corresponds to a folder on the filesystem, with
// each version corresponding to an individual file under that folder.
//
// FileDocumentStore is registered with the scheme "file". For example, you can
// initialize a new FileDocumentStore via:
//
//     import "github.com/tummychow/goose/document"
//     import _ "github.com/tummychow/goose/document/file"
//     store, err := document.NewStore("file:///var/goose/docs")
//
// This would return a FileDocumentStore using the files under /var/goose/docs.
// FileDocumentStore's URI format takes no options, hosts or user info. The
// target folder must be on the current system, readable and writable to the
// user under which Goose is running.
//
// The path must be absolute. For example, "file://goose/docs" is invalid,
// because "goose" would be interpreted as the host and "/docs" would be the
// path. To avoid this mistake, a nonempty host string in the URI will raise an
// error at instantiation time.
//
// FileDocumentStore is primarily for development. It has poor performance, and
// it can potentially block forever or behave inconsistently because it depends
// on a mutex that is shared across copies. If two separate non-copy instances
// of FileDocumentStore are initialized with the same URI, incorrect behaviors
// could occur.
//
// FileDocumentStore does not support Windows. The characters \/:*?"<>| are
// forbidden in Windows filenames, but most of these are legal in a Document's
// Name, which would create issues when trying to store such a Document on a
// Windows filesystem.
//
// The flat file schema for FileDocumentStore is an implementation detail, and
// modifications to the schema are not considered breaking. Do not rely on the
// schema.
type FileDocumentStore struct {
	// root is the root directory of the FileDocumentStore.
	root string
	// mutex is the global mutex shared between this FileDocumentStore and all
	// its copies.
	mutex *sync.RWMutex
}

func (s *FileDocumentStore) Close() {}

func (s *FileDocumentStore) Copy() (document.DocumentStore, error) {
	return &FileDocumentStore{root: s.root, mutex: s.mutex}, nil
}

func (s *FileDocumentStore) Get(name string) (document.Document, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	docdir, err := s.readDirFiles(name)
	if err != nil {
		return document.Document{}, err
	}

	return s.readDocument(name, docdir[len(docdir)-1])
}

func (s *FileDocumentStore) GetAll(name string) ([]document.Document, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	docdir, err := s.readDirFiles(name)
	if err != nil {
		return []document.Document{}, err
	}

	ret := make([]document.Document, 0, len(docdir))
	for i := len(docdir) - 1; i >= 0; i-- {
		doc, err := s.readDocument(name, docdir[i])
		if err != nil {
			return []document.Document{}, err
		}
		ret = append(ret, doc)
	}

	return ret, nil
}

func (s *FileDocumentStore) Update(name, content string) error {
	// Update has to check the name before attempting to write the file
	if !document.ValidateName(name) {
		return document.InvalidNameError{name}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := os.MkdirAll(filepath.Join(s.root, name), 0755)
	if err != nil {
		return err
	}

	docstamp := time.Now().UTC()
	err = ioutil.WriteFile(filepath.Join(s.root, name, docstamp.Format(fileTimeFormat)), []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileDocumentStore) Clear() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	contents, err := ioutil.ReadDir(s.root)
	if err != nil {
		return err
	}

	// delete each file or folder in the root, but not the root itself
	for _, target := range contents {
		err = os.RemoveAll(filepath.Join(s.root, target.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// readDirFiles returns the sorted list of files for the named Document, from
// oldest to newest. Returns NotFoundError or InvalidNameError as needed.
func (s *FileDocumentStore) readDirFiles(name string) ([]os.FileInfo, error) {
	if !document.ValidateName(name) {
		return []os.FileInfo{}, document.InvalidNameError{name}
	}

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return []os.FileInfo{}, document.NotFoundError{name}
			}
		}
		return []os.FileInfo{}, err
	}

	ret := make([]os.FileInfo, 0, len(docdir))
	for _, fileinfo := range docdir {
		if fileinfo.IsDir() {
			continue
		}
		ret = append(ret, fileinfo)
	}
	if len(ret) == 0 {
		return []os.FileInfo{}, document.NotFoundError{name}
	}
	return ret, nil
}

// readDocument takes a single file and unmarshals it into a Document. It does
// not perform name validation, since the target file should be obtained from
// readDirFiles (which does the name validation for you).
func (s *FileDocumentStore) readDocument(name string, target os.FileInfo) (document.Document, error) {
	timestamp, err := time.Parse(fileTimeFormat, target.Name())
	if err != nil {
		return document.Document{}, err
	}
	content, err := ioutil.ReadFile(filepath.Join(s.root, name, target.Name()))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return document.Document{}, document.NotFoundError{name}
			}
		}
		return document.Document{}, err
	}
	return document.Document{
		Name:      name,
		Content:   string(content),
		Timestamp: timestamp,
	}, nil
}
