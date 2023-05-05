package paging

import (
	"fmt"
	"io"
	"os"
)

const MAX_PAGES_PER_TABLE uint32 = 100

type Page struct {
	SerializedRows io.ReadWriter
}

type Pager struct {
	Pages             []*Page
	CurrentPage       uint32
	File              *os.File
	SizesWritten      []uint32
	CurrentValueIndex uint32
}

func NewPager(filename string) *Pager {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	file.Chmod(0666)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	pager := &Pager{
		Pages:             make([]*Page, 0, MAX_PAGES_PER_TABLE),
		File:              file,
		SizesWritten:      make([]uint32, 0),
		CurrentValueIndex: 0,
	}

	pager.Pages = append(pager.Pages, &Page{
		// SerializedRows: &bytes.Buffer{},
		SerializedRows: file,
	})
	pager.CurrentPage = 0

	return pager
}

func (p *Pager) AddToCurrentPage(data any) {
	// p.File.Seek(0, 2)
	// serialization.Serialize(&data, p.File)
	num, _ := p.File.Write(data.([]byte))
	p.SizesWritten = append(p.SizesWritten, uint32(num))
}

func (p *Pager) ReadNextValue() []byte {
	p.CurrentValueIndex++
	if p.CurrentValueIndex >= uint32(len(p.SizesWritten)) {
		return nil
	}
	size := p.SizesWritten[p.CurrentValueIndex]
	b := make([]byte, size)
	p.File.Read(b)
	return b
}

func (p *Pager) ReadWholeCurrentPage() [][]byte {
	values := make([][]byte, 0)

	p.File.Sync()
	p.File.Seek(0, 0)

	fmt.Println(p.SizesWritten)

	for _, size := range p.SizesWritten {
		b := make([]byte, size)
		p.File.Read(b)
		fmt.Println(len(b))
		values = append(values, b)
	}

	return values
}

func (p *Pager) ClearPager() {
	p.File.Close()
}
