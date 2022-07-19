package main

import (
	"testing"
)

// - Support functions to prevent repetition ------------------------------------------------------------------------------------

type StateCase struct {
	expectedChar  rune
	expectedState int
	expectedToken int
	expectedValue string
}

func (c StateCase) verify(t *testing.T, caseId int, state int, nextChar rune, token Token, err error) {
	if err != nil {
		t.Errorf("CaseID %d: %v", caseId, err.Error())
	}
	if nextChar != c.expectedChar {
		t.Errorf("CaseID %d: wrong char, expected %s got %s", caseId, string(c.expectedChar), string(nextChar))
	}
	if state != c.expectedState {
		t.Errorf("CaseID %d: wrong state, expected %d, got %d", caseId, c.expectedState, state)
	}
	if token.token != c.expectedToken {
		t.Errorf("CaseID %d: wrong token, expected %d, got %d", caseId, c.expectedToken, token.token)
	}
	if token.value != c.expectedValue {
		t.Errorf("CaseID %d: wrong value, expected \"%s\", got \"%s\"", caseId, c.expectedValue, token.value)
	}
}

type TokenizerCase struct {
	sourceCode    string
	expectedToken int
	expectedValue string
	expectedChar  rune
}

func (c TokenizerCase) verify(t *testing.T, caseId int) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(c.sourceCode)

	// Read token
	token, err := nextToken()
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	if token.token != c.expectedToken {
		t.Errorf("wrong token: expected %d, got %d", c.expectedToken, token.token)
	}
	if token.value != c.expectedValue {
		t.Errorf("wrong value: expected \"%s\", got \"%s\"", c.expectedValue, token.value)
	}
	// Check we start off ok, next time
	nextChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	if nextChar != c.expectedChar {
		t.Errorf("wrong char: expected %s, got %s", string(c.expectedChar), string(nextChar))
	}
}

// - Test Token -----------------------------------------------------------------------------------------------------------

func TestNewToken(t *testing.T) {
	token := NewToken()
	if token.token != TK_UNKNOWN {
		t.Errorf("wrong token, expected %d, got %d", TK_UNKNOWN, token.token)
	}
	if token.value != "" {
		t.Errorf("wrong value, expected \"\", got \"%s\"", token.value)
	}
}

func TestAppend(t *testing.T) {
	token := NewToken()
	token = token.append(rune('A'))
	if token.token != TK_UNKNOWN {
		t.Errorf("wrong token, expected %d, got %d", TK_UNKNOWN, token.token)
	}
	if token.value != "A" {
		t.Errorf("wrong value, expected \"A\", got \"%s\"", token.value)
	}
}

// - Test state functions -------------------------------------------------------------------------------------------------------

func TestWhiteSpace(t *testing.T) {

	sourceCode = NewSourceCode()
	sourceCode.LoadString(" X")

	testCases := []StateCase{
		{rune('X'), ST_WHITE_SPACE, TK_UNKNOWN, ""},
		{rune('X'), ST_TOKEN_START, TK_UNKNOWN, ""}}

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	state := ST_WHITE_SPACE
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = white_space(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
	}
}

func TestTokenStart(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(":(){}/Aa_-.07\n")

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('('), ST_END, TK_COLON, ""},
		{rune(')'), ST_END, TK_BRACKET_OPEN, ""},
		{rune('{'), ST_END, TK_BRACKET_CLOSE, ""},
		{rune('}'), ST_END, TK_BRACE_OPEN, ""},
		{rune('/'), ST_END, TK_BRACE_CLOSE, ""},
		{rune('A'), ST_COMMENT_START, TK_UNKNOWN, ""},
		{rune('a'), ST_IDENTIFIER, TK_UNKNOWN, "A"},
		{rune('_'), ST_IDENTIFIER, TK_UNKNOWN, "a"},
		{rune('-'), ST_IDENTIFIER, TK_UNKNOWN, "_"},
		{rune('.'), ST_NEGATIVE, TK_UNKNOWN, "-"},
		{rune('0'), ST_FRACTION_START, TK_UNKNOWN, "."},
		{rune('7'), ST_NUMBER_PREFIX, TK_UNKNOWN, "0"},
		{rune('\n'), ST_NUMBER, TK_UNKNOWN, "7"},
		{rune(0x04), ST_END, TK_END_OF_LINE, ""},
	}

	state := ST_TOKEN_START
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = token_start(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	_, _, _, err = token_start(rune('!'), token) // unknown
	if err == nil {
		t.Errorf("Expected \"unknown token\" error")
	}
}

func TestCommentStart(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString("/")

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune(0x04), ST_COMMENT, TK_UNKNOWN, ""}}

	state := ST_COMMENT_START
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = comment(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	_, _, _, err = comment_start(rune('!'), token) // unknown
	if err == nil {
		t.Errorf("expected \"unknown token (expected '/')\" error")
	}
}

func TestComment(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString("Aa0_-.\n")

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('a'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune('0'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune('_'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune('-'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune('.'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune('\n'), ST_COMMENT, TK_UNKNOWN, ""},
		{rune(0x04), ST_END, TK_END_OF_LINE, ""}}

	state := ST_COMMENT
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = comment(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}
}

func TestIdentifier(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString("AZaz09_-!")

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('Z'), ST_IDENTIFIER, TK_UNKNOWN, "A"},
		{rune('a'), ST_IDENTIFIER, TK_UNKNOWN, "Z"},
		{rune('z'), ST_IDENTIFIER, TK_UNKNOWN, "a"},
		{rune('0'), ST_IDENTIFIER, TK_UNKNOWN, "z"},
		{rune('9'), ST_IDENTIFIER, TK_UNKNOWN, "0"},
		{rune('_'), ST_IDENTIFIER, TK_UNKNOWN, "9"},
		{rune('-'), ST_IDENTIFIER, TK_UNKNOWN, "_"},
		{rune('!'), ST_IDENTIFIER, TK_UNKNOWN, "-"},
		{rune('!'), ST_END, TK_IDENTIFIER, ""},
	}

	state := ST_IDENTIFIER
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = identifier(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}
}

func TestNegative(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(("09"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('9'), ST_NUMBER, TK_UNKNOWN, "0"},
		{rune(0x04), ST_NUMBER, TK_UNKNOWN, "9"},
	}

	state := ST_NEGATIVE
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = negative(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	_, _, _, err = negative(rune('-'), token)
	if err == nil {
		t.Errorf("expected \"invalid token (malformed number)\" error")
	}

	_, _, _, err = negative(rune('!'), token)
	if err == nil {
		t.Errorf("expected \"invalid token (malformed number)\" error")
	}
}

func TestNumberPrefix(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString((".xX"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('x'), ST_FRACTION_START, TK_UNKNOWN, "."},
		{rune('X'), ST_HEXADECIMAL, TK_UNKNOWN, ""},
		{rune(0x04), ST_HEXADECIMAL, TK_UNKNOWN, ""},
	}

	state := ST_NUMBER_PREFIX
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = number_prefix(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	testCase := StateCase{rune('-'), ST_END, TK_INTEGER, ""}
	state, thisChar, token, err = number_prefix(rune('-'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()

	testCase = StateCase{rune('!'), ST_END, TK_INTEGER, ""}
	state, thisChar, token, err = number_prefix(rune('!'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()
}

func TestNumber(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(("09.!"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('9'), ST_NUMBER, TK_UNKNOWN, "0"},
		{rune('.'), ST_NUMBER, TK_UNKNOWN, "9"},
		{rune('!'), ST_FRACTION_START, TK_UNKNOWN, "."},
		{rune('!'), ST_END, TK_INTEGER, ""},
	}

	state := ST_NUMBER
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = number(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}
}

func TestHexadecimal(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(("09afAF!"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('9'), ST_HEXADECIMAL, TK_UNKNOWN, "0"},
		{rune('a'), ST_HEXADECIMAL, TK_UNKNOWN, "9"},
		{rune('f'), ST_HEXADECIMAL, TK_UNKNOWN, "a"},
		{rune('A'), ST_HEXADECIMAL, TK_UNKNOWN, "f"},
		{rune('F'), ST_HEXADECIMAL, TK_UNKNOWN, "A"},
		{rune('!'), ST_HEXADECIMAL, TK_UNKNOWN, "F"},
		{rune('!'), ST_END, TK_HEXADECIMAL, ""},
	}

	state := ST_HEXADECIMAL
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = hexadecimal(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	testCase := StateCase{rune('g'), ST_END, TK_HEXADECIMAL, ""}
	state, thisChar, token, err = hexadecimal(rune('g'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()

	testCase = StateCase{rune('G'), ST_END, TK_HEXADECIMAL, ""}
	state, thisChar, token, err = hexadecimal(rune('G'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()
}

func TestFractionStart(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(("09"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('9'), ST_FRACTION, TK_UNKNOWN, "0"},
		{rune(0x04), ST_FRACTION, TK_UNKNOWN, "9"},
	}

	state := ST_FRACTION_START
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = fraction_start(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	_, _, _, err = fraction_start(rune('.'), token)
	if err == nil {
		t.Errorf("expected \"invalid token (malformed number)\" error")
	}

	_, _, _, err = fraction_start(rune('!'), token)
	if err == nil {
		t.Errorf("expected \"invalid token (malformed number)\" error")
	}
}

func TestFraction(t *testing.T) {
	sourceCode = NewSourceCode()
	sourceCode.LoadString(("09"))

	thisChar, err := sourceCode.NextRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	testCases := []StateCase{
		{rune('9'), ST_FRACTION, TK_UNKNOWN, "0"},
		{rune(0x04), ST_FRACTION, TK_UNKNOWN, "9"},
	}

	state := ST_FRACTION_START
	token := NewToken()
	for id, c := range testCases {
		state, thisChar, token, err = fraction(thisChar, token)
		c.verify(t, id, state, thisChar, token, err)
		token = token.clear()
	}

	testCase := StateCase{rune('.'), ST_END, TK_FLOAT, ""}
	state, thisChar, token, err = fraction(rune('.'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()

	testCase = StateCase{rune('!'), ST_END, TK_FLOAT, ""}
	state, thisChar, token, err = fraction(rune('!'), token)
	testCase.verify(t, -1, state, thisChar, token, err)
	token.clear()

}

// - Test Tokenizer -------------------------------------------------------------------------------------------------------------

func TestNextToken(t *testing.T) {
	// Cases to test
	// TK_IDENTIFIER
	// TK_INTEGER
	// TK_HEXADECIMAL
	// TK_FLOAT
	// TK_COLON
	// TK_BRACKET_OPEN
	// TK_BRACKET_CLOSE
	// TK_BRACE_OPEN
	// TK_BRACE_CLOSE
	// TK_END_OF_LINE

	testCases := []TokenizerCase{
		{"Identifier", TK_IDENTIFIER, "Identifier", rune(0x04)},
		{"Identifier\n", TK_IDENTIFIER, "Identifier", rune('\n')},
		{"_ID", TK_IDENTIFIER, "_ID", rune(0x04)},
		{"0", TK_INTEGER, "0", rune(0x04)},
	}

	for i, c := range testCases {
		c.verify(t, i)
	}
}
