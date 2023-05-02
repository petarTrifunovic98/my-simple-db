package commands

import (
	"fmt"
	"os"
)

type NonStatementCommandType int8

const (
	NS_EXIT NonStatementCommandType = iota
	NS_UNRECOGNIZED
)

type NonStatementExit struct {
	code             CommandExecutionStatusCode
	nonStatementType NonStatementCommandType
}

func (ns *NonStatementExit) Execute() CommandExecutionStatusCode {
	ns.code = SUCCESS
	os.Exit(0)
	return ns.code
}

func (ns *NonStatementExit) PrintPreExecution() {
	fmt.Println("Exiting...")
	fmt.Println("~")
}

func NewNonStatementExit() *NonStatementExit {
	nonStatement := &NonStatementExit{
		nonStatementType: NS_EXIT,
	}

	return nonStatement
}

type NonStatementUnrecognized struct {
	code             CommandExecutionStatusCode
	nonStatementType NonStatementCommandType
}

func (ns *NonStatementUnrecognized) Execute() CommandExecutionStatusCode {
	ns.code = UNRECOGNIZED
	return ns.code
}

func (ns *NonStatementUnrecognized) PrintPreExecution() {
	fmt.Println("Unrecognized non-statement command")
}

func NewNonStatementUnrecognized() *NonStatementUnrecognized {
	nonStatement := &NonStatementUnrecognized{
		nonStatementType: NS_UNRECOGNIZED,
	}

	return nonStatement
}
