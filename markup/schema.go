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
	Validator validatorFn
}

var schema = &Schema{
	Tags: map[string]*TagSchema{
		"filename": &TagSchema{
			HasBody: true,
			Multiple: false,
		},
		"variable": &TagSchema{
			HasBody: true,
			Multiple: true,
			Attributes: map[string]*AttrSchema{
				"name": &AttrSchema{Unique: true},
			},
		},
		"template": &TagSchema{
			HasBody: true,
			Multiple: false,
			Attributes: map[string]*AttrSchema{
				"left-delimiter": &AttrSchema{Unique: false},
				"right-delimiter": &AttrSchema{Unique: false},
				"trim-indentation": &AttrSchema{Unique: false, Validator: IsBoolean},
			},
		},
		"partial": &TagSchema{
			HasBody: true,
			Multiple: false,
			Attributes: map[string]*AttrSchema{
				"name": &AttrSchema{Unique: true},
				"left-delimiter": &AttrSchema{Unique: false},
				"right-delimiter": &AttrSchema{Unique: false},
				"trim-indentation": &AttrSchema{Unique: false},
			},
		},
	},
}

func IsBoolean(value string) bool {
	return value == "true" || value == "false"
}