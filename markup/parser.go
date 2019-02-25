package markup

type Attribute struct {
	Name string
	Value interface{}
}

type Tag struct {
	Attributes map[string]Attribute
	Body string
	Name string
}