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
	// p.File.Seek(0, 2) -- Works without this line. Maybe Write always writes to the end of the file?
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

func (p *Pager) ReadWholeCurrentPage() []byte {
	p.File.Sync()
	p.File.Seek(0, 0)

	stat, _ := p.File.Stat()
	s := stat.Size()

	values := make([]byte, s)

	p.File.Read(values)

	return values
}

func (p *Pager) ClearPager() {
	p.File.Close()
}
