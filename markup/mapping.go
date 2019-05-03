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

type validatorFn func() bool

type tagRule struct {
	name string
	hasBody bool
	multiple bool
}

type attrRule struct {
	name string
	unique bool
	validator validatorFn
}