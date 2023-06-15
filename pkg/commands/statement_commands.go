package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/petarTrifunovic98/my-simple-db/pkg/row"
	"github.com/petarTrifunovic98/my-simple-db/pkg/serialization"
	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
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

func (s *StatementSelect) Execute(t *table.Table) CommandExecutionStatusCode {
	s.code = SUCCESS

	values := t.Select()
	if len(values) <= 0 {
		return s.code
	}

	b := bytes.NewBuffer(values)
	for {
		r := &row.Row{}
		err := serialization.Deserialize(b, r)
		if err != nil {
			break
		}
		r.Print()
	}
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

func (s *StatementInsert) Execute(t *table.Table) CommandExecutionStatusCode {
	if len(s.args) != 3 {
		s.code = FAILURE
	} else {
		newRow := &row.Row{}
		id, err := strconv.Atoi(s.args[0])
		if err != nil {
			s.code = FAILURE
			return s.code
		}

		newRow.Id = uint32(id)
		copy(newRow.Username[:], []byte(s.args[1]))
		copy(newRow.Email[:], []byte(s.args[2]))

		rowBytes := serialization.Serialize(newRow)

		keyBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(keyBytes, newRow.Id)
		t.Insert(keyBytes, rowBytes)

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

func (s *StatementUnrecognized) Execute(t *table.Table) CommandExecutionStatusCode {
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
