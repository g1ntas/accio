package template

/*
import (
	"errors"
	"fmt"
	"github.com/g1ntas/accio/markup"
	"fs/ioutil"
)

type Markup struct {
	Filename *script
	Skip *script
	Body template
	Vars map[string]*Var
	Partials map[string]*Partial

}

type Var struct {
	Name string
	Value *script
}

type Partial struct {
	Name string
	Body string
}

func ParseFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	p, err := markup.Parse(path, string(b), "", "")
	if err != nil {
		return err
	}
	parse(p)
}

func parseBody(tag *markup.TagNode) (*Var, error) {

}

func parsePartial(tag *markup.TagNode) (*Partial, error) {

}

func parseVariable(tag *markup.TagNode) (*Var, error) {
	v := &Var{}
	for _, attr := range tag.Attributes {
		switch attr.Name {
		case "name":
			v.Name = attr.Value
		default:
			return nil, fmt.Errorf("unknown attribute %q of variable tag", attr.Name)
		}
	}
	if len(v.Name) == 0 {
		return nil, errors.New("missing \"name\" attribute for variable tag")
	}
	if tag.Body == nil {
		return nil, errors.New("missing body for variable tag")
	}
	// todo: compile starlark
	v.Value = tag.Body.Content
	return v, nil
}

/*func parseFilename(tag *markup.TagNode) (string, error) {
	if tag.Body == nil {
		return "", errors.New("missing body for filename markup")
	}
	return tag.Body.Content, nil
}*/
/*
func parse(p *markup.Parser) (*Markup, error) {
	mrkp := &Markup{}
	for _, tag := range p.Tags {
		var err error
		switch tag.Name {
		// accio specific
		case "filename":
			mrkp.Filename, err = parseFilename(tag)
		case "skipif":
		// generic
		case "variable":
			v, err := parseVariable(tag)
			if err != nil {
				return nil, err
			}
			if _, ok := mrkp.Vars[v.Name]; ok {
				return nil, fmt.Errorf("definition of variable %q already exists within this template", v.Name)
			}
		case "template":

		case "partial":
		default:
		}
		if err != nil {
			return nil, err
		}
	}
	return mrkp, nil
}
*/