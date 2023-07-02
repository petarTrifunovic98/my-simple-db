package paging

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const MAX_PAGES_PER_TABLE uint32 = 100

type Pager struct {
	Pages             []IPage
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

	numPages := uint32(size / PAGE_SIZE)
	rootPage := uint32(0)
	tempBytes := make([]byte, PAGE_SIZE)
	if size > 20480 {
		file.ReadAt(tempBytes, 0)
		numPages = binary.LittleEndian.Uint32(tempBytes)
		rootPage = binary.LittleEndian.Uint32(tempBytes[4:])
	}
	fmt.Println("Num pages:", numPages)

	pager := &Pager{
		Pages:             make([]IPage, numPages, MAX_PAGES_PER_TABLE),
		File:              file,
		SizesWritten:      make([]uint32, 0),
		CurrentValueIndex: 0,
		NumPages:          uint32(numPages),
		RootPage:          rootPage,
	}

	return pager
}

func (p *Pager) getNextPageInd() uint32 {
	return p.NumPages
}

func (p *Pager) insertNewPage(page IPage, ind uint32) {
	if ind == p.NumPages {
		p.Pages = append(p.Pages, page)
	} else {
		p.Pages[ind] = page
	}
	p.NumPages++
}

func (p *Pager) findNodeToInsert(currentPageInd uint32, key []byte) uint32 {
	currentPage := p.GetPage(currentPageInd)
	if currentPage.getType() != LEAF_NODE {
		if !currentPage.hasSufficientSpace(uint16(4 + len(key))) {
			newPage := NewIPageWithParams(INTERNAL_NODE, false, 0, 0, 0)
			var parent IPage
			var parentInd uint32

			if currentPage.getIsRoot() {
				parent = NewIPageWithParams(INTERNAL_NODE, true, 0, 0, 0)
				parentInd = p.getNextPageInd()
				p.insertNewPage(parent, parentInd)
				p.RootPage = parentInd
			} else {
				parentInd = currentPage.getParent()
				parent = p.GetPage(parentInd)
			}

			newRightChildInd := p.getNextPageInd()
			p.insertNewPage(newPage, newRightChildInd)

			if currentPage.getIsRoot() {
				currentPage.transferCells(parentInd, currentPageInd, newRightChildInd, parent, newPage)
			} else {
				currentPage.transferCellsNotRoot(parentInd, currentPageInd, newRightChildInd, parent, newPage)
			}

			p.updateParentOfChildren(newRightChildInd)
			currentPage = parent
		}

		internalPage := currentPage.(*InternalPage)
		nextPageInd := internalPage.getPointer(internalPage.findIndexForKey(key))
		return p.findNodeToInsert(nextPageInd, key)
	} else {
		return currentPageInd
	}
}

func (p *Pager) AddNewData(key []byte, data []byte) {
	if p.NumPages == 0 {
		p.NumPages = 1
		p.Pages = append(p.Pages, NewIPageWithParams(LEAF_NODE, true, 0, 0, 0))
	}

	// root := p.GetPage(p.RootPage)
	// var pageToInsert *Page

	pageToInsertInd := p.findNodeToInsert(p.RootPage, key)
	pageToInsert := p.GetPage(pageToInsertInd)

	if !pageToInsert.hasSufficientSpace(uint16(len(data))) {
		/**
		 * This executes when root is full, in order to split it.
		 * Currently works only when root was leaf, and should now
		 * be split into two children with new root.
		 * TODO: Remove hard coded parts
		 */
		newPage := NewIPageWithParams(LEAF_NODE, false, 0, 0, 0)
		var parent IPage
		var parentInd uint32
		if pageToInsert.getIsRoot() {
			parent = NewIPageWithParams(INTERNAL_NODE, true, 0, 0, 0)
			parentInd = p.getNextPageInd()
			p.insertNewPage(parent, parentInd)
			p.RootPage = parentInd
		} else {
			parentInd = pageToInsert.getParent()
			parent = p.GetPage(parentInd)
		}

		newRightChildInd := p.getNextPageInd()
		p.insertNewPage(newPage, newRightChildInd)

		if pageToInsert.getIsRoot() {
			pageToInsert.transferCells(parentInd, pageToInsertInd, newRightChildInd, parent, newPage)
		} else {
			pageToInsert.transferCellsNotRoot(parentInd, pageToInsertInd, newRightChildInd, parent, newPage)
		}

		// The leftmost key in the right child decides which child the new key belongs to
		decisionKey := newPage.getKey(0)

		/**
		 * Compare the decision key and the key of new data,
		 * in order to decide which child gets the new element.
		 * Should be implemented as recursive search through nodes.
		 */
		compareResult := bytes.Compare(decisionKey, key)
		if compareResult == -1 {
			pageToInsert = newPage
		}
	}

	index := pageToInsert.findIndexForKey(key)
	leafPage := pageToInsert.(*LeafPage)
	leafPage.insertDataAtIndex(index, key, data)
}

func (p *Pager) ReadAllPages() []byte {
	/**
	 * Reads all the pages in a sorted order.
	 */
	values := make([]byte, 0, p.NumPages*PAGE_SIZE)
	p.ReadPageAtIndRec(p.RootPage, &values)
	return values
}

func (p *Pager) ReadPageAtIndRec(ind uint32, values *[]byte) {
	currentPage := p.GetPage(ind)
	if currentPage.getType() == LEAF_NODE {
		leafPage := currentPage.(*LeafPage)
		for i := 0; i < int(currentPage.getNumCells()); i++ {
			*values = append(*values, leafPage.getData(uint16(i))...)
		}
	} else {
		for i := 0; i <= int(currentPage.getNumCells()); i++ {
			internalPage := currentPage.(*InternalPage)
			p.ReadPageAtIndRec(internalPage.getPointer(uint16(i)), values)
		}
	}
}

// func (p *Pager) ReadPageAtInd(ind uint32) []byte {
// 	values := make([]byte, 0)
// 	// var relevantLen uint32 = 0

// 	currentPage := p.GetPage(ind)

// 	for i := 0; i < int(currentPage.getNumCells()); i++ {
// 		values = append(values, currentPage.getData(uint16(i))...)
// 	}

// 	// fmt.Println("values len:", relevantLen)
// 	// fmt.Println("num pages:", len(p.Pages))

// 	return values
// }

// func (p *Pager) ReadWholeCurrentPage() []byte {
// 	values := make([]byte, 0)
// 	// var relevantLen uint32 = 0

// 	var ind uint32
// 	for ind = 0; ind < p.NumPages; ind++ {
// 		currentPage := p.GetPage(ind)

// 		for i := 0; i < int(currentPage.getNumCells()); i++ {
// 			values = append(values, currentPage.getData(uint16(i))...)
// 		}

// 		//values2 = append(values2, currentPage.data2[:]...)
// 		// relevantLen += currentPage.getRelevantLen()
// 	}

// 	// fmt.Println("values len:", relevantLen)
// 	// fmt.Println("num pages:", len(p.Pages))

// 	return values
// }

func (p *Pager) GetPage(ind uint32) IPage {
	if ind < p.NumPages {
		if p.Pages[ind] == nil {
			tempBytes := make([]byte, PAGE_SIZE)
			p.File.ReadAt(tempBytes, int64((ind+1)*PAGE_SIZE))
			nodeHeader := &NodeHeader{}
			nodeHeader.Deserialize(tempBytes)
			nodeBodyBytes := tempBytes[NODE_HEADER_SIZE:]

			p.Pages[ind] = NewIPageWithParams(
				nodeHeader.nodeType,
				nodeHeader.isRoot,
				nodeHeader.parent,
				nodeHeader.numCells,
				nodeHeader.totalBodySize,
			)

			p.Pages[ind].setNodeBody(nodeBodyBytes)
		}
	} else {
		// newPages := make([]IPage, ind-p.NumPages+1)
		// p.Pages = append(p.Pages, newPages...)
		// p.Pages[ind] = NewPage()
		// p.NumPages++
		panic("Page index out of range!")
	}
	return p.Pages[ind]
}

func (p *Pager) ClearPager() {
	pagerMetadataBytesToWrite := make([]byte, PAGE_SIZE)
	pagerMetadataBytes := p.SerializeMetadata()
	copy(pagerMetadataBytesToWrite, pagerMetadataBytes)
	p.File.WriteAt(pagerMetadataBytes, 0)

	for ind, page := range p.Pages {
		if page != nil {
			pageBytes := make([]byte, PAGE_SIZE)
			nodeBytes := page.getHeader().Serialize()
			copy(pageBytes, nodeBytes)

			copy(pageBytes[NODE_HEADER_SIZE:], page.getBody())

			// n, _ := p.File.Write(pageBytes)
			n, _ := p.File.WriteAt(pageBytes, int64((ind+1)*PAGE_SIZE))
			fmt.Println("Written", n, "bytes for the page")
		}
	}

	p.File.Close()
}

func (p *Pager) updateParentOfChildren(newParentInd uint32) {
	page := p.GetPage(newParentInd)
	internalPage := page.(*InternalPage)
	for i := 0; i <= int(page.getNumCells()); i++ {
		childPage := p.GetPage(internalPage.getPointer(uint16(i)))
		childPage.setParent(newParentInd)
	}
}

func (p *Pager) PrintPages() {
	for ind, page := range p.Pages {
		if page == nil {
			page = p.GetPage(uint32(ind))
		}
		//page.Print()
		fmt.Println(page.getNumCells())
		fmt.Println("Implement page printing")
	}
}

func (p *Pager) SerializeMetadata() []byte {
	pagerMetadataBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(pagerMetadataBytes[0:4], p.NumPages)
	binary.LittleEndian.PutUint32(pagerMetadataBytes[4:8], p.RootPage)

	return pagerMetadataBytes
}
