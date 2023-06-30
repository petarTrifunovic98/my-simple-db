package paging

import (
	"bytes"
	"encoding/binary"
)

type LeafPage struct {
	PageBase
}

func (lp *LeafPage) getStartOfCells() uint16 {
	return lp.nodeHeader.numCells * OFFSET_SIZE
}

func (lp *LeafPage) getOffset(ind uint16) uint16 {
	return binary.LittleEndian.Uint16(lp.nodeBody[ind*OFFSET_SIZE:])
}

func (lp *LeafPage) getKey(ind uint16) []byte {
	cellStart := lp.nodeBody[lp.getStartOfCells()+lp.getOffset(ind):]
	return cellStart[:lp.nodeHeader.keySize]
}

func (lp *LeafPage) findIndexForKey(key []byte) uint16 {
	var leftIndex uint16 = 0
	var rightIndex uint16 = lp.nodeHeader.numCells
	currentIndex := rightIndex / 2

	for leftIndex < rightIndex {
		compareResult := bytes.Compare(lp.getKey(currentIndex), key)

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

func (lp *LeafPage) transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, dest *Page) {
	// find the offset and the key of the middle element
	middleElementOffset := lp.getOffset(lp.nodeHeader.numCells / 2)
	middleElementKey := lp.getKey(lp.nodeHeader.numCells / 2)

	// preserve old values from the "p" page
	oldStartOfCells := lp.getStartOfCells()
	oldTotalBodySize := lp.nodeHeader.totalBodySize

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	dest.nodeHeader.numCells = (lp.nodeHeader.numCells + 1) / 2
	lp.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.nodeHeader.totalBodySize =
		dest.nodeHeader.numCells*OFFSET_SIZE + (lp.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset)
	lp.nodeHeader.totalBodySize = lp.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	lp.nodeHeader.isRoot = false
	lp.nodeHeader.parent = newParentInd
	dest.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + lp.nodeHeader.keySize
	copy(newParent.nodeBody[pointerOffset+2*4+lp.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+lp.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++
	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < dest.nodeHeader.numCells; i++ {
		offset := lp.getOffset(lp.nodeHeader.numCells + i)
		offset -= middleElementOffset
		binary.LittleEndian.PutUint16(dest.nodeBody[i*OFFSET_SIZE:], offset)
	}

	/**
	 * Transfer data to new pages
	 */
	copy(dest.nodeBody[dest.nodeHeader.numCells*OFFSET_SIZE:],
		lp.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize])
	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(lp.nodeBody[lp.nodeHeader.numCells*OFFSET_SIZE:], lp.nodeBody[oldStartOfCells:lp.nodeHeader.totalBodySize])
}

func (lp *LeafPage) transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, dest *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := lp.getOffset(lp.nodeHeader.numCells / 2)
	middleElementKey := lp.getKey(lp.nodeHeader.numCells / 2)

	// preserve old values from the "p" page
	oldStartOfCells := lp.getStartOfCells()
	oldTotalBodySize := lp.nodeHeader.totalBodySize

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	newParent.nodeHeader.numCells++
	dest.nodeHeader.numCells = (lp.nodeHeader.numCells + 1) / 2
	lp.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	newParent.nodeHeader.totalBodySize += 2*4 + lp.nodeHeader.keySize
	dest.nodeHeader.totalBodySize =
		dest.nodeHeader.numCells*OFFSET_SIZE + (lp.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset)
	lp.nodeHeader.totalBodySize = lp.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	lp.nodeHeader.isRoot = false
	lp.nodeHeader.parent = newParentInd
	dest.nodeHeader.parent = newParentInd

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	binary.LittleEndian.PutUint32(newParent.nodeBody[:], oldChildInd)
	copy(newParent.nodeBody[4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[4+lp.nodeHeader.keySize:], newChildInd)
	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < dest.nodeHeader.numCells; i++ {
		offset := lp.getOffset(lp.nodeHeader.numCells + i)
		offset -= middleElementOffset
		binary.LittleEndian.PutUint16(dest.nodeBody[i*OFFSET_SIZE:], offset)
	}

	/**
	 * Transfer data to new pages
	 */
	copy(dest.nodeBody[dest.nodeHeader.numCells*OFFSET_SIZE:],
		lp.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize])
	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(lp.nodeBody[lp.nodeHeader.numCells*OFFSET_SIZE:], lp.nodeBody[oldStartOfCells:lp.nodeHeader.totalBodySize])
}
