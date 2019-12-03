package template

import (
	"errors"
	"github.com/g1ntas/accio/markup"
	"log"
)

const (
	tagFilename = "filename"
	tagSkip     = "skipif"
	tagTemplate = "template"
	tagPartial  = "partial"
	tagVariable = "variable"
	attrName    = "name"
)

var errSkipTag = errors.New("skip tag")

type schema struct {
	filename string
	skip     string
	body     string
	vars     vars
	partials map[string]string
}

// vars holds variables in their original order
type vars [][2]string

func newMarkup() *schema {
	return &schema{
		vars:     make(vars, 0, 5),
		partials: make(map[string]string),
	}
}

func parse(p *markup.Parser) *schema {
	m := newMarkup()
	for _, tag := range p.Tags {
		switch tag.Name {
		case tagFilename:
			m.filename = parseScriptBody(tag)
		case tagSkip:
			m.skip = parseScriptBody(tag)
		case tagTemplate:
			m.body = parseBody(tag)
		case tagPartial, tagVariable:
			name, err := parseTagNameAttr(tag)
			if err == errSkipTag {
				continue
			}
			if tag.Body == nil {
				log.Printf("Tag %q with name %q has no content. Skipping...\n", tag.Name, name)
				continue
			}
			switch tag.Name {
			case tagPartial:
				if _, ok := m.partials[name]; ok {
					log.Printf("Partial definition for %q already exists. Overwriting...", name)
				}
				m.partials[name] = parseBody(tag)
			case tagVariable:
				m.vars = append(m.vars, [2]string{name, parseScriptBody(tag)})
			}
		}
	}
	return m
}

func parseTagNameAttr(tag *markup.TagNode) (string, error) {
	var name string
	for _, attr := range tag.Attributes {
		if attr.Name == attrName {
			name = attr.Value
		}
	}
	if len(name) == 0 {
		log.Printf("Tag %q is missing %q attribute. Skipping...", tag.Name, attrName)
		return "", errSkipTag
	}
	return name, nil
}

func parseBody(tag *markup.TagNode) string {
	if tag.Body == nil {
		return ""
	}
	return tag.Body.Content
}

func parseScriptBody(tag *markup.TagNode) string {
	if tag.Body == nil {
		return ""
	}
	if tag.Body.Inline {
		return wrapInlineScript(tag.Body.Content)
	}
	return tag.Body.Content
}
