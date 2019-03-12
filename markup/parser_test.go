package markup_test

import (
	"github.com/g1ntas/accio/markup"
	"testing"
)

func TestParser(t *testing.T) {
	var cases = []struct {
		s   string
		ast []*markup.Tag
		err string
	}{
		// Tag naming
		{
			s:   "tag",
			ast: []*markup.Tag{{Name: "tag"}},
		},
		{
			s:   "tag123",
			ast: []*markup.Tag{{Name: "tag123"}},
		},
		{
			s:   "tag_123",
			ast: []*markup.Tag{{Name: "tag_123"}},
		},
		{
			s:   "tag-123",
			ast: []*markup.Tag{{Name: "tag-123"}},
		},

		// Attribute data types
		{
			s: "tag -string=\"double quotes\"",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"string": "double quotes",
			}}},
		},
		{
			s: "tag -string='single quotes'",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"string": "single quotes",
			}}},
		},
		{
			s: "tag -integer=10",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"integer": 10,
			}}},
		},
		{
			s: "tag -float=10.0",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"float": 10.0,
			}}},
		},
		{
			s: "tag -bool=true",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"true": true,
			}}},
		},
		{
			s: "tag -bool=false",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"bool": false,
			}}},
		},
		{
			s: "tag -flag",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"flag": true,
			}}},
		},

		// Multiple attributes
		{
			s: "tag -arg1 -arg2='test' -arg3=1",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"arg1": true,
				"arg2": "test",
				"arg3": 1,
			}}},
		},
		{
			s: "tag -arg1 -arg2='test'",
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"arg1": true,
				"arg2": "test",
				"arg3": 1,
			}}},
		},

		// Whitespace around attributes
		{
			s: `tag -arg1  -arg2
					-arg3  -arg4
					-arg5  -arg6`,
			ast: []*markup.Tag{{Name: "tag", Attributes: map[string]interface{}{
				"arg1": true,
				"arg2": true,
				"arg3": true,
				"arg4": true,
				"arg5": true,
				"arg6": true,
			}}},
		},

		// Body
		{
			s: "inline-tag <<test body>>",
			ast: []*markup.Tag{{
				Name: "inline-tag",
				Body: "test body",
			}},
		},
		{
			s: "inline-tag -arg1 <<test body>>",
			ast: []*markup.Tag{{
				Name:       "inline-tag",
				Body:       "test body",
				Attributes: map[string]interface{}{"arg1": true},
			}},
		},
		{
			s: "empty-tag <<>>",
			ast: []*markup.Tag{{
				Name: "empty-tag",
				Body: "",
			}},
		},
		{
			s: `
multiline-tag <<
	line1
	line2
>> `,
			ast: []*markup.Tag{{
				Name: "empty-tag",
				Body: "\tline1\n\tline2",
			}},
		},

		// Multiple tags
		{
			s: "tag1 -index=1\ntag2 -index=2",
			ast: []*markup.Tag{
				{Name: "tag1", Attributes: map[string]interface{}{"index": 1}},
				{Name: "tag2", Attributes: map[string]interface{}{"index": 2}},
			},
		},

		// Changing body delimiters
		{
			s: `
aml -start="{{" -end="}}"
tag1 -attr="test" {{inline body}}
tag2 {{
multiline
body
}}
`,
			ast: []*markup.Tag{
				{Name: "aml", Attributes: map[string]interface{}{"start": "{{", "end": "}}"}},
				{Name: "tag1", Attributes: map[string]interface{}{"attr": "test"}, Body: "inline body"},
				{Name: "tag2", Body: "multiline\nbody"},
			},
		},

		// Comments are not parsed
		{
			s: "--- Comments are ignored\ntag1",
			ast: []*markup.Tag{
				{Name: "tag1"},
			},
		},

		// Errors
		// todo: body starting delimiter can not contain any other whitespaces than single space before it
		// todo: multiline body ending delimiter can not contain any symbol before it except new line
		// todo: special aml tag can only be defined as first tag
		// todo: special aml tag delimiters can not contain any whitespace characters
		// todo: special aml tag delimiters can not contain any quote characters
		// todo: special aml tag delimiters can not be equal to triple dashes (comments)
		// todo: tag names can not contain any special characters than dashes or underscores
		// todo: tags can not contain any whitespace characters before definition, except newline
		// todo:
	}
}
