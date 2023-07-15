package ioprovider

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type StdIOProvider struct {
	scanner *bufio.Scanner
}

func NewStdIOProvider() *StdIOProvider {
	stdinInputProvider := &StdIOProvider{
		scanner: bufio.NewScanner(os.Stdin),
	}

	return stdinInputProvider
}

func (sip *StdIOProvider) GetInput() (string, error) {
	scanner := sip.scanner
	scanner.Scan()
	err := scanner.Err()
	input := scanner.Text()
	input = strings.TrimSpace(input)
	return input, err
}

func (sip *StdIOProvider) Print(data string) {
	fmt.Println(data)
}
