package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// SourceCode expands on bytes.Buffer to afford a few extra features
type SourceCode struct {
	buffer *bytes.Buffer
}

// LoadFile loads an entire file into the buffer
func (sc *SourceCode) LoadFile(fileName string) (err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	sc.buffer = new(bytes.Buffer)
	_, err = sc.buffer.ReadFrom(file)
	return
}

// LoadString loads a string into the buffer
func (sc *SourceCode) LoadString(s string) (err error) {
	sc.buffer = bytes.NewBufferString(s)
	return
}

// NextRune reads the nextchar from the buffer
// it replaces the io.EOF error by the UNICODE EOT (End of Transmission) character to allow
// for far easier processing in a read-ahead parser.
func (sc *SourceCode) NextRune() (c rune, err error) {
	c, _, err = sc.buffer.ReadRune()
	if err == io.EOF {
		c = rune(0x04)
		err = nil
	}
	return
}

// PrevRune unreads the last rune so it can be re-processed
func (sc *SourceCode) PrevRune() (err error) {
	err = sc.buffer.UnreadRune()
	return
}

// String inplements the stringer interface so we can show contents
func (sc *SourceCode) String() string {
	return sc.buffer.String()
}

func NewSourceCode() (sc *SourceCode) {
	sc = new(SourceCode)
	sc.buffer = new(bytes.Buffer)
	return
}

var sourceCode *SourceCode

// - Interface ------------------------------------------------------------------------------------------------------------------

func main() {
	if len(os.Args) < 1 {
		fmt.Printf("Missing source file name\n")
		return
	}

	sourceCode = NewSourceCode()
	err := sourceCode.LoadFile(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = nextToken()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Print(sourceCode.String())
}
