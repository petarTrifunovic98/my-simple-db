package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/petarTrifunovic98/my-simple-db/pkg/commands"
	"github.com/petarTrifunovic98/my-simple-db/pkg/ioprovider"
	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
)

const prompt string = "my-db> "

func main() {

	t := table.NewTable()
	defer t.DestroyTable()
	fmt.Println("~ Started my db... ")

	server, _ := net.Listen("tcp", "localhost:9988")

	for {
		ioProvider := ioprovider.NewSocketIOProvider(server)
		go proccessRequests(ioProvider, t)
		// ioProvider := ioprovider.NewStdIOProvider()
	}

}

func proccessRequests(ioProvider ioprovider.IIOProvider, t *table.Table) {
	for {
		printPrompt()

		input, err := ioProvider.GetInput()

		if err != nil {
			fmt.Printf("An error occurred while reading input! %v", err)
			return
		}

		inputType := getCommandType(input)

		if inputType == commands.NON_STATEMENT_COMMAND {
			nonStatement := getNonStatementCommand(input)
			nonStatement.PrintPreExecution()
			nonStatement.Execute(t, ioProvider)
		} else {
			statement := getStatementCommand(input)
			statement.PrintPreExecution()
			switch statement.Execute(t, ioProvider) {
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
	} else if input == ".print" {
		return commands.NewNonStatementPrint()
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
	} else if inputParts[0] == "selectOne" {
		return commands.NewStatementSelectOne(input)
	} else {
		return commands.NewStatementUnrecognized(input)
	}
}
