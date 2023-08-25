package commands

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/petarTrifunovic98/my-simple-db/pkg/ioprovider"
	"github.com/petarTrifunovic98/my-simple-db/pkg/row"
	"github.com/petarTrifunovic98/my-simple-db/pkg/serialization"
	"github.com/petarTrifunovic98/my-simple-db/pkg/table"
)

type StatementCommandType int8

const (
	STATEMENT_INSERT StatementCommandType = iota
	STATEMENT_SELECT
	STATEMENT_SELECT_ONE
	STATEMENT_UNRECOGNIZED
)

type StatementSelect struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
}

func (s *StatementSelect) Execute(t *table.Table, ip ioprovider.IIOProvider) CommandExecutionStatusCode {
	s.code = SUCCESS

	values := t.Select()
	if len(values) <= 0 {
		return s.code
	}

	rows := make([]*row.RowDTO, 0)

	b := bytes.NewBuffer(values)
	for {
		r := &row.Row{}
		err := serialization.Deserialize(b, r)
		if err != nil {
			break
		}
		rows = append(rows, r.ToRowDTO())
		// forPrinting := r.ToString()
		// ip.Print(forPrinting)
	}

	jsonBytes, _ := json.Marshal(rows)
	ip.Print(string(jsonBytes))

	// ip.Print("end")

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

type StatementSelectOne struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
	key           string
}

func (s *StatementSelectOne) Execute(t *table.Table, ip ioprovider.IIOProvider) CommandExecutionStatusCode {
	s.code = SUCCESS

	id, err := strconv.Atoi(s.key)
	if err != nil {
		s.code = FAILURE
		return s.code
	}

	keyBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(keyBytes, uint32(id))

	value := t.SelectOne(keyBytes)
	if len(value) <= 0 {
		return s.code
	}

	b := bytes.NewBuffer(value)
	var r *row.Row

	r = &row.Row{}
	err = serialization.Deserialize(b, r)

	jsonBytes, _ := json.Marshal(r.ToRowDTO())
	ip.Print(string(jsonBytes))
	return s.code
}

func (s *StatementSelectOne) PrintPreExecution() {
	fmt.Println("Executing select statement")
}

func NewStatementSelectOne(input string) *StatementSelectOne {
	statement := &StatementSelectOne{
		statementType: STATEMENT_SELECT_ONE,
	}

	inputParts := strings.Split(input, " ")
	statement.key = inputParts[1]

	return statement
}

type StatementInsert struct {
	code          CommandExecutionStatusCode
	statementType StatementCommandType
	args          []string
}

func (s *StatementInsert) Execute(t *table.Table, ip ioprovider.IIOProvider) CommandExecutionStatusCode {
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

func (s *StatementUnrecognized) Execute(t *table.Table, ip ioprovider.IIOProvider) CommandExecutionStatusCode {
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
