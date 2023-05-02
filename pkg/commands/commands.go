package commands

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
	Execute() CommandExecutionStatusCode
	PrintPreExecution()
}
