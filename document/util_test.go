package document_test

import (
	"github.com/tummychow/goose/document"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type UtilSuite struct{}

var _ = check.Suite(&UtilSuite{})

var namesTable = []struct {
	Name  string
	Valid bool
}{
	{"/Foo/Bar/Baz", true},
	{"/base/o(n)/10%/val Foo", true},
	{"/<>,.?;:'\"[]{}/\\|=+-_`~/!@#$%^&*()", true},
	{"/", false},
	{"Foo/Bar", false},
	{"/Foo/Bar/", false},
	{"", false},
	{"/Foo//Bar", false},
	{"/世/界/Bar", false},
	{"/./Bar", false},
	{"/Bar/.", false},
	{"/Bar/..", false},
	{"/../Bar", false},
	{"/.../Bar", true},
}

func (s *UtilSuite) TestNameValidation(c *check.C) {
	for _, entry := range namesTable {
		c.Check(document.ValidateName(entry.Name), check.DeepEquals, entry.Valid, check.Commentf("Name: %#v", entry.Name))
	}
}
