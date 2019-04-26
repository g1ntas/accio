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
		name: name
		pattern: ...

*/

type rule struct {
	tag string
	attr string
	message string
}