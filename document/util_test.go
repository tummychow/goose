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
	Name     string
	Segments []string
}{
	{"/Foo/Bar/Baz", []string{"Foo", "Bar", "Baz"}},
	{"/base/o(n)/10%/val Foo", []string{"base", "o(n)", "10%", "val Foo"}},
	{"/<>,.?;:'\"[]{}/\\|=+-_`~/!@#$%^&*()", []string{"<>,.?;:'\"[]{}", "\\|=+-_`~", "!@#$%^&*()"}},
	{"/", []string{}},
	{"Foo/Bar", []string{}},
	{"/Foo/Bar/", []string{}},
	{"", []string{}},
	{"/Foo//Bar", []string{}},
	{"/世/界/Bar", []string{}},
	{"/./Bar", []string{}},
	{"/Bar/.", []string{}},
	{"/Bar/..", []string{}},
	{"/../Bar", []string{}},
	{"/.../Bar", []string{"...", "Bar"}},
}

func (s *UtilSuite) TestNameToSegments(c *check.C) {
	for _, entry := range namesTable {
		c.Check(document.NameToSegments(entry.Name), check.DeepEquals, entry.Segments, check.Commentf("Name: %#v", entry.Name))
	}
}

var segmentsTable = []struct {
	Segments []string
	Name     string
}{
	{[]string{"Foo", "Bar", "Baz"}, "/Foo/Bar/Baz"},
	{[]string{"base", "o(n)", "10%", "val Foo"}, "/base/o(n)/10%/val Foo"},
	{[]string{"<>,.?;:'\"[]{}", "\\|=+-_`~", "!@#$%^&*()"}, "/<>,.?;:'\"[]{}/\\|=+-_`~/!@#$%^&*()"},
	{[]string{}, ""},
	{[]string{"Bar", ""}, ""},
	{[]string{"世", "界", "Bar"}, ""},
	{[]string{"either/or", "one/theother", "foo"}, ""},
	{[]string{".", "foo"}, ""},
	{[]string{"..", "foo"}, ""},
	{[]string{"...", "foo"}, "/.../foo"},
}

func (s *UtilSuite) TestSegmentsToName(c *check.C) {
	for _, entry := range segmentsTable {
		c.Check(document.SegmentsToName(entry.Segments), check.DeepEquals, entry.Name, check.Commentf("Segments: %#v", entry.Segments))
	}
}
