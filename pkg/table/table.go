package table

import (
	"bytes"

	"github.com/petarTrifunovic98/my-simple-db/pkg/paging"
)

const PAGE_SIZE uint32 = 4096
const MAX_PAGES_PER_TABLE uint32 = 100

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
		// Pages:   make([]*Page, 0, MAX_PAGES_PER_TABLE),
	}

	return table
}

func (t *Table) Insert(key any, data any) {
	t.Pager.AddToCurrentPage(data)
	// serialization.Serialize(&data, &(t.Pages[0].SerializedRows))
}

func (t *Table) Select() [][]byte {
	values := t.Pager.ReadWholeCurrentPage()
	return values
}

func (t *Table) DestroyTable() {
	t.Pager.ClearPager()
}
