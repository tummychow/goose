package document

import (
	"strings"
	"unicode"
)

// ValidSegmentChars is a unicode.RangeTable containing all the characters that
// are valid in a name segment.
var ValidSegmentChars = &unicode.RangeTable{
	R16: []unicode.Range16{
		{
			Lo:     0x20,
			Hi:     0x2E,
			Stride: 1,
		},
		{
			Lo:     0x30,
			Hi:     0x7E,
			Stride: 1,
		},
	},
}

// NameToSegments takes a valid Document Name, and splits it into a series of
// valid name segments. If the argument is not a valid name, the empty slice
// will be returned.
func NameToSegments(name string) []string {
	if len(name) == 0 {
		return []string{}
	}
	if name[0] != '/' {
		return []string{}
	}
	if name[len(name)-1] == '/' {
		return []string{}
	}
	if strings.Index(name, "//") != -1 {
		return []string{}
	}
	if strings.IndexFunc(name, func(c rune) bool {
		return !unicode.Is(ValidSegmentChars, c) && c != '/'
	}) != -1 {
		return []string{}
	}

	return strings.Split(name, "/")[1:]
}
