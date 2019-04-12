package markup

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all it's components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
	// tree returns the containing *Tree.
	// It is unexported so all implementations of Node are in this package.
	tree() *Tree
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (nt NodeType) Type() NodeType {
	return nt
}

const (
	NodeTag        NodeType = iota
	NodeAttr
	NodeString
	NodeBody
)