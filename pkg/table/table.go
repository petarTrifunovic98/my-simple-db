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

func (t *Table) Insert(key uint32, data []byte) {
	t.Pager.AddNewData(key, data)
}

func (t *Table) Select() []byte {
	values := t.Pager.ReadWholeCurrentPageTemp()
	return values
}

func (t *Table) PrintInternalStructure() {
	t.Pager.PrintPages()
}

func (t *Table) DestroyTable() {
	t.Pager.ClearPagerTemp()
}
