package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/petarTrifunovic98/my-simple-db/pkg/commands"
)

type Row struct {
	id       uint32
	username string
	email    string
}

const prompt string = "my-db> "

func main() {
	fmt.Println("~ Started my db... ")

	for {
		printPrompt()

		input, err := getInput()
		if err != nil {
			fmt.Printf("An error occurred while reading input! %v", err)
			return
		}

		inputType := getCommandType(input)

		if inputType == commands.NON_STATEMENT_COMMAND {
			nonStatement := getNonStatementCommand(input)
			nonStatement.PrintPreExecution()
			nonStatement.Execute()
		} else {
			statement := getStatementCommand(input)
			statement.PrintPreExecution()
			switch statement.Execute() {
			case commands.SUCCESS:
				fmt.Println("Success")
			case commands.FAILURE:
				fmt.Println("Failure")
			case commands.UNRECOGNIZED:
				fmt.Println("Unrecognized")
			}

		}

	}
}

func printPrompt() {
	fmt.Print(prompt)
}

func getInput() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	input := scanner.Text()
	input = strings.TrimSpace(input)
	return input, err
}

func getCommandType(input string) commands.CommandType {
	if len(input) < 1 || input[0] == '.' {
		return commands.NON_STATEMENT_COMMAND
	} else {
		return commands.STATEMENT_COMMAND
	}
}

func getNonStatementCommand(input string) commands.Command {
	if input == ".exit" {
		return commands.NewNonStatementExit()
	} else {
		return commands.NewNonStatementUnrecognized()
	}
}

func getStatementCommand(input string) commands.Command {
	inputParts := strings.Split(input, " ")

	if inputParts[0] == "insert" {
		return commands.NewStatementInsert(input)
	} else if inputParts[0] == "select" {
		return commands.NewStatementSelect(input)
	} else {
		return commands.NewStatementUnrecognized(input)
	}
}
