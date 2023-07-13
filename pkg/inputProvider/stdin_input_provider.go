package inputprovider

import (
	"bufio"
	"os"
	"strings"
)

type StdinInputProvider struct {
	scanner *bufio.Scanner
}

func NewStdinInputProvider() *StdinInputProvider {
	stdinInputProvider := &StdinInputProvider{
		scanner: bufio.NewScanner(os.Stdin),
	}

	return stdinInputProvider
}

func (sip *StdinInputProvider) GetInput() (string, error) {
	scanner := sip.scanner
	scanner.Scan()
	err := scanner.Err()
	input := scanner.Text()
	input = strings.TrimSpace(input)
	return input, err
}
