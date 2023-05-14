package paging

import "fmt"

const PAGESIZE = 4096

type Cell struct {
	key  uint32
	data []byte
}

func (c *Cell) Print() {
	fmt.Println("key:", c.key)
}

type Page struct {
	nodeType     NodeType
	isRoot       bool
	parent       uint32
	cells        []*Cell
	numCells     uint32
	data2        [PAGESIZE]byte
	currentIndex int
}

func NewPage() *Page {
	p := &Page{
		cells:        make([]*Cell, 0),
		numCells:     0,
		currentIndex: 0,
	}

	return p
}

func NewPageWithParams(nodeType NodeType, isRoot bool, parent uint32) *Page {
	p := &Page{
		nodeType:     nodeType,
		isRoot:       isRoot,
		parent:       parent,
		cells:        make([]*Cell, 0),
		numCells:     0,
		currentIndex: 0,
	}

	return p
}

func (p *Page) findIndexForKey(key uint32) uint32 {
	var leftIndex uint32 = 0
	var rightIndex uint32 = p.numCells
	currentIndex := rightIndex / 2

	for leftIndex < rightIndex {
		if p.cells[currentIndex].key < key {
			leftIndex = currentIndex + 1
		} else if p.cells[currentIndex].key > key {
			rightIndex = currentIndex
		} else {
			return currentIndex
		}

		currentIndex = (leftIndex + rightIndex) / 2
	}

	return currentIndex
}

func (p *Page) insertDataAtIndex(index uint32, key uint32, data []byte) {
	newCell := &Cell{
		key:  key,
		data: data,
	}

	if index < p.numCells {
		lastCell := p.cells[p.numCells-1]
		for i := p.numCells - 1; i > index; i-- {
			p.cells[i] = p.cells[i-1]
		}

		p.cells[index] = newCell

		p.cells = append(p.cells, lastCell)
	} else {
		p.cells = append(p.cells, newCell)
	}

	p.numCells++

	for _, c := range p.cells {
		c.Print()
	}
}

func (p *Page) hasSufficientSpace(newData []byte) bool {
	// if (len(p.data) + len(newData)) > cap(p.data) {
	// 	return false
	// } else {
	// 	return true
	// }

	if (p.currentIndex + len(newData)) >= PAGESIZE {
		return false
	} else {
		return true
	}
}

func (p *Page) appendBytes(newData []byte) {
	copy(p.data2[p.currentIndex:p.currentIndex+len(newData)], newData)
	p.currentIndex += len(newData)
}

func (p *Page) getRelevantLen() uint32 {
	return uint32(p.currentIndex)
}
