package commands

import (
	"github.com/petarTrifunovic98/my-simple-db/pkg/ioprovider"
	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
)

type CommandType int8

const (
	NON_STATEMENT_COMMAND CommandType = iota
	STATEMENT_COMMAND
)

type CommandExecutionStatusCode int8

const (
	SUCCESS CommandExecutionStatusCode = iota
	FAILURE
	UNRECOGNIZED
)

type Command interface {
	Execute(t *table.Table, ip ioprovider.IIOProvider) CommandExecutionStatusCode
	PrintPreExecution()
}
