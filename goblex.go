package goblex

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// LexFn is a function that can be run by the Lexer.
//
// Every LexFn can use the provided lexer to parse the lexers input from the current state and can
// return the next LexFn in the chain. Returning nil signals the lexer to consume the rest of the
// input until the end is reached emitting no tokens.
type LexFn func(lexer *Lexer) LexFn

// Lexer is the main object used to lex some input and emit tokens
type Lexer struct {
	// Name is a string used to identify this lexer for debugging purposes
	Name string
	// Debug is a flag that when set to true outputs debug logs to the console
	Debug bool
	// AutoEatWhitespace is a flag to toggle discarding all *beginning* whitespace when capturing.
	// defaults to true
	AutoEatWhitespace bool
	ignoreTokens      map[string]bool
	inputBuffer       *bufio.Reader
	tokens            chan Token
	state             LexFn
	begin             LexFn
	tokenBuffer       bytes.Buffer
	currentRune       rune
	lastKnownToken    string
	runeCache         []rune
	logIndent         int
}

// NewLexer creates a new Lexer instance with the given name and set input as the text to parse using
// the begin LexFn as the entry point when parsing.
func NewLexer(name, input string, begin LexFn) *Lexer {
	l := &Lexer{
		Name:              name,
		Debug:             false,
		AutoEatWhitespace: true,
		ignoreTokens:      make(map[string]bool),
		inputBuffer:       bufio.NewReader(strings.NewReader(input)),
		state:             begin,
		begin:             begin,
		tokens:            make(chan Token, 3),
		logIndent:         0,
	}

	l.read()
	return l
}

// AddIgnoreTokens adds the list of tokens to be ignored when capturing tokens to be emitted.
// This can be called at anytime during lexing to ignore certain tokens from being captured.
func (lxr *Lexer) AddIgnoreTokens(tokens ...string) {
	for _, tkn := range tokens {
		if strings.TrimSpace(tkn) != "" {
			lxr.ignoreTokens[tkn] = true
		}
	}
}

// RemoveIgnoreTokens removes the list of tokens from the ignore list previously added with
// AddIgnoreTokens.
//
// This can be called at anytime during lexing.
func (lxr *Lexer) RemoveIgnoreTokens(tokens ...string) {
	for _, tkn := range tokens {
		if strings.TrimSpace(tkn) != "" {
			lxr.ignoreTokens[tkn] = false
		}
	}
}

// Run will start the lexing process and recusively call the LexFn functions in the chain until the
// end of the input is reached.
//
// This function can be used for testing one-off lex functions or simple chains. This function should
// not be used for building parser and instead consumers should use the NextEmittedToken function to
// start/control processing of input
func (lxr *Lexer) Run() {
	for state := lxr.begin; state != nil; {
		state = state(lxr)
	}

	lxr.shutdown()
}

// NextEmittedToken returns the next Token that has been emitted by the lexer.
//
// This is the main function parsers should use in a loop until the end of input is reached.
func (lxr *Lexer) NextEmittedToken() Token {
	lxr.enterDebug("NextEmittedToken")
	for {
		select {
		case token := <-lxr.tokens:
			lxr.logDebug("sending token %+v", token)
			lxr.exitDebug("NextEmittedToken")
			return token
		default:
			if lxr.state != nil {
				lxr.state = lxr.state(lxr)
			} else {
				ioutil.ReadAll(lxr.inputBuffer)
				lxr.currentRune = RuneEOF
				lxr.logDebug("sending tokenEOF")
				lxr.exitDebug("NextEmittedToken")
				return defaultToken{tokenType: TokenTypeEOF, value: StringEOF}
			}
		}
	}

}

// Emit creates a new Token of type tokeType whose value is the value of the current capture buffer.
// The Token is emitted and a new capture buffer is started.
func (lxr *Lexer) Emit(tokenType TokenType) {
	lxr.enterDebug("Emit")
	lxr.logDebug("emitting token %s", lxr.tokenBuffer.String())
	lxr.tokens <- defaultToken{tokenType: tokenType, value: lxr.tokenBuffer.String()}
	lxr.tokenBuffer.Reset()
	lxr.exitDebug("Emit")
}

// EmitToken emits the provided token but does not clear the current capture buffer
// This can be sed to emit custom tokens during lexing without upsetting the parsing flow
func (lxr *Lexer) EmitToken(token Token) {
	lxr.enterDebug("EmitToken")
	lxr.tokens <- token
	lxr.exitDebug("EmitToken")
}

// Flush clears the current capture buffer and returns it's previously held value.
// This can be used to get values to build custom tokens to be emitted by EmitToken
func (lxr *Lexer) Flush() string {
	lxr.enterDebug("Flush")
	retVal := lxr.tokenBuffer.String()
	lxr.tokenBuffer.Reset()
	lxr.exitDebug("Flush")

	return retVal
}

// CaptureUntil reads characters from the input buffer and writes them to the capture buffer stopping
// when it reaches the until token and returns whether or not the until token was actually reached.
//
// If skipWitespace is true, no whitespace will be written to the capture buffer.
func (lxr *Lexer) CaptureUntil(skipWhitespace bool, until string) bool {
	lxr.enterDebug("find until %s", until)
	if until == "" {
		lxr.exitDebug("find until %s", until)
		return false
	}

	b := lxr.CaptureUntilOneOf(skipWhitespace, until) != ""
	lxr.exitDebug("find until %s", until)

	return b

}

// CaptureUntilOneOf does the same thing as CaptureUntil but accepts multiple until tokens to look for
// and returns the token that was found. If an until token was not found, the return value will be ""
//
// If skipWitespace is true, no whitespace will be written to the capture buffer.
func (lxr *Lexer) CaptureUntilOneOf(skipWhitespace bool, tokens ...string) string {
	lxr.enterDebug("ReadUntilOneOf")
	if len(tokens) < 1 || lxr.currentRune == RuneEOF {
		lxr.exitDebug("ReadUntilOneOf")
		return ""
	}

	lxr.logDebug("searching for tokens %q", tokens)
	foundToken := ""

	for {
		if skipWhitespace && lxr.EatWhitespace() {
			lxr.logDebug("skipped WS")
			continue
		}

		ch := lxr.currentRune
		lxr.logDebug("testing char %q", ch)
		if ch == RuneEOF {
			break
		}

		if lxr.skipIgnores() {
			lxr.logDebug("skipped Ignores")
			continue
		}

		for _, tkn := range tokens {
			if tkn == "" {
				lxr.logDebug("skipping blank token")
				continue
			}

			if lxr.CurrentTokenIs(tkn) {
				lxr.logDebug("found token '%s'", tkn)
				foundToken = tkn
				break
			}
		}

		if foundToken != "" {
			break
		}

		lxr.logDebug("writing to buffer %q", ch)
		lxr.tokenBuffer.WriteRune(ch)
		lxr.read()
	}

	lxr.exitDebug("ReadUntilOneOf")
	lxr.lastKnownToken = foundToken

	return foundToken
}

// CaptureIdent reads all valide IDENT characters from the input stream and writes them to the capture
// buffer stopping when a non-ident character is reached and returns whether an ident character was
// indeed captured.
func (lxr *Lexer) CaptureIdent() bool {
	lxr.enterDebug("ReadIdent")
	foundIdent := false
	if lxr.AutoEatWhitespace {
		lxr.EatWhitespace()
	}
	for {

		ch := lxr.currentRune
		lxr.logDebug("currentRune is: %q", lxr.currentRune)
		if ch == RuneEOF {
			lxr.logDebug("EOF, exiting")
			break
		}

		if lxr.skipIgnores() {
			if lxr.AutoEatWhitespace {
				lxr.EatWhitespace()
			}
			continue
		}

		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			lxr.logDebug("not an ident character, exiting")
			break
		}

		foundIdent = true
		lxr.logDebug("writing to buffer %q", ch)
		lxr.tokenBuffer.WriteRune(ch)
		lxr.read()

	}

	if lxr.AutoEatWhitespace {
		lxr.EatWhitespace()
	}

	if lxr.skipIgnores() && lxr.AutoEatWhitespace {
		lxr.EatWhitespace()
	}

	lxr.exitDebug("ReadIdent")
	return foundIdent
}

// ConsumeCurrentToken consumes the token found by a previous call to CaptureUntil or CaptureUntilOneOf
// and writes it to the capture buffer returning whether or not a token was indeed consumed.
//
// If clearPrevious is true the previous buffer will be discarded and the token will be written to a
// new buffer.
//
// If no previous token was found this method will return false without clearing the buffer.
func (lxr *Lexer) ConsumeCurrentToken(clearPrevious bool) bool {
	if lxr.lastKnownToken == "" || !lxr.CurrentTokenIs(lxr.lastKnownToken) || lxr.currentRune == RuneEOF {
		return false
	}

	if clearPrevious {
		lxr.tokenBuffer.Reset()
	}

	lxr.tokenBuffer.WriteRune(lxr.currentRune)
	numRunes := utf8.RuneCountInString(lxr.lastKnownToken) - 1
	for i := 0; i < numRunes; i++ {
		ch := lxr.read()
		lxr.tokenBuffer.WriteRune(ch)
	}
	lxr.read()

	if lxr.AutoEatWhitespace {
		lxr.EatWhitespace()
	}

	for lxr.skipIgnores() {
		if lxr.AutoEatWhitespace {
			lxr.EatWhitespace()
		}
	}
	return true
}

// SkipCurrentToken does the same thing as ConsumeCurrentToken but discards the token instead of
// writing it to the capture buffer.
//
// If clearPrevious is true the previous buffer will be discarded and a new buffer will be started.
//
// If no previous token was found this method will return false without clearing the buffer.
func (lxr *Lexer) SkipCurrentToken(clearPrevious bool) bool {
	lxr.enterDebug("Skip Current Token")
	lxr.logDebug("checking lastKnowToken %q", lxr.lastKnownToken)
	gotLastKnown := lxr.CurrentTokenIs(lxr.lastKnownToken)
	lxr.logDebug("got lastKnownToken? %t", gotLastKnown)
	lxr.logDebug("lastKnowToken %q", lxr.lastKnownToken)
	if lxr.lastKnownToken == "" || !gotLastKnown || lxr.currentRune == RuneEOF {
		lxr.logDebug("last known token not found, returning")
		return false
	}

	if clearPrevious {
		lxr.tokenBuffer.Reset()
	}

	numRunes := utf8.RuneCountInString(lxr.lastKnownToken)
	for i := 0; i < numRunes; i++ {
		lxr.read()
	}

	if lxr.AutoEatWhitespace {
		lxr.EatWhitespace()
	}

	for lxr.skipIgnores() {
		if lxr.AutoEatWhitespace {
			lxr.EatWhitespace()
		}
	}
	lxr.logDebug("skipped token, current rune is %q", lxr.currentRune)
	lxr.exitDebug("Skip Current Token")
	return true
}

// CurrentTokenIs returns whether the start of the current input stream buffer is on t.
func (lxr *Lexer) CurrentTokenIs(t string) bool {
	found, _ := lxr.CurrentTokenIsOneOf(t)
	return found
}

// CurrentTokenIsOneOf returns whether the start of the current input stream buffer is on t.
func (lxr *Lexer) CurrentTokenIsOneOf(tokens ...string) (bool, string) {
	lxr.enterDebug("CurrentTokenIsOneOf")
	found := ""

	if lxr.currentRune == RuneEOF {
		return false, ""
	}

	for _, tkn := range tokens {
		if tkn == "" {
			continue
		}

		lxr.logDebug("checking token '%s'", tkn)
		lxr.logDebug("current rune is %q", lxr.currentRune)
		bufRunes := []rune{lxr.currentRune}
		tokenRunes := []rune(tkn)

		if tokenRunes[0] != bufRunes[0] {
			continue
		}

		numPeeks := len(tokenRunes) - 1

		if numPeeks > 0 {
			peeks := lxr.peek(numPeeks)

			bufRunes = append(bufRunes, peeks...)
		}
		lxr.logDebug("token runes %q", tokenRunes)
		lxr.logDebug("buffer runes %q", bufRunes)
		if reflect.DeepEqual(tokenRunes, bufRunes) {
			lxr.logDebug("rune slices match!")
			found = tkn
			break
		}
	}

	lxr.logDebug("found token %q", found)
	lxr.exitDebug("CurrentTokenIsOneOf")
	return found != "", found

}

// Errorf formats a string using format and args and emits a Token with TokenTypeError as it's type and
// the formatted string as it's Value
func (lxr *Lexer) Errorf(format string, args ...interface{}) LexFn {
	lxr.tokens <- defaultToken{
		tokenType: TokenTypeError,
		value:     fmt.Sprintf(format, args...),
	}

	return nil
}

// IsEOF returns the true/false if the lexer is at the end of the input stream.
func (lxr *Lexer) IsEOF() bool {
	if lxr.currentRune == RuneEOF {
		return true
	}

	return false
}

// EatWhitespace is effectively an LTrim in that it reads and discards all whitespace starting with
// the current token up until the next non-whitespace character which becomes the current token.
//
// This can be calle manually by consumers or automatically called before capture functions if
// AutoEatWhitespace is set to true on the lexer.
func (lxr *Lexer) EatWhitespace() bool {
	if !unicode.IsSpace(lxr.currentRune) {
		return false
	}

	for {
		lxr.read()
		if lxr.currentRune == RuneEOF {
			return false
		} else if !unicode.IsSpace(lxr.currentRune) {
			break
		}
	}

	return true
}

func (lxr *Lexer) shutdown() {
	close(lxr.tokens)
}

func (lxr *Lexer) read() rune {

	var ch rune
	if len(lxr.runeCache) > 0 {
		ch = lxr.runeCache[0]
		lxr.runeCache = lxr.runeCache[1:]
		lxr.currentRune = ch
		return ch
	}

	ch, _, err := lxr.inputBuffer.ReadRune()
	if err != nil {
		lxr.currentRune = RuneEOF
		return RuneEOF
	}

	lxr.currentRune = ch
	return ch
}

func (lxr *Lexer) peek(numRunes int) []rune {
	var readBuf []rune
	var peekbuf []rune
	var ch rune
	var err error

	for i := 0; i < numRunes; i++ {
		if len(lxr.runeCache) > 0 {
			ch = lxr.runeCache[0]
			lxr.runeCache = lxr.runeCache[1:]
			err = nil
		} else {
			ch, _, err = lxr.inputBuffer.ReadRune()
		}
		if err == nil {
			readBuf = append(readBuf, ch)
			peekbuf = append(peekbuf, ch)
		}
	}

	lxr.runeCache = append(lxr.runeCache, readBuf...)

	return peekbuf
}

func (lxr *Lexer) skipIgnores() bool {
	if lxr.currentRune == RuneEOF {
		return false
	}

	lxr.enterDebug("skipIgnores")
	foundIgnore := false
	for ignore, doit := range lxr.ignoreTokens {
		lxr.logDebug("testing ignore: %s", ignore)
		if doit && lxr.CurrentTokenIs(ignore) {
			lxr.logDebug("ignoring: %s", ignore)
			foundIgnore = true
			numRunes := utf8.RuneCountInString(ignore)

			for i := 0; i < numRunes; i++ {
				ch := lxr.read()
				lxr.logDebug("read char: %q", ch)
			}
			break
		}
	}

	lxr.exitDebug("skipIgnores")
	return foundIgnore
}

func (lxr *Lexer) enterDebug(format string, a ...interface{}) {
	if lxr.Debug {
		lxr.logIndent++
		s := fmt.Sprintf(format, a...)
		for i := 0; i < lxr.logIndent-1; i++ {
			fmt.Print("    | ")
		}
		fmt.Print("     ")
		fmt.Printf("##### %s #####\n", s)

	}
}

func (lxr *Lexer) exitDebug(format string, a ...interface{}) {
	if lxr.Debug {
		for i := 0; i < lxr.logIndent; i++ {
			fmt.Print("    | ")
		}
		fmt.Print("end\n")
		lxr.logIndent--
	}
}

func (lxr *Lexer) logDebug(format string, a ...interface{}) {
	if lxr.Debug {
		for i := 0; i < lxr.logIndent; i++ {
			fmt.Print("    | ")
		}
		fmt.Printf(format, a...)
		fmt.Print("\n")
	}
}
