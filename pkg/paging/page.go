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
	currentCellsSize uint32
}

func NewPage() *Page {
	p := &Page{
		cells: make([]*Cell, 0),
	}

	return p
}

// NEXT STEP: implement internal node creation
func NewPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint32) *Page {
	p := &Page{
		nodeHeader: &NodeHeader{
			nodeType: nodeType,
			isRoot:   isRoot,
			parent:   parent,
			numCells: numCells,
		},
		// TODO: make use of the numCells parameter for more efficient cell addition
		cells:            make([]*Cell, 0),
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

func (p *Page) transferCells(startIndSource int, destination *Page) {
	destination.cells = append(destination.cells, p.cells[startIndSource:]...)
	destination.nodeHeader.numCells = uint32(len(destination.cells))
	destination.calculateAndSetCurrentCellsSize()

	p.cells = p.cells[:startIndSource]
	p.nodeHeader.numCells = uint32(len(p.cells))
	p.calculateAndSetCurrentCellsSize()
}

func (p *Page) calculateAndSetCurrentCellsSize() {
	var size uint32 = 0
	for _, cell := range p.cells {
		size += KEY_SIZE + DATA_SIZE_SIZE + uint32(len(cell.data))
	}

	p.currentCellsSize = size
}

func (p *Page) getMaxKey() (bool, uint32) {
	if p.nodeHeader.numCells <= 0 {
		return false, 0
	} else {
		return true, p.cells[p.nodeHeader.numCells-1].key
	}
}

func (p *Page) hasSufficientSpaceTemp(newData []byte) bool {
	if p.currentCellsSize+uint32(len(newData))+KEY_SIZE+DATA_SIZE_SIZE >= PAGE_SIZE {
		return false
	} else {
		return true
	}
}

func (p *Page) Print() {
	p.nodeHeader.Print()
	for i := 0; uint32(i) < p.nodeHeader.numCells; i++ {
		cell := p.cells[i]
		fmt.Println("cell number", i, ": key =", cell.key)
	}
}
