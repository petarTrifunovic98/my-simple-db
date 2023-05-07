package table

import (
	"bytes"

	"github.com/petarTrifunovic98/my-simple-db/pkg/paging"
)

type Page struct {
	SerializedRows bytes.Buffer
}

type Table struct {
	NumRows uint32
	Pager   *paging.Pager
}

func NewTable() *Table {
	table := &Table{
		NumRows: 0,
		Pager:   paging.NewPager("./db"),
	}

	return table
}

func (t *Table) Insert(key any, data any) {
	t.Pager.AddToCurrentPage(data)
}

func (t *Table) Select() []byte {
	values := t.Pager.ReadWholeCurrentPage()
	return values
}

func (t *Table) DestroyTable() {
	t.Pager.ClearPager()
}
