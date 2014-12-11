// Package sql provides an implementation of DocumentStore using a PostgreSQL
// database.
package sql

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/tummychow/goose/document"
	"net/url"
)

func init() {
	document.RegisterStore("postgres", func(target *url.URL) (document.DocumentStore, error) {
		db, err := sql.Open("postgres", target.String())
		if err != nil {
			return nil, err
		}

		get, err := db.Prepare("SELECT name, content, stamp FROM documents WHERE name = $1 ORDER BY stamp DESC LIMIT 1;")
		if err != nil {
			db.Close()
			return nil, err
		}

		getAll, err := db.Prepare("SELECT name, content, stamp FROM documents WHERE name = $1 ORDER BY stamp DESC;")
		if err != nil {
			get.Close()
			db.Close()
			return nil, err
		}

		getDescendants, err := db.Prepare(`
			SELECT DISTINCT name
			    FROM documents
			    WHERE name LIKE ($1 || '/%')
			    ORDER by name ASC;`)
		if err != nil {
			get.Close()
			getAll.Close()
			db.Close()
			return nil, err
		}

		update, err := db.Prepare("INSERT INTO documents (name, content) VALUES ($1, $2);")
		if err != nil {
			get.Close()
			getAll.Close()
			getDescendants.Close()
			db.Close()
			return nil, err
		}

		return &SqlDocumentStore{
			db:             db,
			get:            get,
			getAll:         getAll,
			getDescendants: getDescendants,
			update:         update,
			refcount:       1,
		}, nil
	})
}

// SqlDocumentStore is an implementation of DocumentStore, using a standard SQL
// database. Currently, only PostgreSQL is supported.
//
// SqlDocumentStore is registered with the scheme "postgres". For example, you
// can initialize a new SqlDocumentStore via:
//
//     import "github.com/tummychow/goose/document"
//     import _ "github.com/tummychow/goose/document/sql"
//     store, err := document.NewStore("postgres://gooser:goosepw@localhost:5432/goosedb")
//
// The URI is passed directly to the Go PostgreSQL driver, lib/pq. Refer to its
// documentation for more details (http://godoc.org/github.com/lib/pq and
// http://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING).
//
// SqlDocumentStore expects the database to be using a UTF-8 locale. It should
// contain the following table:
//
//     CREATE TABLE documents (
//         name TEXT NOT NULL,
//         content TEXT NOT NULL,
//         stamp TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
//         PRIMARY KEY (name, stamp)
//     );
type SqlDocumentStore struct {
	db             *sql.DB
	get            *sql.Stmt
	getAll         *sql.Stmt
	getDescendants *sql.Stmt
	update         *sql.Stmt
	refcount       int
}

func (s *SqlDocumentStore) Close() {
	s.refcount--
	if s.refcount == 0 {
		s.get.Close()
		s.getAll.Close()
		s.getDescendants.Close()
		s.update.Close()

		s.db.Close()
	}
}

func (s *SqlDocumentStore) Copy() (document.DocumentStore, error) {
	s.refcount++
	return s, nil
}

func (s *SqlDocumentStore) Get(name string) (document.Document, error) {
	if !document.ValidateName(name) {
		return document.Document{}, document.InvalidNameError{name}
	}

	ret := document.Document{}
	row := s.get.QueryRow(name)

	err := row.Scan(&ret.Name, &ret.Content, &ret.Timestamp)
	if err == sql.ErrNoRows {
		return document.Document{}, document.NotFoundError{name}
	} else if err != nil {
		return document.Document{}, err
	}

	ret.Timestamp = ret.Timestamp.UTC()
	return ret, nil
}

func (s *SqlDocumentStore) GetAll(name string) ([]document.Document, error) {
	if !document.ValidateName(name) {
		return []document.Document{}, document.InvalidNameError{name}
	}

	rows, err := s.getAll.Query(name)
	if err != nil {
		return []document.Document{}, err
	}
	defer rows.Close()

	ret := []document.Document{}
	for rows.Next() {
		cur := document.Document{}

		err = rows.Scan(&cur.Name, &cur.Content, &cur.Timestamp)
		if err != nil {
			return []document.Document{}, err
		}

		cur.Timestamp = cur.Timestamp.UTC()
		ret = append(ret, cur)
	}

	err = rows.Err()
	if err != nil {
		return []document.Document{}, err
	}

	if len(ret) == 0 {
		return []document.Document{}, document.NotFoundError{name}
	}
	return ret, nil
}

func (s *SqlDocumentStore) GetDescendants(ancestor string) ([]string, error) {
	if ancestor != "" && !document.ValidateName(ancestor) {
		return []string{}, document.InvalidNameError{ancestor}
	}

	rows, err := s.getDescendants.Query(ancestor)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	ret := []string{}
	for rows.Next() {
		cur := ""
		err := rows.Scan(&cur)
		if err != nil {
			return []string{}, err
		}
		ret = append(ret, cur)
	}

	err = rows.Err()
	if err != nil {
		return []string{}, err
	}

	return ret, nil
}

func (s *SqlDocumentStore) Update(name, content string) error {
	if !document.ValidateName(name) {
		return document.InvalidNameError{name}
	}

	_, err := s.update.Exec(name, content)
	return err
}

func (s *SqlDocumentStore) Clear() error {
	_, err := s.db.Exec("DELETE FROM documents;")
	return err
}
