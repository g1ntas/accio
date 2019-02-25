package markup_test

import (
	"github.com/g1ntas/accio/markup"
	"testing"
)

func TestParser(t *testing.T) {
	var cases = []struct {
		s string
		ast *markup.Tag
		err string
	}{
		// Empty tag
		{
			s: "tag",
			ast: &markup.Tag{Name: "tag"},
		},

		// Tag name with numbers
		{
			s: "tag",
			ast: &markup.Tag{Name: "tag"},
		},
	}
}