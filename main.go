package main

import (
	"bytes"
	"fmt"
	"os"
)

var sourceCode *bytes.Buffer

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
