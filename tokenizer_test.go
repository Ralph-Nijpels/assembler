package main

import (
	"bytes"
	"testing"
)

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

func TestWhiteSpace(t *testing.T) {
	sourceCode = bytes.NewBufferString(" X")

	thisChar, _, err := sourceCode.ReadRune()
	if err != nil {
		t.Errorf(err.Error())
	}

	state, nextChar, token, err := white_space(thisChar, NewToken())
	if err != nil {
		t.Errorf(err.Error())
	}
	if nextChar != rune('X') {
		t.Errorf("wrong char, expected %s got %s", "X", string(nextChar))
	}
	if state != ST_WHITE_SPACE {
		t.Errorf("wrong state, expected %d, got %d", ST_WHITE_SPACE, state)
	}
	if token.token != TK_UNKNOWN {
		t.Errorf("wrong token, expected %d, got %d", TK_UNKNOWN, token.token)
	}
	if token.value != "" {
		t.Errorf("wrong value, expected \"A\", got \"%s\"", token.value)
	}

	state, nextChar, token, err = white_space(nextChar, token)
	if err != nil {
		t.Errorf(err.Error())
	}
	if nextChar != rune('X') {
		t.Errorf("wrong char, expected %s got %s", "X", string(nextChar))
	}
	if state != ST_TOKEN_START {
		t.Errorf("wrong state, expected %d, got %d", ST_TOKEN_START, state)
	}
	if token.token != TK_UNKNOWN {
		t.Errorf("wrong token, expected %d, got %d", TK_UNKNOWN, token.token)
	}
	if token.value != "" {
		t.Errorf("wrong value, expected \"A\", got \"%s\"", token.value)
	}
}
