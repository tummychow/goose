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
// FileDocumentStore is primarily for development. It has several limitations:
// - no Windows support (because \/:*?"<>| are forbidden in Windows filenames)
// - no concurrency support; only one instance and its copies are allowed at a
//   time. This is implemented via a mutex shared between all the copies. A
//   future improvement would be to use lock files.
// - locks are at the level of the entire instance, not at the level of an
//   individual Document. Modifying one Document will lock all the others, even
//   though they are not affected.
// - mutex locks are acquired without the use of timeouts, so an operation can
//   theoretically block forever in an extreme error case.
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
// path. To avoid this error, a nonempty host string in the URI will result in
// an error at instantiation time.
type FileDocumentStore struct {
	// root is the root directory of the FileDocumentStore.
	root string
	// mutex is the global mutex shared between this FileDocumentStore and all its
	// copies.
	mutex *sync.RWMutex
}

func (s *FileDocumentStore) Close() {}

func (s *FileDocumentStore) Copy() (document.DocumentStore, error) {
	return &FileDocumentStore{root: s.root, mutex: s.mutex}, nil
}

func (s *FileDocumentStore) Get(name string) (document.Document, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return document.Document{}, document.DocumentNotFoundError{name}
			}
		}
		return document.Document{}, err
	}
	if len(docdir) == 0 {
		return document.Document{}, document.DocumentNotFoundError{name}
	}

	targetName := docdir[len(docdir)-1].Name()
	target, err := ioutil.ReadFile(filepath.Join(s.root, name, targetName))
	if err != nil {
		return document.Document{}, err
	}
	docstamp, err := time.Parse(fileTimeFormat, targetName)
	if err != nil {
		return document.Document{}, err
	}

	return document.Document{
		Name:      name,
		Content:   string(target),
		Timestamp: docstamp,
		Source:    s,
	}, nil
}

func (s *FileDocumentStore) GetAll(name string) ([]document.Document, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return []document.Document{}, document.DocumentNotFoundError{name}
			}
		}
		return []document.Document{}, err
	}
	if len(docdir) == 0 {
		return []document.Document{}, document.DocumentNotFoundError{name}
	}

	ret := make([]document.Document, 0, len(docdir))
	for i := len(docdir) - 1; i >= 0; i-- {
		targetName := docdir[i].Name()
		target, err := ioutil.ReadFile(filepath.Join(s.root, name, targetName))
		if err != nil {
			return []document.Document{}, err
		}
		docstamp, err := time.Parse(fileTimeFormat, targetName)
		if err != nil {
			return []document.Document{}, err
		}
		ret = append(ret, document.Document{
			Name:      name,
			Content:   string(target),
			Timestamp: docstamp,
			Source:    s,
		})
	}

	return ret, nil
}

func (s *FileDocumentStore) Update(name, content string) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := os.MkdirAll(filepath.Join(s.root, name), 0755)
	if err != nil {
		return 0, err
	}

	docstamp := time.Now().UTC()
	err = ioutil.WriteFile(filepath.Join(s.root, name, docstamp.Format(fileTimeFormat)), []byte(content), 0644)
	if err != nil {
		return 0, err
	}

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		return 0, err
	}
	return len(docdir), nil
}

func (s *FileDocumentStore) Revert(name string, version time.Time) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return 0, document.DocumentNotFoundError{name}
			}
		}
		return 0, err
	}
	if len(docdir) == 0 {
		return 0, document.DocumentNotFoundError{name}
	}

	i := len(docdir) - 1 // we need this outside the loop scope
	for ; i >= 0; i-- {
		docstamp, err := time.Parse(fileTimeFormat, docdir[i].Name())
		if err != nil {
			return 0, err
		}
		if docstamp.Equal(version) || docstamp.After(version) {
			err := os.Remove(filepath.Join(s.root, name, docdir[i].Name()))
			if err != nil {
				return 0, err
			}
		} else {
			break
		}
	}
	return len(docdir) - i - 1, nil
}

func (s *FileDocumentStore) Truncate(name string, version time.Time) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	docdir, err := ioutil.ReadDir(filepath.Join(s.root, name))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err.Error() == "no such file or directory" {
				return 0, document.DocumentNotFoundError{name}
			}
		}
		return 0, err
	}
	if len(docdir) == 0 {
		return 0, document.DocumentNotFoundError{name}
	}

	i := 0 // we need this outside the loop scope
	for ; i < len(docdir); i++ {
		docstamp, err := time.Parse(fileTimeFormat, docdir[i].Name())
		if err != nil {
			return 0, err
		}
		if docstamp.Equal(version) || docstamp.Before(version) {
			err := os.Remove(filepath.Join(s.root, name, docdir[i].Name()))
			if err != nil {
				return 0, err
			}
		} else {
			break
		}
	}
	return i, nil
}
