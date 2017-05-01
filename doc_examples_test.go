package goblex

import "fmt"

func Example() {
	input := "I like #unicorns and #cheese"
	var tagType TokenType = 1
	var lexTagname LexFn
	var lexPound LexFn
	var token Token

	lexPound = func(l *Lexer) LexFn {
		if l.CaptureUntil(true, "#") {
			l.ConsumeCurrentToken(true)
			return lexTagname
		}

		return nil
	}

	lexTagname = func(l *Lexer) LexFn {
		if l.CaptureIdent() {
			l.Emit(tagType)
		}

		return lexPound
	}

	l := NewLexer("myLexer", input, lexPound)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case tagType:
			fmt.Print(token.String() + ",")
		}
	}

	// Output: #unicorns,#cheese,

}
func ExampleTokenType() {
	const (
		DollarSignTokenType TokenType = iota
		AtSymbolTokenType
	)
}

func ExampleNewLexer() {
	fn := func(l *Lexer) LexFn { return nil }

	l := NewLexer("myLexer", "some text", fn)
	fmt.Println(l.Name)
}

func ExampleLexer_AddIgnoreTokens() {
	fn := func(l *Lexer) LexFn { return nil }

	l := NewLexer("myLexer", "some // text", fn)

	// when parsing, ignore all comment tokens
	l.AddIgnoreTokens("//", "/*", "*/")

}

func ExampleLexer_RemoveIgnoreTokens() {
	fn := func(l *Lexer) LexFn {
		// when this particular lexer runs, we want to see single line comment tokens
		l.RemoveIgnoreTokens("//")
		return nil
	}

	l := NewLexer("myLexer", "some // text", fn)

	// when parsing, ignore all comment tokens
	l.AddIgnoreTokens("//", "/*", "*/")

}

func ExampleLexer_NextEmittedToken() {
	var hashtag TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		l.ConsumeCurrentToken(true)
		l.CaptureIdent()
		l.Emit(hashtag)
		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case hashtag:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: #text
}

func ExampleLexer_CaptureUntil_skipwhitespace() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		if l.CaptureUntil(true, "!") {
			l.Emit(chars)
		}
		return nil
	}

	l := NewLexer("myLexer", "some text!", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: sometext
}

func ExampleLexer_CaptureUntil_includewhitespace() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		if l.CaptureUntil(false, "!") {
			l.Emit(chars)
		}
		return nil
	}

	l := NewLexer("myLexer", "some text!", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some text
}

func ExampleLexer_CaptureUntilOneOf() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		if tkn := l.CaptureUntilOneOf(true, "#", "!"); tkn != "" {
			l.Emit(chars)
		}
		return nil
	}

	l := NewLexer("myLexer", "some #text!", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some
}

func ExampleLexer_CaptureIdent() {
	var ident TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		if l.CaptureIdent() {
			l.Emit(ident)
		}
		return nil
	}

	l := NewLexer("myLexer", "some text!", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case ident:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some
}

func ExampleLexer_ConsumeCurrentToken_newbuffer() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		if l.ConsumeCurrentToken(true) {
			l.Emit(chars)
		}

		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: #
}

func ExampleLexer_ConsumeCurrentToken_oldbuffer() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		if l.ConsumeCurrentToken(false) {
			l.Emit(chars)
		}

		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some#
}

func ExampleLexer_SkipCurrentToken_newbuffer() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		if l.SkipCurrentToken(true) {
			l.Emit(chars)
		}

		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output:
}

func ExampleLexer_SkipCurrentToken_oldbuffer() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		if l.SkipCurrentToken(false) {
			l.Emit(chars)
		}

		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some
}

func ExampleLexer_CurrentTokenIs() {
	var chars TokenType = 1
	var token Token
	var tokenValue = ""

	fn := func(l *Lexer) LexFn {
		l.CaptureUntil(true, "#")
		if l.CurrentTokenIs("#") {
			l.Emit(chars)
		}

		return nil
	}

	l := NewLexer("myLexer", "some #text", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case chars:
			tokenValue = token.String()
		}
	}

	fmt.Println(tokenValue)
	// Output: some
}

var sliceType TokenType = 100

type sliceToken struct {
	tokenType TokenType
	slice     []string
}

func (s sliceToken) Type() TokenType {
	return s.tokenType
}

func (s sliceToken) String() string {
	return fmt.Sprintf("%+v", s.slice)
}

func (s sliceToken) Slice() []string {
	return s.slice
}

func ExampleLexer_EmitToken() {

	var token Token
	var tokenValue = ""

	customToken := sliceToken{tokenType: sliceType, slice: make([]string, 0)}

	fn := func(l *Lexer) LexFn {
		if l.CaptureIdent() {
			customToken.slice = append(customToken.slice, l.Flush())
		}

		if l.CaptureIdent() {
			customToken.slice = append(customToken.slice, l.Flush())
			l.EmitToken(customToken)
		}
		return nil
	}

	l := NewLexer("myLexer", "some text!", fn)

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case sliceType:
			ct := token.(sliceToken)
			for _, v := range ct.Slice() {
				tokenValue = tokenValue + v
			}
		}
	}

	fmt.Println(tokenValue)
	// Output: sometext
}
