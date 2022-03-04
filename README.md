[![Build Status](https://travis-ci.org/brainicorn/goblex.svg?branch=main)](https://travis-ci.org/brainicorn/goblex)
[![codecov](https://codecov.io/gh/brainicorn/goblex/branch/main/graph/badge.svg?token=JEzGtB10Aj)](https://codecov.io/gh/brainicorn/goblex)
[![Go Report Card](https://goreportcard.com/badge/github.com/brainicorn/goblex)](https://goreportcard.com/report/github.com/brainicorn/goblex)
[![GoDoc](https://godoc.org/github.com/brainicorn/goblex?status.svg)](https://godoc.org/github.com/brainicorn/goblex)

# goblex

Package goblex (go buffering lexer) is a library for easily building lexers in Go by utilizing an
internal capture buffer so consumers can emit the tokens they care about and forget about the tokens
they don't.

API Documentation: [https://godoc.org/github.com/brainicorn/goblex](https://godoc.org/github.com/brainicorn/goblex)

[Issue Tracker](https://github.com/brainicorn/goblex/issues)

## Yet Another Go Lexer

There are a lot of Go libraries available for building lexers and like this one a lot of them borrow
ideas from [Rob Pike's talk on building lexers.](https://www.youtube.com/watch?v=HxaD_trXwRE)

The difference with this library is that it uses an internal capture buffer to capture/emit tokens
instead of relying on marked positions on the input and/or using regular expressions.

### Basic Use

Consider a simple example:

Say we want to extract a name/ident from between square brackets like

```
[bob]
```

Simple enough for just about any lexer out there. But let's say we need to find these tokens from
within code comments. Let's also say that the author of the comment can put whitespace anywhere she
likes. All of a sudden our use-cases get way more complex:

```go
// [bob]
//
// [
// sally
// ]
//
// [jim
// ]
//
// [
// jane]
/*
 [
  janet
]
*/
```

Even if we tell the lexer to ignore whitespace which all lexers are capable of, we now have:

```go
//[bob]////[//sally//]////[jim//]////[//jane]/*[janet]*/
```

Getting a positional based lexer to do the right thing in this scenario is complicated.

On the other hand, retrieving all of the names with goblex is trivial:

```go
var token goblex.Token
var nameType TokenType = 1
var lexLeft LexFn
var lexRight LexFn

name := ""

lexLeft = func(lexer *goblex.Lexer) goblex.LexFn {
	// find a left bracket
	if lexer.CaptureUntil(true, "[")
		//skip the bracket and start a new capture buffer
		lexer.SkipCurrentToken(true)
		//capture an ident skipping whitespace
		if lexer.CaptureIdent(true) {
			//if an ident was found, run the right side lexer
			return lexRight
		}
	}

	//no more brackets, end lexing
	return nil
}

lexRight = func(lexer *goblex.Lexer) goblex.LexFn {
	// if we have a right bracket emit the name
	if lexer.NextTokenIs(true, "]") {
		lexer.Emit(nameType)
	}

	//start over
	return lexLeft
}

l := goblex.NewLexer("name lexer", input, lexLeft)

//make sure we skip all comment tokens
l.AddIgnoreTokens("//","/*","*/")

for {
	if l.IsEOF() {
		fmt.Println("EOF")
		break
	}

	token = l.NextEmittedToken()
	switch token.Type {
	case nameType:
		name = token.String()
	}

	fmt.Println(name)
}

OUTPUT:
bob
sally
jim
jane
janet
```

## More Information

For more information and detailes usage see:

[Issue Tracker](https://github.com/brainicorn/goblex/issues)

API Documentation: [https://godoc.org/github.com/brainicorn/goblex](https://godoc.org/github.com/brainicorn/goblex)

Doc Examples: [https://github.com/brainicorn/goblex/src/doc_examples_test.go](https://github.com/brainicorn/goblex/src/doc_examples_test.go)

Java-style Annotation Parser using this library: [https://github.com/brainicorn/ganno/](https://github.com/brainicorn/ganno/)

## Contributors

Pull requests, issues and comments welcome. For pull requests:

- Add tests for new features and bug fixes
- Follow the existing style
- Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating
an issue and explaining the intended change.

## License

Apache 2.0 licensed, see [LICENSE.txt](LICENSE.txt) file.
