package goblex

// RuneEOF is a rune representing the end of the input
const RuneEOF rune = 0

const (
	// StringEOF is a string "token" representing th end of the input
	StringEOF = "EOF"
)

var (
	// WhiteSpace is a convience slice of known whitespace characters. This can be used in a lexer to
	// test for any/all whitespace as a set of tokens.
	WhiteSpace = []string{string('\t'), string('\n'), string('\v'), string('\f'), string('\r'), string(' ')}
)

// TokenType is the type used by the Emit method to emit tokens. Implementors should create
// their own types to emit when building a lexer.
type TokenType int8

const (
	// TokenTypeError is a TokenType that can be used to emit errors
	TokenTypeError TokenType = -2

	// TokenTypeEOF is a TokenType that can be used to emit the end of the imput
	TokenTypeEOF = -1
)

// Token is the type that gets emitted by the Emit method.
// This is also the interface used when emitting custom tokens via the EmitToken method
type Token interface {
	// Type returns the TokenType of the emitted token
	Type() TokenType

	// String returns the string value of the emitted token.
	String() string
}

type defaultToken struct {
	tokenType TokenType
	value     string
}

func (t defaultToken) Type() TokenType {
	return t.tokenType
}

func (t defaultToken) String() string {
	return t.value
}
