package document_test

import (
	"github.com/tummychow/goose/document"
	"gopkg.in/check.v1"
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

	if len(document.NameToSegments(doc.Name)) == 0 {
		return false, "obtained Document has invalid Name"
	}
	if len([]byte(doc.Content)) >= document.MAX_CONTENT_SIZE {
		return false, "obtained Document has oversized Content"
	}
	if doc.Timestamp.Location() != time.UTC {
		return false, "obtained Document has non-UTC Timestamp"
	}
	if doc.Source == nil {
		return false, "obtained Document has nil Source"
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
	s.Store.Revert("/foo/bar", time.Time{})
}
func (s *DocumentStoreSuite) TearDownSuite(c *check.C) {
	s.Store.Close()
}

func (s *DocumentStoreSuite) TestEmpty(c *check.C) {
	_, err := s.Store.Get("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.DocumentNotFoundError{})

	docAll, err := s.Store.GetAll("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.DocumentNotFoundError{})
	c.Assert(docAll, check.HasLen, 0)
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
	c.Assert(err, check.FitsTypeOf, document.DocumentNotFoundError{})
	c.Assert(reverts, check.Equals, 0)

	doc, err = s.Store.Get("/foo/bar")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.FitsTypeOf, document.DocumentNotFoundError{})
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
	c.Assert(err, check.FitsTypeOf, document.DocumentNotFoundError{})
}