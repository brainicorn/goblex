//Package goblex (go buffering lexer) implements a low-level library containing functions to easily
//build a lexer in Go.
//
//This library differs from the other lexer tools out there in that it uses an internal buffer to
//capture the tokens to be emitted rather than using regular expressions and/or a position based lexer.
//
//For more information about why this is useful, see the README file located here:
//https://github.com/brainicorn/goblex/src/master/README.md
//
//For a quick start, see the example below
package goblex
