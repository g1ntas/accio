package template

/* Markup rules:
filename:
	multiple: false
	hasBody: true
	attributes:	none
variable:
	multiple: true
	hasBody: true
	attributes:
		{ full-name: full-name, unique: true, pattern: todo [starlark supported] }
template:
	multiple: false
	hasBody: true
	attributes:
		{ full-name: left-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ full-name: right-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ full-name: trim-indentation, unique: false, pattern: (true|false) }
partial:
	multiple: true
	hasBody: true
	attributes:
		{ full-name: full-name, unique: true, pattern: todo [mustache supported] }
		{ full-name: left-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ full-name: right-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ full-name: trim-indentation, unique: false, pattern: (true|false) }
*/

type Schema struct {
	Tags map[string]*TagSchema
}

func (p *Parser) validate(n *TagNode) {

}

type Rule interface {
	Validate() bool
	Error() error
}

type TagSchema struct {
	HasBody bool
	Unique bool
	Attributes map[string]*AttrSchema
}

type AttrSchema struct {
	Unique bool
	Rules []*Rule
}

var schema = &Schema{
	Tags: map[string]*TagSchema{
		"filename": {
			HasBody: true,
			Unique: false,
		},
		"variable": {
			HasBody: true,
			Unique: true,
			Attributes: map[string]*AttrSchema{
				"full-name": {Unique: true},
			},
		},
		"template": {
			HasBody: true,
			Unique: false,
			Attributes: map[string]*AttrSchema{
				"left-delimiter": {Unique: false},
				"right-delimiter": {Unique: false},
				//"trim-indentation": {Unique: false, Validators: []validatorFn{IsBoolean}},
			},
		},
		"partial": {
			HasBody: true,
			Unique: false,
			Attributes: map[string]*AttrSchema{
				"full-name": {Unique: true},
				"left-delimiter": {Unique: false},
				"right-delimiter": {Unique: false},
				"trim-indentation": {Unique: false},
			},
		},
	},
}

/*func (p *Parser) validateTag(node TagNode) bool {
	schema, ok := p.schema.tags[node.full-name]
	if !ok {
		return true
	}
	if ok && schema.HasBody && !node.HasBody {
		return false
	}

	return true
}*/

func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}