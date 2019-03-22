package markup

import (
	"fmt"
)

var tokenName = map[tokenType]string {
	tokenError:       "error",
	tokenEOF:         "EOF",
	tokenSpace:       "space",
	tokenNewline:     "newline",
	tokenIdentifier:  "identifier",
	tokenLeftDelim:   "left delimiter",
	tokenRightDelim:  "right delimiter",
	tokenString:      "string",
	tokenBody:        "body",
	tokenAttrDeclare: "-",
	tokenAssign:      "=",
}

func (t tokenType) String() string {
	s := tokenName[t]
	if s == "" {
		return fmt.Sprintf("item%d", int(t))
	}
	return s
}

type lexTest struct {
	name string
	input string
	tokens []token
}

func mkItem(typ tokenType, text string) token {
	return token{
		typ: typ,
		val: text,
	}
}

var (
	tEOF = mkItem(tokenEOF, "")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEOF}},
}

// collect gathers the emitted tokens into a slice
func collect(t *lexTest, left, right string) (tokens []token) {
	lx := lex(t.name, t.input, left, right)
	for {
		token := lx.nextToken()
		tokens = append(tokens, token)
		if token.typ == tokenEOF || token.typ == tokenError {
			break
		}
	}
	return
}

/*func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens :=
	}
}*/