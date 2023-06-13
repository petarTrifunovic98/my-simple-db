package paging

import (
	"fmt"
	"os"
)

const MAX_PAGES_PER_TABLE uint32 = 100

type Pager struct {
	Pages             []*Page
	File              *os.File
	SizesWritten      []uint32
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
		File:              file,
		SizesWritten:      make([]uint32, 0),
		CurrentValueIndex: 0,
		NumPages:          uint32(numPages),
		RootPage:          0,
	}

	return pager
}
func (p *Pager) AddNewData2(key []byte, data []byte) {
	if p.NumPages == 0 {
		p.NumPages = 1
		p.Pages = append(p.Pages, NewPageWithParams(LEAF_NODE, true, 0, 0, 0))
	}

	root := p.GetPage2(p.RootPage)
	var pageToInsert *Page

	if root.hasSufficientSpace2(data) {
		pageToInsert = root
	} else {
		// NEXT STEP: create a new root node
		// newPage := NewPageWithParams(LEAF_NODE, false, 0, 0)

		// // TODO: implement usage of old page index; don't just append in any case
		// p.Pages = append(p.Pages, newPage)
		// p.NumPages++

		// root.transferCells(int(root.nodeHeader.numCells)/2, newPage)

		// if _, leftNodeMaxKey := root.getMaxKey(); key <= leftNodeMaxKey {
		// 	pageToInsert = root
		// } else {
		// 	pageToInsert = newPage
		// }
	}

	index := pageToInsert.findIndexForKey2(key)
	pageToInsert.insertDataAtIndex2(index, key, data)
}

func (p *Pager) ReadWholeCurrentPage2() []byte {
	values2 := make([]byte, 0)
	// var relevantLen uint32 = 0

	var ind uint32
	for ind = 0; ind < p.NumPages; ind++ {
		currentPage := p.GetPage2(ind)

		for i := 0; i < int(currentPage.nodeHeader.numCells); i++ {
			values2 = append(values2, currentPage.getData(uint16(i))...)
		}

		//values2 = append(values2, currentPage.data2[:]...)
		// relevantLen += currentPage.getRelevantLen()
	}

	// fmt.Println("values len:", relevantLen)
	// fmt.Println("num pages:", len(p.Pages))

	return values2
}

func (p *Pager) GetPage2(ind uint32) *Page {
	if ind < p.NumPages {
		if p.Pages[ind] == nil {
			tempBytes := make([]byte, PAGE_SIZE)
			p.File.ReadAt(tempBytes, int64(PAGE_SIZE*ind))
			nodeHeader := &NodeHeader{}
			nodeHeader.Deserialize(tempBytes)
			nodeBodyBytes := tempBytes[NODE_HEADER_SIZE:]

			p.Pages[ind] = NewPageWithParams(LEAF_NODE, true, nodeHeader.parent, nodeHeader.numCells, nodeHeader.totalBodySize)

			copy(p.Pages[ind].nodeBody[:], nodeBodyBytes)
		}
	} else {
		newPages := make([]*Page, ind-p.NumPages+1)
		p.Pages = append(p.Pages, newPages...)
		p.Pages[ind] = NewPage()
		p.NumPages++
	}
	return p.Pages[ind]
}

func (p *Pager) ClearPager2() {
	for _, page := range p.Pages {
		pageBytes := make([]byte, PAGE_SIZE)
		nodeBytes := page.nodeHeader.Serialize()
		copy(pageBytes, nodeBytes)

		copy(pageBytes[NODE_HEADER_SIZE:], page.nodeBody[:])

		n, _ := p.File.Write(pageBytes)
		fmt.Println("Written", n, "bytes for the page")
	}

	p.File.Close()
}

func (p *Pager) PrintPages() {
	for ind, page := range p.Pages {
		if page == nil {
			page = p.GetPage2(uint32(ind))
		}
		page.Print()
	}
}
