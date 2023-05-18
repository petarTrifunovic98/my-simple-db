package paging

import "fmt"

const PAGE_SIZE = 4096
const KEY_SIZE uint32 = 4
const DATA_SIZE_SIZE uint32 = 4

type Cell struct {
	data     []byte
	dataSize uint32
	key      uint32
}

func (c *Cell) Print() {
	fmt.Println("key:", c.key)
}

type Page struct {
	nodeHeader       *NodeHeader
	cells            []*Cell
	data2            [PAGE_SIZE]byte
	currentIndex     int
	currentCellsSize uint32
}

func NewPage() *Page {
	p := &Page{
		cells:        make([]*Cell, 0),
		currentIndex: 0,
	}

	return p
}

func NewPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint32) *Page {
	p := &Page{
		nodeHeader: &NodeHeader{
			nodeType: nodeType,
			isRoot:   isRoot,
			parent:   parent,
			numCells: numCells,
		},
		cells:            make([]*Cell, 0),
		currentIndex:     0,
		currentCellsSize: 0,
	}

	return p
}

func (p *Page) findIndexForKey(key uint32) uint32 {
	var leftIndex uint32 = 0
	var rightIndex uint32 = p.nodeHeader.numCells
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
		key:      key,
		dataSize: uint32(len(data)),
		data:     data,
	}

	if index < p.nodeHeader.numCells {
		lastCell := p.cells[p.nodeHeader.numCells-1]
		for i := p.nodeHeader.numCells - 1; i > index; i-- {
			p.cells[i] = p.cells[i-1]
		}

		p.cells[index] = newCell

		p.cells = append(p.cells, lastCell)
	} else {
		p.cells = append(p.cells, newCell)
	}

	p.nodeHeader.numCells++
	p.currentCellsSize += KEY_SIZE + DATA_SIZE_SIZE + uint32(len(data))

	for _, c := range p.cells {
		c.Print()
	}
}

func (p *Page) hasSufficientSpace(newData []byte) bool {
	if (p.currentIndex + len(newData)) >= PAGE_SIZE {
		return false
	} else {
		return true
	}
}

func (p *Page) hasSufficientSpaceTemp(newData []byte) bool {
	if p.currentCellsSize+uint32(len(newData))+KEY_SIZE+DATA_SIZE_SIZE >= PAGE_SIZE {
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
