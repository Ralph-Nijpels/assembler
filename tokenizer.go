package main

import (
	"fmt"
	"unicode"
)

// - Token ----------------------------------------------------------------------------------------------------------------------

const (
	TK_UNKNOWN = iota
	TK_IDENTIFIER
	TK_INTEGER
	TK_HEXADECIMAL
	TK_FLOAT
	TK_COLON
	TK_BRACKET_OPEN
	TK_BRACKET_CLOSE
	TK_BRACE_OPEN
	TK_BRACE_CLOSE
	TK_END_OF_LINE
)

type Token struct {
	token int
	value string
}

func (thisToken Token) append(c rune) (nextToken Token) {
	nextToken = thisToken
	nextToken.value += string(c)
	return
}

func (thisToken Token) clear() (nextToken Token) {
	nextToken = thisToken
	nextToken.value = ""
	return
}

func NewToken() (token Token) {
	token = Token{}
	return
}

// - Tokenizer ------------------------------------------------------------------------------------------------------------------

const (
	ST_WHITE_SPACE    = iota // Reads leading whitespace before the token
	ST_TOKEN_START           // Interprets the first character of the token
	ST_COMMENT_START         // Tries to 'prove' a comment
	ST_COMMENT               // Reads the comment
	ST_IDENTIFIER            // Reads an identifier
	ST_NEGATIVE              // Reads a negative number
	ST_NUMBER_PREFIX         // Sorts out the type of number
	ST_NUMBER                // Reading decimals digits
	ST_HEXADECIMAL           // reading hexadecimal digits
	ST_FRACTION_START        // reading first decimal after dot
	ST_FRACTION              // reading next decimals after dot
	ST_END            = 999  // Token read, all is well
)

type State func(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error)

// white_space skips over any empty stuff before anything actually happens
func white_space(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	if unicode.IsSpace(thisChar) {
		nextChar, err = sourceCode.NextRune()
		return
	}
	nextChar = thisChar
	nextToken = thisToken
	state = ST_TOKEN_START
	return
}

// token_start makes the initial categorization of the token
func token_start(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// colon is a single symbol token all by itself
	if thisChar == rune(':') {
		nextToken.token = TK_COLON
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	// Brackets are single symbols all by themselves
	if thisChar == rune('(') {
		nextToken.token = TK_BRACKET_OPEN
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	if thisChar == rune(')') {
		nextToken.token = TK_BRACKET_CLOSE
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	// Braces are single symbols all by themselves
	if thisChar == rune('{') {
		nextToken.token = TK_BRACE_OPEN
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	if thisChar == rune('}') {
		nextToken.token = TK_BRACE_CLOSE
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	// Comments starts with '//'
	if thisChar == rune('/') {
		nextChar, err = sourceCode.NextRune()
		state = ST_COMMENT_START
		return
	}
	// an identifier has started
	if unicode.IsLetter(thisChar) || thisChar == rune('_') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_IDENTIFIER
		return
	}
	// a negative number has started
	if thisChar == rune('-') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_NEGATIVE
		return
	}
	// a float between <0..1> has started
	if thisChar == rune('.') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_FRACTION_START
		return
	}
	// a Hexadecimal or Float number may have started
	if thisChar == rune('0') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_NUMBER_PREFIX
		return
	}
	// a number has started
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_NUMBER
		return
	}
	// End of line is a token all by itself
	if thisChar == rune('\n') {
		nextToken.token = TK_END_OF_LINE
		nextChar, err = sourceCode.NextRune()
		state = ST_END
		return
	}
	err = fmt.Errorf("unknown token")
	return
}

// comment_start checks if there is a second '/' if not, we stop
func comment_start(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	if thisChar == rune('/') {
		nextChar, err = sourceCode.NextRune()
		state = ST_COMMENT
		return
	}
	err = fmt.Errorf("unknown token (expected '/')")
	return
}

// comment skips the content of the comment until EOLN
func comment(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	if thisChar != rune('\n') {
		nextChar, err = sourceCode.NextRune()
		state = ST_COMMENT
		return
	}
	nextChar, err = sourceCode.NextRune()
	nextToken.token = TK_END_OF_LINE
	state = ST_END
	return
}

// identifierToken reads the rest of an identifier
func identifier(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// the identifier continues
	if unicode.IsLetter(thisChar) || unicode.IsDigit(thisChar) || thisChar == rune('_') || thisChar == rune('-') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_IDENTIFIER
		return
	}
	// the identifier is done
	nextToken.token = TK_IDENTIFIER
	nextToken.value = thisToken.value
	nextChar = thisChar
	state = ST_END
	return
}

// negativeNumberToken reads the rest of a negative number
func negative(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_NUMBER
		return
	}
	// oops
	err = fmt.Errorf("invalid token (malformed number)")
	return
}

// number_prefix checks if we are reading a hexadecimal number or a float that happens to start with '0'
func number_prefix(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// Check if it is a floating point number
	if thisChar == rune('.') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_FRACTION_START
		return
	}
	// Check if it is a hexadecimal number
	if thisChar == rune('x') || thisChar == rune('X') {
		nextToken.value = "" // the value is without the 0X prefix
		nextChar, err = sourceCode.NextRune()
		state = ST_HEXADECIMAL
		return
	}
	// oops
	err = fmt.Errorf("invalid token (malformed number)")
	return
}

// number reads the whole part of a number
func number(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_NUMBER
		return
	}
	// Cloud be float
	if thisChar == rune('.') {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_FRACTION_START
		return
	}
	// the number is done
	nextToken.token = TK_INTEGER
	nextToken.value = thisToken.value
	nextChar = thisChar
	state = ST_END
	return
}

// hexadecimal reacs all hexadecimal digits
func hexadecimal(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {

	hexadecimals := unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: uint16(rune('0')), Hi: uint16(rune('9')), Stride: 1},
			{Lo: uint16(rune('A')), Hi: uint16(rune('F')), Stride: 1},
			{Lo: uint16(rune('a')), Hi: uint16(rune('f')), Stride: 1},
		},
	}

	// This must be a hexadecimal digit
	if unicode.Is(&hexadecimals, thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_HEXADECIMAL
		return
	}

	// hexadecimal is done
	nextToken.token = TK_HEXADECIMAL
	nextToken.value = thisToken.value
	nextChar = thisChar
	state = ST_END
	return
}

func fraction_start(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		state = ST_FRACTION
		return
	}
	err = fmt.Errorf("invalid token (expected decimal)")
	return
}

func fraction(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, err = sourceCode.NextRune()
		return
	}
	// the float is done
	nextToken.token = TK_FLOAT
	nextToken.value = thisToken.value
	state = 999
	return
}

// nextToken reads the next token from the buffer, using a classic handcrafted state machine.
func nextToken() (token Token, err error) {

	stateTable := []State{
		white_space,
		token_start,
		comment_start,
		comment,
		identifier,
		negative,
		number_prefix,
		number,
		hexadecimal,
		fraction_start,
		fraction}

	state := 0
	token = NewToken()
	thisChar, err := sourceCode.NextRune()
	for err == nil && state != 999 {
		state, thisChar, token, err = stateTable[state](thisChar, token)
	}
	if err == nil {
		sourceCode.PrevRune()
	}

	return
}
