package goblex_test

import (
	"fmt"
	"testing"

	"github.com/brainicorn/goblex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	unicornInputEOF   = "I love #unicorns"
	unicornInputSpace = "I love #unicorns yup"
	unicornInputTab   = "I love #unicorns	yup"
	unicornInputNL    = `I love #unicorns
	yup`
	unicornInputComment = `//
	// I love #unicorns
	// yup`

	ignoreInput        = "I love *unicorns!"
	ignoreCommentInput = `
	/*
		I love
	*/

	//unicorns!`

	templateInput        = "this has a {{template}} of ${.somesort} in it!  "
	templateInputComment = "this has a {{//template}} of ${.somesort //} in it!  "
	smallTemplateInput   = "a {{template}}"
)

const (
	basicTokenType goblex.TokenType = iota
)

type GoblexTestSuite struct {
	suite.Suite
}

func TestGoblexSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(GoblexTestSuite))
}

func (suite *GoblexTestSuite) SetupSuite() {

}

func (suite *GoblexTestSuite) TestReadUntilOneOfNoToken() {
	suite.T().Parallel()

	tkn := "doh!"

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		tkn = lexer.CaptureUntilOneOf(false)
		return nil
	}

	l := goblex.NewLexer("simple", "some text", lexFun)
	l.Run()

	assert.Equal(suite.T(), "", tkn)
}

func (suite *GoblexTestSuite) TestReadUntilOneOfBlankToken() {
	suite.T().Parallel()

	tkn := "doh!"

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		tkn = lexer.CaptureUntilOneOf(false, "")
		return nil
	}

	l := goblex.NewLexer("simple", "some text", lexFun)
	l.Run()

	assert.Equal(suite.T(), "", tkn)
}

func (suite *GoblexTestSuite) TestReadUntilBlankToken() {
	suite.T().Parallel()

	tkn := true

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		tkn = lexer.CaptureUntil(false, "")
		return nil
	}

	l := goblex.NewLexer("simple", "some text", lexFun)
	l.Run()

	assert.False(suite.T(), tkn)
}

func (suite *GoblexTestSuite) TestReadUntilDebugLogging() {
	suite.T().Parallel()

	tkn := true

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		tkn = lexer.CaptureUntil(false, "x")
		return nil
	}

	l := goblex.NewLexer("simple", "some text", lexFun)
	l.Debug = true
	l.Run()

	assert.True(suite.T(), tkn)
}

func (suite *GoblexTestSuite) TestAddIgnoreToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(false, "!")
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", ignoreInput, lexFun)
	l.AddIgnoreTokens("*")
	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "I love unicorns"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but go '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestRemoveIgnoreToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.RemoveIgnoreTokens("*")
		lexer.CaptureUntil(false, "!")
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", ignoreInput, lexFun)
	l.AddIgnoreTokens("*")
	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "I love *unicorns"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but go '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestIgnoreCommentsToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "!")
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", ignoreCommentInput, lexFun)
	l.AddIgnoreTokens("/*", "*/", "//")
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "Iloveunicorns"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestConsumeMultiRuneToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.ConsumeCurrentToken(true)
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "{{"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestConsumeTokenSkipIgnores() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.ConsumeCurrentToken(true)
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInputComment, lexFun)
	l.AddIgnoreTokens("//")
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "{{"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestConsumeBlankToken() {
	suite.T().Parallel()

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)

	assert.False(suite.T(), l.ConsumeCurrentToken(false))
}

func (suite *GoblexTestSuite) TestSkipMultiRuneToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := "doh!"

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.SkipCurrentToken(true)
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := ""
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestSkipBlankToken() {
	suite.T().Parallel()

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)

	assert.False(suite.T(), l.SkipCurrentToken(false))
}

func (suite *GoblexTestSuite) TestReadIdent() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.SkipCurrentToken(true)
		lexer.CaptureIdent()
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "template"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestReadIdentSkipIgnores() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "$")
		lexer.SkipCurrentToken(true)
		lexer.AddIgnoreTokens("{.", "//")
		lexer.CaptureIdent()
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInputComment, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "somesort"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestReadIdentEOF() {
	suite.T().Parallel()

	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "!")
		lexer.SkipCurrentToken(true)
		lexer.CaptureIdent()
		lexer.Emit(basicTokenType)
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	l.Debug = true
	for {
		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			fmt.Println("got token")
			tkn = token.String()
			break
		case goblex.TokenTypeEOF:
			fmt.Println("got EOF")
			tkn = "EOF"
			break
		}

		if tkn == "EOF" {
			fmt.Println("EOF")
			break
		}
	}
	expected := "EOF"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestNextTokenIs() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.SkipCurrentToken(true)
		lexer.CaptureIdent()
		if lexer.CurrentTokenIs("}}") {
			lexer.Emit(basicTokenType)
		}

		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "template"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestNextTokenIsSkipIgnores() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "$")
		lexer.SkipCurrentToken(true)
		if lexer.CurrentTokenIs(".") {
			lexer.CaptureUntil(true, ".")
			lexer.SkipCurrentToken(true)
			lexer.CaptureIdent()
			lexer.Emit(basicTokenType)
		}

		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	l.AddIgnoreTokens("{")
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "somesort"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestNextTokenIsEOF() {
	suite.T().Parallel()

	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "!")
		lexer.SkipCurrentToken(true)
		if lexer.CurrentTokenIs("}}") {
			lexer.Emit(basicTokenType)
		}

		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	l.Debug = true
	for {
		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			fmt.Println("got token")
			tkn = token.String()
			break
		case goblex.TokenTypeEOF:
			fmt.Println("got EOF")
			tkn = "EOF"
			break
		}

		if tkn == "EOF" {
			fmt.Println("EOF")
			break
		}
	}
	expected := "EOF"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestNextTokenIsOneOfBlankToken() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.SkipCurrentToken(true)
		lexer.CaptureIdent()
		if _, tkn := lexer.CurrentTokenIsOneOf("", "}}"); tkn != "" {
			lexer.Emit(basicTokenType)
		}

		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	expected := "template"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestNextTokenIsOneOfPastEOF() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.CaptureUntil(true, "{{")
		lexer.SkipCurrentToken(true)
		lexer.CaptureIdent()
		lexer.CaptureUntil(true, "}}")
		if found, _ := lexer.CurrentTokenIsOneOf("}}}"); !found {
			tkn = "toofar"
		}

		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			tkn = token.String()
		}
	}
	assert.Equal(suite.T(), "toofar", tkn, "expected '%s' but got '%s'", "toofar", tkn)
}

func (suite *GoblexTestSuite) TestEmitError() {
	suite.T().Parallel()
	var token goblex.Token
	tkn := ""

	lexFun := func(lexer *goblex.Lexer) goblex.LexFn {
		lexer.Errorf("error %s", "yup")
		return nil
	}

	l := goblex.NewLexer("simple", templateInput, lexFun)
	for {
		if l.IsEOF() {
			fmt.Println("EOF")
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case goblex.TokenTypeError:
			tkn = token.String()
		}
	}
	expected := "error yup"
	assert.Equal(suite.T(), expected, tkn, "expected '%s' but got '%s'", expected, tkn)
}

func (suite *GoblexTestSuite) TestHashtagEOF() {
	suite.T().Parallel()

	suite.LexHashtag(unicornInputEOF, hashtagEOF)
}

func (suite *GoblexTestSuite) TestHashtagSpace() {
	suite.T().Parallel()

	suite.LexHashtag(unicornInputSpace, hashtagWS)
}

func (suite *GoblexTestSuite) TestHashtagTab() {
	suite.T().Parallel()

	suite.LexHashtag(unicornInputTab, hashtagWS)
}

func (suite *GoblexTestSuite) TestHashtagNL() {
	suite.T().Parallel()

	suite.LexHashtag(unicornInputNL, hashtagWS)
}

func (suite *GoblexTestSuite) TestHashtagComments() {
	suite.T().Parallel()

	suite.lexHashtagCommentFlag(unicornInputComment, hashtagComments, true)
}

func (suite *GoblexTestSuite) LexHashtag(input string, lexFun goblex.LexFn) {
	suite.lexHashtagCommentFlag(input, lexFun, false)

}

func (suite *GoblexTestSuite) lexHashtagCommentFlag(input string, lexFun goblex.LexFn, ignoreComments bool) {
	var token goblex.Token

	l := goblex.NewLexer("simple", input, lexFun)

	if ignoreComments {
		l.AddIgnoreTokens("/*", "*/", "//")
	}

	hashtag := "nada"

	for {
		if l.IsEOF() {
			break
		}

		token = l.NextEmittedToken()
		switch token.Type() {
		case basicTokenType:
			hashtag = token.String()
		}
	}

	assert.Equal(suite.T(), "#unicorns", hashtag, "expected #unicorns but was %s", hashtag)

}

func hashtagComments(lexer *goblex.Lexer) goblex.LexFn {
	if lexer.CaptureUntil(true, "#") {
		lexer.ConsumeCurrentToken(true)
		lexer.CaptureUntilOneOf(true, "yup")
		lexer.Emit(basicTokenType)
	}

	return nil
}

func hashtagEOF(lexer *goblex.Lexer) goblex.LexFn {
	return hashtaglex(lexer, []string{"EOF"})
}

func hashtagWS(lexer *goblex.Lexer) goblex.LexFn {
	return hashtaglex(lexer, goblex.WhiteSpace)
}

func hashtaglex(lexer *goblex.Lexer, end []string) goblex.LexFn {

	if lexer.CaptureUntil(true, "#") {
		lexer.ConsumeCurrentToken(true)
		lexer.CaptureUntilOneOf(false, end...)
		lexer.Emit(basicTokenType)
	}

	return nil
}
