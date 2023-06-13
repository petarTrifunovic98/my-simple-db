package commands

import (
	"fmt"
	"os"

	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
)

type NonStatementCommandType int8

const (
	NS_EXIT NonStatementCommandType = iota
	NS_PRINT
	NS_UNRECOGNIZED
)

type NonStatementBase struct {
	code             CommandExecutionStatusCode
	nonStatementType NonStatementCommandType
}

type NonStatementExit struct {
	NonStatementBase
}

func (ns *NonStatementExit) Execute(t *table.Table) CommandExecutionStatusCode {
	ns.code = SUCCESS
	t.DestroyTable2()
	os.Exit(0)
	return ns.code
}

func (ns *NonStatementExit) PrintPreExecution() {
	fmt.Println("Exiting...")
	fmt.Println("~")
}

func NewNonStatementExit() *NonStatementExit {
	nonStatement := &NonStatementExit{
		NonStatementBase: NonStatementBase{
			nonStatementType: NS_EXIT,
		},
	}

	return nonStatement
}

type NonStatementPrint struct {
	NonStatementBase
}

func (ns *NonStatementPrint) Execute(t *table.Table) CommandExecutionStatusCode {
	t.PrintInternalStructure()
	ns.code = SUCCESS
	return ns.code
}

func (ns *NonStatementPrint) PrintPreExecution() {
	fmt.Println("Showing internal table structure")
}

func NewNonStatementPrint() *NonStatementPrint {
	nonStatement := &NonStatementPrint{
		NonStatementBase: NonStatementBase{
			nonStatementType: NS_PRINT,
		},
	}

	return nonStatement
}

type NonStatementUnrecognized struct {
	NonStatementBase
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
		NonStatementBase: NonStatementBase{
			nonStatementType: NS_UNRECOGNIZED,
		},
	}

	return nonStatement
}
