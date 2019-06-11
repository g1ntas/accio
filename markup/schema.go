package markup

/* Markup rules:
filename:
	multiple: false
	hasBody: true
	attributes:	none
variable:
	multiple: true
	hasBody: true
	attributes:
		{ name: name, unique: true, pattern: todo [starlark supported] }
template:
	multiple: false
	hasBody: true
	attributes:
		{ name: left-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ name: right-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ name: trim-indentation, unique: false, pattern: (true|false) }
partial:
	multiple: true
	hasBody: true
	attributes:
		{ name: name, unique: true, pattern: todo [mustache supported] }
		{ name: left-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ name: right-delimiter, unique: false, pattern: todo: [mustache supported] }
		{ name: trim-indentation, unique: false, pattern: (true|false) }
*/

type validatorFn func(string) bool

type Schema struct {
	Tags map[string]*TagSchema
}

type TagSchema struct {
	HasBody bool
	Multiple bool
	Attributes map[string]*AttrSchema
}

type AttrSchema struct {
	Unique bool
	Validators []validatorFn
}

var schema = &Schema{
	Tags: map[string]*TagSchema{
		"filename": {
			HasBody: true,
			Multiple: false,
		},
		"variable": {
			HasBody: true,
			Multiple: true,
			Attributes: map[string]*AttrSchema{
				"name": {Unique: true},
			},
		},
		"template": {
			HasBody: true,
			Multiple: false,
			Attributes: map[string]*AttrSchema{
				"left-delimiter": {Unique: false},
				"right-delimiter": {Unique: false},
				"trim-indentation": {Unique: false, Validators: []validatorFn{IsBoolean}},
			},
		},
		"partial": {
			HasBody: true,
			Multiple: false,
			Attributes: map[string]*AttrSchema{
				"name": {Unique: true},
				"left-delimiter": {Unique: false},
				"right-delimiter": {Unique: false},
				"trim-indentation": {Unique: false},
			},
		},
	},
}

/*func (p *Parser) validateTag(node TagNode) bool {
	schema, ok := p.schema.tags[node.name]
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