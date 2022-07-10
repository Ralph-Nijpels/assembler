package main

import (
	"bytes"
	"fmt"
	"os"
	"unicode"
)

var sourceCode *bytes.Buffer

// - Tokenizer ------------------------------------------------------------------------------------------------------------------

const (
	TK_IDENTIFIER  = 1
	TK_INTEGER     = 2
	TK_HEXADECIMAL = 3
	TK_FLOAT       = 4
	TK_COLON       = 5
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

type State func(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error)

// leadingWhitespace skips over any empty stuff before anything actually happens
func leadingWhitespace(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	if unicode.IsSpace(thisChar) {
		nextChar, _, err = sourceCode.ReadRune()
		return
	}
	nextChar = thisChar
	nextToken = thisToken
	state = 1
	return
}

// startToken makes the initial categorization of the token
func startToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// double colon is a single symbol token all by itself
	if thisChar == rune(':') {
		nextToken = thisToken.append(thisChar)
		nextToken.token = TK_COLON
		nextChar, _, err = sourceCode.ReadRune()
		state = 999 // end
		return
	}
	// an identifier has started
	if unicode.IsLetter(thisChar) || thisChar == rune('_') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 2 // reading identifiers
		return
	}
	// a negative number has started
	if thisChar == rune('-') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 3 // reading negative number
		return
	}
	// a Hexadecimal or Float number may have started
	if thisChar == rune('0') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 4 // maybe hex, maybe float
		return
	}
	// a number has started
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 5 // reading numbers
		return
	}
	err = fmt.Errorf("unknown token")
	return
}

// identifierToken reads the rest of an identifier
func identifierToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// the identifier continues
	if unicode.IsLetter(thisChar) || unicode.IsDigit(thisChar) || thisChar == rune('_') || thisChar == rune('-') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 2 // reading identifiers
		return
	}
	// the identifier is done
	nextToken.token = TK_IDENTIFIER
	nextToken.value = thisToken.value
	state = 999
	return
}

// negativeNumberToken reads the rest of a negative number
func negativeNumberToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 5 // reading numbers
		return
	}
	// oops
	err = fmt.Errorf("invalid token (malformed number)")
	return
}

// hexOrFloatToken checks if we are reading a hexadecimal number or a float that happens to start with '0'
func hexOrFloatToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// Check if it is a floating point number
	if thisChar == rune('.') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 7
		return
	}
	// Check if it is a hexadecimal number
	if thisChar == rune('x') || thisChar == rune('X') {
		nextToken.value = "" // the value is without the 0X prefix
		nextChar, _, err = sourceCode.ReadRune()
		state = 6
		return
	}
	// oops
	err = fmt.Errorf("invalid token (malformed number)")
	return
}

// numberToken reads the whole part of a number
func numberToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 5 // reading numbers
		return
	}
	// Cloud be float
	if thisChar == rune('.') {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 7 // reading fraction
	}
	// the number is done
	nextToken.token = TK_INTEGER
	nextToken.value = thisToken.value
	state = 999
	return
}

func hexadecimalToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {

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
		nextChar, _, err = sourceCode.ReadRune()
		state = 7
		return
	}

	// hexadecimal is done
	nextToken.token = TK_HEXADECIMAL
	nextToken.value = thisToken.value
	state = 999
	return
}

func fractionToken(thisChar rune, thisToken Token) (state int, nextChar rune, nextToken Token, err error) {
	// This must be a digit
	if unicode.IsDigit(thisChar) {
		nextToken = thisToken.append(thisChar)
		nextChar, _, err = sourceCode.ReadRune()
		state = 4 // reading numbers
		return
	}
	// the float is done
	nextToken.token = TK_FLOAT
	nextToken.value = thisToken.value
	state = 999
	return
}

// nextToken reads the next token from the buffer, using a classic handcrafted state machine.
func nextToken() (token string, err error) {

	stateTable := []State{
		leadingWhitespace,   // state: 0
		startToken,          // state: 1
		identifierToken,     // state: 2
		negativeNumberToken, // state: 3
		hexOrFloatToken,     // state: 4
		numberToken,         // state: 5
		hexadecimalToken,    // state: 6
		fractionToken,       // state 7
	}

	state := 0
	thisToken := Token{}
	thisChar, _, err := sourceCode.ReadRune()
	for err == nil && state != 999 {
		state, thisChar, thisToken, err = stateTable[state](thisChar, thisToken)
	}
	if err == nil {
		sourceCode.UnreadRune()
	}

	return
}

// - File Handling --------------------------------------------------------------------------------------------------------------

func readSource(fileName string) (err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	sourceCode = new(bytes.Buffer)
	_, err = sourceCode.ReadFrom(file)

	return
}

// - Interface ------------------------------------------------------------------------------------------------------------------

func main() {
	if len(os.Args) < 1 {
		fmt.Printf("Missing source file name\n")
		return
	}
	err := readSource(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = nextToken()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Print(sourceCode.String())
}
