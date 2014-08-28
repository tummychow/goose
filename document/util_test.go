package document_test

import (
	"github.com/tummychow/goose/document"
	"testing"
)

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
}

func TestNameToSegments(t *testing.T) {
	for _, entry := range namesTable {
		actual := document.NameToSegments(entry.Name)
		equality, _ := CompareSlices(actual, entry.Segments)

		if !equality {
			t.Errorf("Name %#v produced segments %#v (expected %#v)", entry.Name, actual, entry.Segments)
		}
	}
}

// CompareSlices compares two slices of strings for equality. The bool return
// indicates whether the arguments are equal. The int return indicates where
// the slices stopped being equal (if they were equal, this return value is
// undefined). If the slices had different lengths, the int return will be -1.
//
// If the two slices were equal length, and all elements were pairwise equal:
// return (true, <any integer>)
//
// If the two slices have unequal lengths: return (false, -1)
//
// If the two slices were equal length, and all elements were pairwise equal up
// to and excluding the ith index: return (false, i)
func CompareSlices(a, b []string) (bool, int) {
	if len(a) != len(b) {
		return false, -1
	}
	for i := range a {
		if a[i] != b[i] {
			return false, i
		}
	}

	return true, -1
}
