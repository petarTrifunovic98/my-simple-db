package paging

import (
	"bytes"
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
func (p *Pager) AddNewData(key []byte, data []byte) {
	if p.NumPages == 0 {
		p.NumPages = 1
		p.Pages = append(p.Pages, NewPageWithParams(LEAF_NODE, true, 0, 0, 0))
	}

	root := p.GetPage(p.RootPage)
	var pageToInsert *Page

	if root.hasSufficientSpace(data) {
		pageToInsert = root
	} else {
		/**
		 * This executes when root is full, in order to split it.
		 * Currently works only when root was leaf, and should now
		 * be split into two children with new root.
		 * Parts are hard coded, such as the index of RootPage,
		 * or the parameter "2" in the transferCells function.
		 * TODO: Remove hard coded parts
		 */
		newPage := NewPageWithParams(LEAF_NODE, false, 0, 0, 0)
		newRoot := NewPageWithParams(INTERNAL_NODE, true, 0, 0, 0)

		p.Pages = append(p.Pages, newPage, newRoot)
		p.NumPages += 2
		// Make sure that the root is always th first page, for easier persistance to disk
		p.Pages[0] = newRoot
		p.Pages[2] = root
		root.transferCells(0, 2, 1, newRoot, newPage)

		// Currently, new root can have only one value, created when splitting the old root
		rootKey := newRoot.getKeyInternal(0)

		/**
		 * Compare the root key and the key of new data,
		 * in order to decide which child gets the new element.
		 * Should be implemented as recursive search through nodes.
		 */
		compareResult := bytes.Compare(rootKey, key)
		if compareResult == -1 {
			pageToInsert = newPage
		} else {
			pageToInsert = root
		}
	}

	index := pageToInsert.findIndexForKey(key)
	pageToInsert.insertDataAtIndex(index, key, data)
}

func (p *Pager) ReadAllPages() []byte {
	/**
	 * Reads all the pages in a sorted order.
	 * Sorting is currently hard coded and works only
	 * for a tree with a root node which has only
	 * two children.
	 */
	if p.NumPages > 1 {
		values := p.ReadPageAtInd(2)
		fmt.Println("Reading from more than one page...")
		values = append(values, p.ReadPageAtInd(1)...)
		fmt.Println("Values len:", len(values))
		return values
	} else {
		return p.ReadPageAtInd(0)
	}

}

func (p *Pager) ReadPageAtInd(ind uint32) []byte {
	values := make([]byte, 0)
	// var relevantLen uint32 = 0

	currentPage := p.GetPage(ind)

	for i := 0; i < int(currentPage.nodeHeader.numCells); i++ {
		values = append(values, currentPage.getData(uint16(i))...)
	}

	// fmt.Println("values len:", relevantLen)
	// fmt.Println("num pages:", len(p.Pages))

	return values
}

func (p *Pager) ReadWholeCurrentPage() []byte {
	values := make([]byte, 0)
	// var relevantLen uint32 = 0

	var ind uint32
	for ind = 0; ind < p.NumPages; ind++ {
		currentPage := p.GetPage(ind)

		for i := 0; i < int(currentPage.nodeHeader.numCells); i++ {
			values = append(values, currentPage.getData(uint16(i))...)
		}

		//values2 = append(values2, currentPage.data2[:]...)
		// relevantLen += currentPage.getRelevantLen()
	}

	// fmt.Println("values len:", relevantLen)
	// fmt.Println("num pages:", len(p.Pages))

	return values
}

func (p *Pager) GetPage(ind uint32) *Page {
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

func (p *Pager) ClearPager() {
	for _, page := range p.Pages {
		if page != nil {
			pageBytes := make([]byte, PAGE_SIZE)
			nodeBytes := page.nodeHeader.Serialize()
			copy(pageBytes, nodeBytes)

			copy(pageBytes[NODE_HEADER_SIZE:], page.nodeBody[:])

			n, _ := p.File.Write(pageBytes)
			fmt.Println("Written", n, "bytes for the page")
		}
	}

	p.File.Close()
}

func (p *Pager) PrintPages() {
	for ind, page := range p.Pages {
		if page == nil {
			page = p.GetPage(uint32(ind))
		}
		page.Print()
	}
}
