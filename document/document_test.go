package document_test

import (
	"fmt"
	"github.com/tummychow/goose/document"
	_ "github.com/tummychow/goose/document/file"
	"gopkg.in/check.v1"
	"os"
	"time"
)

type DocumentStoreSuite struct {
	Store document.DocumentStore
}

type documentChecker struct {
	*check.CheckerInfo
}

var DocumentEquals check.Checker = &documentChecker{
	&check.CheckerInfo{Name: "DocumentEquals", Params: []string{"obtained", "Name", "Content"}},
}

func init() {
	if len(os.Getenv("GOOSE_TEST_FILE")) != 0 {
		fileStore, err := document.NewStore(os.Getenv("GOOSE_TEST_FILE"))
		if err != nil {
			fmt.Printf("Could not initialize FileDocumentStore %q, skipping\n(error was: %v)\n", os.Getenv("GOOSE_TEST_FILE"), err)
		} else {
			fmt.Printf("Running tests against FileDocumentStore %q\n", os.Getenv("GOOSE_TEST_FILE"))
			check.Suite(&DocumentStoreSuite{Store: fileStore})
		}
	}
}

// Check compares a Document against an expected Name and Content. The Document
// is checked for validity, and the Name and Content are then matched. Passing
// nil for the Name or Content will cause that comparison to be skipped.
func (checker *documentChecker) Check(params []interface{}, names []string) (result bool, error string) {
	if params[0] == nil {
		return false, "obtained value is nil"
	}
	doc, ok := params[0].(document.Document)
	if !ok {
		return false, "obtained value is not a Document"
	}

	if !document.ValidateName(doc.Name) {
		return false, "obtained Document has invalid Name"
	}
	if len([]byte(doc.Content)) >= document.MAX_CONTENT_SIZE {
		return false, "obtained Document has oversized Content"
	}
	if doc.Timestamp.Location() != time.UTC {
		return false, "obtained Document has non-UTC Timestamp"
	}

	if params[1] != nil {
		expectedName, ok := params[1].(string)
		if !ok {
			return false, "Name is not a string"
		}
		if doc.Name != expectedName {
			return false, "obtained Document has wrong Name"
		}
	}
	if params[2] != nil {
		expectedContent, ok := params[2].(string)
		if !ok {
			return false, "Content is not a string"
		}
		if doc.Content != expectedContent {
			return false, "obtained Document has wrong Content"
		}
	}

	return true, ""
}

func (s *DocumentStoreSuite) SetUpTest(c *check.C) {
	s.Store.Revert("/foo", time.Time{})
	s.Store.Revert("/foo/bar", time.Time{})
	s.Store.Revert("/foo/bar/baz", time.Time{})
}
func (s *DocumentStoreSuite) TearDownSuite(c *check.C) {
	s.Store.Close()
}

func (s *DocumentStoreSuite) TestEmpty(c *check.C) {
	_, err := s.Store.Get("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})

	docAll, err := s.Store.GetAll("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
	c.Assert(docAll, check.HasLen, 0)

	reverts, err := s.Store.Revert("/foo/bar", time.Time{})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
	c.Assert(reverts, check.Equals, 0)

	truncates, err := s.Store.Truncate("/foo/bar", time.Time{})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
	c.Assert(truncates, check.Equals, 0)
}

func (s *DocumentStoreSuite) TestInvalidNames(c *check.C) {
	_, err := s.Store.Get("/foo/bar/")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.InvalidNameError{})

	docAll, err := s.Store.GetAll("/foo/bar/")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.InvalidNameError{})
	c.Assert(docAll, check.HasLen, 0)

	_, err = s.Store.Update("/foo/bar/", "foo bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.InvalidNameError{})

	reverts, err := s.Store.Revert("/foo/bar/", time.Time{})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.InvalidNameError{})
	c.Assert(reverts, check.Equals, 0)

	truncates, err := s.Store.Truncate("/foo/bar/", time.Time{})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.InvalidNameError{})
	c.Assert(truncates, check.Equals, 0)
}

func (s *DocumentStoreSuite) TestBasic(c *check.C) {
	ver, err := s.Store.Update("/foo/bar", "foo bar")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 1)

	doc, err := s.Store.Get("/foo/bar")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo/bar", "foo bar")

	reverts, err := s.Store.Revert("/foo/bar", doc.Timestamp)
	c.Assert(err, check.IsNil)
	c.Assert(reverts, check.Equals, 1)

	reverts, err = s.Store.Revert("/foo/bar", doc.Timestamp)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
	c.Assert(reverts, check.Equals, 0)

	doc, err = s.Store.Get("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
}

func (s *DocumentStoreSuite) TestMultipleVersions(c *check.C) {
	ver, err := s.Store.Update("/foo/bar", "the duck quacked")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 1)

	ver, err = s.Store.Update("/foo/bar", "qux and baz oh my")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 2)

	doc, err := s.Store.Get("/foo/bar")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo/bar", "qux and baz oh my")

	docAll, err := s.Store.GetAll("/foo/bar")
	c.Assert(err, check.IsNil)
	c.Assert(docAll, check.HasLen, 2)
	c.Assert(docAll[0], DocumentEquals, "/foo/bar", "qux and baz oh my")
	c.Assert(docAll[1], DocumentEquals, "/foo/bar", "the duck quacked")

	reverts, err := s.Store.Revert("/foo/bar", docAll[0].Timestamp)
	c.Assert(err, check.IsNil)
	c.Assert(reverts, check.Equals, 1)

	doc, err = s.Store.Get("/foo/bar")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo/bar", "the duck quacked")

	truncates, err := s.Store.Truncate("/foo/bar", docAll[1].Timestamp)
	c.Assert(err, check.IsNil)
	c.Assert(truncates, check.Equals, 1)

	doc, err = s.Store.Get("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.NotFoundError{})
}

func (s *DocumentStoreSuite) TestMultipleDocuments(c *check.C) {
	ver, err := s.Store.Update("/foo", "foo v1")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 1)
	ver, err = s.Store.Update("/foo", "foo v2")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 2)
	ver, err = s.Store.Update("/foo/bar", "bar v1")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 1)
	ver, err = s.Store.Update("/foo/bar", "bar v2")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 2)
	ver, err = s.Store.Update("/foo/bar/baz", "baz v1")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 1)
	ver, err = s.Store.Update("/foo/bar/baz", "baz v2")
	c.Assert(err, check.IsNil)
	c.Assert(ver, check.Equals, 2)

	doc, err := s.Store.Get("/foo")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo", "foo v2")
	doc, err = s.Store.Get("/foo/bar")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo/bar", "bar v2")
	doc, err = s.Store.Get("/foo/bar/baz")
	c.Assert(err, check.IsNil)
	c.Assert(doc, DocumentEquals, "/foo/bar/baz", "baz v2")
}
