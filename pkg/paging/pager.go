package paging

import (
	"encoding/binary"
	"fmt"
	"os"
)

const MAX_PAGES_PER_TABLE uint32 = 100

type Pager struct {
	Pages             []*Page
	File              *os.File
	SizesWritten      []uint32
	CurrentPageIndex  int32
	CurrentValueIndex uint32
	NumPages          uint32
	RootPage          uint32
}

func NewPager(filename string) *Pager {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	file.Chmod(0666)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	stat, _ := file.Stat()
	size := stat.Size()

	numPages := size / PAGE_SIZE
	if size%PAGE_SIZE != 0 {
		numPages++
	}
	fmt.Println("Num pages:", numPages)

	pager := &Pager{
		Pages:             make([]*Page, numPages, MAX_PAGES_PER_TABLE),
		CurrentPageIndex:  int32(numPages) - 1,
		File:              file,
		SizesWritten:      make([]uint32, 0),
		CurrentValueIndex: 0,
		NumPages:          uint32(numPages),
		RootPage:          0,
	}

	return pager
}

func (p *Pager) AddNewData(key uint32, data []byte) {
	if p.NumPages == 0 {
		p.NumPages = 1
		p.CurrentPageIndex = 0
		p.Pages = append(p.Pages, NewPageWithParams(LEAF_NODE, true, 0, 0))
	}

	root := p.GetPageTemp(p.RootPage)
	index := root.findIndexForKey(key)
	//if root.cells[index].key != key {
	root.insertDataAtIndex(index, key, data)
	//}
}

func (p *Pager) AddToCurrentPage(key uint32, data []byte) {
	if p.NumPages == 0 {
		p.NumPages = 1
		p.CurrentPageIndex = 0
		p.Pages = append(p.Pages, NewPage())
	}

	currentPage := p.GetPage(uint32(p.CurrentPageIndex))

	if currentPage.hasSufficientSpaceTemp(data) {
		currentPage.appendBytes(data)
	} else {
		p.Pages = append(p.Pages, NewPage())
		p.CurrentPageIndex++
		if p.CurrentPageIndex >= int32(p.NumPages) {
			p.NumPages = uint32(p.CurrentPageIndex) + 1
		}

		currentPage = p.Pages[p.CurrentPageIndex]
		currentPage.appendBytes(data)
	}
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
	values2 := make([]byte, 0)
	// var relevantLen uint32 = 0

	var ind uint32
	for ind = 0; ind < p.NumPages; ind++ {
		currentPage := p.GetPage(ind)
		values2 = append(values2, currentPage.data2[:]...)
		// relevantLen += currentPage.getRelevantLen()
	}

	// fmt.Println("values len:", relevantLen)
	// fmt.Println("num pages:", len(p.Pages))

	return values2
}

func (p *Pager) ReadWholeCurrentPageTemp() []byte {
	values2 := make([]byte, 0)
	// var relevantLen uint32 = 0

	var ind uint32
	for ind = 0; ind < p.NumPages; ind++ {
		currentPage := p.GetPageTemp(ind)
		for _, cell := range currentPage.cells {
			values2 = append(values2, cell.data...)
		}

		//values2 = append(values2, currentPage.data2[:]...)
		// relevantLen += currentPage.getRelevantLen()
	}

	// fmt.Println("values len:", relevantLen)
	// fmt.Println("num pages:", len(p.Pages))

	return values2
}

func (p *Pager) GetPage(ind uint32) *Page {
	if ind < p.NumPages {
		if p.Pages[ind] == nil {
			p.Pages[ind] = NewPage()
			tempBytes := make([]byte, PAGE_SIZE)
			read, _ := p.File.ReadAt(tempBytes, int64(PAGE_SIZE*ind))
			p.Pages[ind].appendBytes(tempBytes[:read])
		}
	} else {
		newPages := make([]*Page, ind-p.NumPages+1)
		p.Pages = append(p.Pages, newPages...)
		p.Pages[ind] = NewPage()
		p.NumPages++
	}
	return p.Pages[ind]
}

func (p *Pager) GetPageTemp(ind uint32) *Page {
	if ind < p.NumPages {
		if p.Pages[ind] == nil {
			tempBytes := make([]byte, PAGE_SIZE)
			p.File.ReadAt(tempBytes, int64(PAGE_SIZE*ind))
			nodeHeader := &NodeHeader{}
			nodeHeader.Deserialize(tempBytes)
			cellBytes := tempBytes[NODE_HEADER_SIZE:]
			currentIndex := 0

			p.Pages[ind] = NewPageWithParams(LEAF_NODE, true, nodeHeader.parent, nodeHeader.numCells)

			for i := 0; uint32(i) < nodeHeader.numCells; i++ {
				key := binary.LittleEndian.Uint32(cellBytes[currentIndex : currentIndex+4])
				currentIndex += 4
				dataSize := binary.LittleEndian.Uint32(cellBytes[currentIndex : currentIndex+4])
				currentIndex += 4
				data := cellBytes[currentIndex : currentIndex+int(dataSize)]
				currentIndex += int(dataSize)
				p.Pages[ind].cells = append(p.Pages[ind].cells, &Cell{
					data:     data,
					dataSize: dataSize,
					key:      key,
				})
			}
		}
	} else {
		newPages := make([]*Page, ind-p.NumPages+1)
		p.Pages = append(p.Pages, newPages...)
		p.Pages[ind] = NewPage()
		p.NumPages++
	}
	return p.Pages[ind]
}

func (p *Pager) ClearPager() {
	for ind, page := range p.Pages {
		if ind < len(p.Pages)-1 {
			p.File.Write(page.data2[:])
		} else {
			p.File.Write(page.data2[:page.currentIndex])
		}
	}

	p.File.Close()
}

func (p *Pager) ClearPagerTemp() {
	for _, page := range p.Pages {
		pageBytes := make([]byte, 0, PAGE_SIZE)
		nodeBytes := page.nodeHeader.Serialize()
		pageBytes = append(pageBytes, nodeBytes...)

		for _, cell := range page.cells {
			keyBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(keyBytes, cell.key)
			pageBytes = append(pageBytes, keyBytes...)

			dataSizeBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(dataSizeBytes, cell.dataSize)
			pageBytes = append(pageBytes, dataSizeBytes...)

			pageBytes = append(pageBytes, cell.data...)
		}

		additionalBytes := make([]byte, PAGE_SIZE-len(pageBytes))
		pageBytes = append(pageBytes, additionalBytes...)

		n, _ := p.File.Write(pageBytes)
		fmt.Println("Written", n, "bytes for the page")
	}

	p.File.Close()
}
