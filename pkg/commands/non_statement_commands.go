package commands

import (
	"fmt"
	"os"

	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
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

func (ns *NonStatementExit) Execute(t *table.Table) CommandExecutionStatusCode {
	ns.code = SUCCESS
	t.DestroyTable()
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

func (ns *NonStatementUnrecognized) Execute(t *table.Table) CommandExecutionStatusCode {
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
