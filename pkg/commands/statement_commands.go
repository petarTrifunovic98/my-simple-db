package commands

import (
	"fmt"
	"strings"
)

type StatementCommandType int8

const (
	STATEMENT_INSERT StatementCommandType = iota
	STATEMENT_SELECT
	STATEMENT_UNRECOGNIZED
)

type StatementSelect struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
}

func (s *StatementSelect) Execute() CommandExecutionStatusCode {
	s.code = SUCCESS
	return s.code
}

func (s *StatementSelect) PrintPreExecution() {
	fmt.Println("Executing select statement")
}

func NewStatementSelect(input string) *StatementSelect {
	statement := &StatementSelect{
		statementType: STATEMENT_SELECT,
	}

	return statement
}

type StatementInsert struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
	args          []string
}

func (s *StatementInsert) Execute() CommandExecutionStatusCode {
	if len(s.args) != 3 {
		s.code = FAILURE
	} else {
		s.code = SUCCESS
	}

	return s.code
}

func (s *StatementInsert) PrintPreExecution() {
	fmt.Println("Executing insert statement")
}

func NewStatementInsert(input string) *StatementInsert {
	statement := &StatementInsert{
		statementType: STATEMENT_INSERT,
	}

	inputParts := strings.Split(input, " ")
	statement.args = inputParts[1:]

	return statement
}

type StatementUnrecognized struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
}

func (s *StatementUnrecognized) Execute() CommandExecutionStatusCode {
	s.code = UNRECOGNIZED
	return s.code
}

func (s *StatementUnrecognized) PrintPreExecution() {
	fmt.Println("Unrecognized statement")
}

func NewStatementUnrecognized(input string) *StatementUnrecognized {
	statement := &StatementUnrecognized{
		statementType: STATEMENT_UNRECOGNIZED,
	}

	return statement
}
