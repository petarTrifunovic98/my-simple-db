package paging

import (
	"bytes"
	"encoding/binary"
)

const PAGE_SIZE = 4096
const KEY_SIZE uint16 = 4
const DATA_SIZE_SIZE uint16 = 2

/**
 * Leaf node body outline:
 * - first a list of offsets; each offset is 2 bytes; each value represents
 *	the offset from the beginning of the list of cells
 * - after offset list, a list of cells; each cell consists of a key of size
 * which is recorded in the node header, a 2 byte value which represents the
 * size of data in bytes, and the actual data
 * |    offset list		|                            cells list                                |
 * |--------------------|----------------------------------------------------------------------|
 * |number of cells * 2B|key (keysize*1B), data size (DATA_SIZE_SIZE*1b), data (data size * 1B)|
 */

/**
 * Internal node body outline:
 * - a list of pointers and keys
 * - first, there is a pointer to the leftmost child
 * - after that, a list of cells is actually a list of key-pointer pairs
 * - numCells in the header actually says what the number of keys is
 * - number of pointers is always numCells+1
 * - each key is of keysize size
 * - each child pointer is 32B
 * | leftmost child pointer | key | child pointer | key | child pointer |...
 */

const OFFSET_SIZE = 2

type Page struct {
	nodeHeader NodeHeader
	nodeBody   [PAGE_SIZE - NODE_HEADER_SIZE]byte
}

func NewPage() *Page {
	p := &Page{}

	return p
}

// NEXT STEP: implement internal node creation
func NewPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint16, totalBodySize uint16) *Page {
	var p IPage
	
	if nodeType == LEAF_NODE {
		p = &LeafPage{
			PageBase: PageBase{
				nodeHeader: NodeHeader{
					nodeType:      nodeType,
					isRoot:        isRoot,
					parent:        parent,
					numCells:      numCells,
					totalBodySize: totalBodySize,
					keySize:       KEY_SIZE,
				},
			},
		}
	} else {
		p = &
	}

	p := &Page{
		nodeHeader: NodeHeader{
			nodeType:      nodeType,
			isRoot:        isRoot,
			parent:        parent,
			numCells:      numCells,
			totalBodySize: totalBodySize,
			keySize:       KEY_SIZE,
		},
	}

	return p
}

func (p *Page) getOffset(ind uint16) uint16 {
	return binary.LittleEndian.Uint16(p.nodeBody[ind*OFFSET_SIZE:])
}

func (p *Page) getStartOfCells() uint16 {
	return p.nodeHeader.numCells * OFFSET_SIZE
}

func (p *Page) getKey(ind uint16) []byte {
	cellStart := p.nodeBody[p.getStartOfCells()+p.getOffset(ind):]
	return cellStart[:p.nodeHeader.keySize]
}

func (p *Page) getData(ind uint16) []byte {
	cellStart := p.nodeBody[p.getStartOfCells()+p.getOffset(ind):]
	dataSize := binary.LittleEndian.Uint16(cellStart[p.nodeHeader.keySize:])
	return cellStart[p.nodeHeader.keySize+DATA_SIZE_SIZE : p.nodeHeader.keySize+DATA_SIZE_SIZE+dataSize]
}

// Change to a less hard-coded variant
func (p *Page) getOffsetInternal(ind uint16) uint16 {
	return ind * (4 + p.nodeHeader.keySize)
}

func (p *Page) getKeyInternal(ind uint16) []byte {
	return p.nodeBody[4+ind*(4+p.nodeHeader.keySize) : 4+ind*(4+p.nodeHeader.keySize)+p.nodeHeader.keySize]
}

func (p *Page) getPointerInternal(ind uint16) uint32 {
	pointerBytes := p.nodeBody[ind*(4+p.nodeHeader.keySize) : ind*(4+p.nodeHeader.keySize)+4]
	return binary.LittleEndian.Uint32(pointerBytes)
}

func (p *Page) findIndexForKeyInternal(key []byte) uint16 {
	var leftIndex uint16 = 0
	var rightIndex uint16 = p.nodeHeader.numCells
	currentIndex := rightIndex / 2

	for leftIndex < rightIndex {
		compareResult := bytes.Compare(p.getKeyInternal(currentIndex), key)

		if compareResult == -1 {
			leftIndex = currentIndex + 1
		} else if compareResult == 1 {
			rightIndex = currentIndex
		} else {
			return currentIndex
		}

		currentIndex = (leftIndex + rightIndex) / 2
	}

	return currentIndex
}

func (p *Page) findIndexForKey(key []byte) uint16 {
	var leftIndex uint16 = 0
	var rightIndex uint16 = p.nodeHeader.numCells
	currentIndex := rightIndex / 2

	for leftIndex < rightIndex {
		compareResult := bytes.Compare(p.getKey(currentIndex), key)

		if compareResult == -1 {
			leftIndex = currentIndex + 1
		} else if compareResult == 1 {
			rightIndex = currentIndex
		} else {
			return currentIndex
		}

		currentIndex = (leftIndex + rightIndex) / 2
	}

	return currentIndex
}

func (p *Page) insertDataAtIndex(ind uint16, key []byte, data []byte) {
	startOfCells := p.getStartOfCells()
	keySize := p.nodeHeader.keySize
	totalBodySize := p.nodeHeader.totalBodySize
	dataLen16 := uint16(len(data))
	lenIncrease := keySize + 2 + dataLen16

	offsets := make([]byte /*0,*/, (p.nodeHeader.numCells+1)*OFFSET_SIZE)
	copy(offsets, p.nodeBody[:startOfCells])
	cells := make([]byte /*0,*/, (totalBodySize-startOfCells)+lenIncrease)
	copy(cells, p.nodeBody[startOfCells:totalBodySize])

	/**
	 * Update offsets list
	 */
	if ind < p.nodeHeader.numCells {
		/**
		 * Insert a new cell among the existing ones.
		 */
		nthOffset := binary.LittleEndian.Uint16(offsets[ind*OFFSET_SIZE:])
		// make room for the new cell by shifting a part of the existing ones to the right
		copy(cells[nthOffset+lenIncrease:], cells[nthOffset:totalBodySize-startOfCells])
		// insert the cell key
		copy(cells[nthOffset:nthOffset+keySize], key)
		// insert the cell data size
		dataLen16Bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(dataLen16Bytes, dataLen16)
		copy(cells[nthOffset+keySize:nthOffset+keySize+2], dataLen16Bytes)
		// insert the cell data
		copy(cells[nthOffset+keySize+2:nthOffset+keySize+2+dataLen16], data)

		// Shift the necessary offsets to the right in the offsets list
		for i := p.nodeHeader.numCells - 1; int16(i) >= int16(ind); i-- {
			// Get the old offset at index i
			oldOffset := binary.LittleEndian.Uint16(offsets[i*OFFSET_SIZE:])
			// Update the old index by adding new cell size
			newOffset := oldOffset + uint16(lenIncrease)
			newOffsetBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(newOffsetBytes, newOffset)
			// Insert the new offset and immediately shift it to the right
			copy(offsets[(i+1)*OFFSET_SIZE:(i+2)*OFFSET_SIZE], newOffsetBytes)
		}
	} else {
		newOffsetBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(newOffsetBytes, totalBodySize-startOfCells)
		copy(offsets[p.nodeHeader.numCells*OFFSET_SIZE:], newOffsetBytes)
		copy(cells[totalBodySize-startOfCells:], key)
		dataLen16Bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(dataLen16Bytes, dataLen16)
		copy(cells[totalBodySize-startOfCells+keySize:], dataLen16Bytes)
		copy(cells[totalBodySize-startOfCells+keySize+2:], data)
	}

	p.nodeHeader.numCells++
	p.nodeHeader.totalBodySize += lenIncrease + OFFSET_SIZE
	copy(p.nodeBody[:], offsets)
	copy(p.nodeBody[p.nodeHeader.numCells*OFFSET_SIZE:], cells)

}

func (p *Page) updateParent(newParentInd uint32) {
	p.nodeHeader.parent = newParentInd
}

func (p *Page) transferCellsInternalNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := p.getOffsetInternal((p.nodeHeader.numCells - 1) / 2)
	middleElementKey := p.getKeyInternal((p.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	destination.nodeHeader.numCells = p.nodeHeader.numCells / 2
	p.nodeHeader.numCells = (p.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	destination.nodeHeader.totalBodySize =
		destination.nodeHeader.numCells*(4+destination.nodeHeader.keySize) + 4
	p.nodeHeader.totalBodySize = p.nodeHeader.numCells*(4+destination.nodeHeader.keySize) + 4

	p.nodeHeader.isRoot = false
	p.nodeHeader.parent = newParentInd
	destination.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + p.nodeHeader.keySize
	copy(newParent.nodeBody[pointerOffset+2*4+p.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset:], oldChildInd)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+p.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++

	/**
	 * Transfer data to new pages
	 */
	copy(destination.nodeBody[:],
		p.nodeBody[middleElementOffset+4+p.nodeHeader.keySize:])
}

/**
 * Currently:
 *  - splits the page "p", sends the middle element to "newParent"
 *  and sends all elements greater than the middle one to "destination"
 *  - works only for splitting a leaf node which was previously root, thus
 *  making the "newParent" page the node root
 * Should be updated to:
 *  - work with any leaf node
 */
func (p *Page) transferCellsInternal(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := p.getOffsetInternal((p.nodeHeader.numCells - 1) / 2)
	middleElementKey := p.getKeyInternal((p.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	destination.nodeHeader.numCells = p.nodeHeader.numCells / 2
	p.nodeHeader.numCells = (p.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	destination.nodeHeader.totalBodySize =
		destination.nodeHeader.numCells*(4+destination.nodeHeader.keySize) + 4
	p.nodeHeader.totalBodySize = p.nodeHeader.numCells*(4+destination.nodeHeader.keySize) + 4

	p.nodeHeader.isRoot = false
	p.nodeHeader.parent = newParentInd
	destination.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + p.nodeHeader.keySize + 4
	copy(newParent.nodeBody[pointerOffset+2*4+p.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset:], oldChildInd)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+p.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++

	/**
	 * Transfer data to new pages
	 */
	copy(destination.nodeBody[:],
		p.nodeBody[middleElementOffset+4+p.nodeHeader.keySize:])
}

/**
 * Currently:
 *  - splits the page "p", sends the middle element to "newParent"
 *  and sends all elements greater than the middle one to "destination"
 *  - works only for splitting a leaf node which was previously root, thus
 *  making the "newParent" page the node root
 * Should be updated to:
 *  - work with any leaf node
 */
func (p *Page) transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := p.getOffset(p.nodeHeader.numCells / 2)
	middleElementKey := p.getKey(p.nodeHeader.numCells / 2)

	// preserve old values from the "p" page
	oldStartOfCells := p.getStartOfCells()
	oldTotalBodySize := p.nodeHeader.totalBodySize

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	destination.nodeHeader.numCells = (p.nodeHeader.numCells + 1) / 2
	p.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	destination.nodeHeader.totalBodySize =
		destination.nodeHeader.numCells*OFFSET_SIZE + (p.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset)
	p.nodeHeader.totalBodySize = p.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	p.nodeHeader.isRoot = false
	p.nodeHeader.parent = newParentInd
	destination.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + p.nodeHeader.keySize
	copy(newParent.nodeBody[pointerOffset+2*4+p.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+p.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++
	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < destination.nodeHeader.numCells; i++ {
		offset := p.getOffset(p.nodeHeader.numCells + i)
		offset -= middleElementOffset
		binary.LittleEndian.PutUint16(destination.nodeBody[i*OFFSET_SIZE:], offset)
	}

	/**
	 * Transfer data to new pages
	 */
	copy(destination.nodeBody[destination.nodeHeader.numCells*OFFSET_SIZE:],
		p.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize])
	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(p.nodeBody[p.nodeHeader.numCells*OFFSET_SIZE:], p.nodeBody[oldStartOfCells:p.nodeHeader.totalBodySize])
}

/**
 * Currently:
 *  - splits the page "p", sends the middle element to "newParent"
 *  and sends all elements greater than the middle one to "destination"
 *  - works only for splitting a leaf node which was previously root, thus
 *  making the "newParent" page the node root
 * Should be updated to:
 *  - work with any leaf node
 */
func (p *Page) transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := p.getOffset(p.nodeHeader.numCells / 2)
	middleElementKey := p.getKey(p.nodeHeader.numCells / 2)

	// preserve old values from the "p" page
	oldStartOfCells := p.getStartOfCells()
	oldTotalBodySize := p.nodeHeader.totalBodySize

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	newParent.nodeHeader.numCells++
	destination.nodeHeader.numCells = (p.nodeHeader.numCells + 1) / 2
	p.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	newParent.nodeHeader.totalBodySize += 2*4 + p.nodeHeader.keySize
	destination.nodeHeader.totalBodySize =
		destination.nodeHeader.numCells*OFFSET_SIZE + (p.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset)
	p.nodeHeader.totalBodySize = p.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	p.nodeHeader.isRoot = false
	p.nodeHeader.parent = newParentInd
	destination.nodeHeader.parent = newParentInd

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	binary.LittleEndian.PutUint32(newParent.nodeBody[:], oldChildInd)
	copy(newParent.nodeBody[4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[4+p.nodeHeader.keySize:], newChildInd)
	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < destination.nodeHeader.numCells; i++ {
		offset := p.getOffset(p.nodeHeader.numCells + i)
		offset -= middleElementOffset
		binary.LittleEndian.PutUint16(destination.nodeBody[i*OFFSET_SIZE:], offset)
	}

	/**
	 * Transfer data to new pages
	 */
	copy(destination.nodeBody[destination.nodeHeader.numCells*OFFSET_SIZE:],
		p.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize])
	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(p.nodeBody[p.nodeHeader.numCells*OFFSET_SIZE:], p.nodeBody[oldStartOfCells:p.nodeHeader.totalBodySize])
}

func (p *Page) hasSufficientSpace(newData []byte) bool {
	// TODO: check if there is enough space for new data
	oldSize := NODE_HEADER_SIZE + p.nodeHeader.totalBodySize
	newSize := oldSize + uint16(len(newData)) + p.nodeHeader.keySize + DATA_SIZE_SIZE
	return newSize <= PAGE_SIZE
}

func (p *Page) hasSufficientSpaceInternal() bool {
	// oldSize := NODE_HEADER_SIZE + p.nodeHeader.totalBodySize
	// newSize := oldSize + 4 + p.nodeHeader.keySize
	// return newSize <= PAGE_SIZE

	// For testing purposes. The commented code above is the correct code.
	return p.nodeHeader.numCells < 3
}

func (p *Page) Print() {
	p.nodeHeader.Print()
	//	for i := 0; uint16(i) < p.nodeHeader.numCells; i++ {
	//		cell := p.cells[i]
	//		fmt.Println("cell number", i, ": key =", cell.key)
	//	}
}
