package paging

import (
	"bytes"
	"encoding/binary"
)

type InternalPage struct {
	PageBase
}

func (ip *InternalPage) getOffset(ind uint16) uint16 {
	return ind * (4 + ip.nodeHeader.keySize)
}

func (ip *InternalPage) getKey(ind uint16) []byte {
	return ip.nodeBody[4+ind*(4+ip.nodeHeader.keySize) : 4+ind*(4+ip.nodeHeader.keySize)+ip.nodeHeader.keySize]
}

func (ip *InternalPage) findIndexForKey(key []byte) uint16 {
	var leftIndex uint16 = 0
	var rightIndex uint16 = ip.nodeHeader.numCells
	currentIndex := rightIndex / 2

	for leftIndex < rightIndex {
		compareResult := bytes.Compare(ip.getKey(currentIndex), key)

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

func (ip *InternalPage) transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, dest *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := ip.getOffset((ip.nodeHeader.numCells - 1) / 2)
	middleElementKey := ip.getKey((ip.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	dest.nodeHeader.numCells = ip.nodeHeader.numCells / 2
	ip.nodeHeader.numCells = (ip.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.nodeHeader.totalBodySize =
		dest.nodeHeader.numCells*(4+dest.nodeHeader.keySize) + 4
	ip.nodeHeader.totalBodySize = ip.nodeHeader.numCells*(4+dest.nodeHeader.keySize) + 4

	ip.nodeHeader.isRoot = false
	ip.nodeHeader.parent = newParentInd
	dest.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + ip.nodeHeader.keySize
	copy(newParent.nodeBody[pointerOffset+2*4+ip.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset:], oldChildInd)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+ip.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++

	/**
	 * Transfer data to new pages
	 */
	copy(dest.nodeBody[:],
		ip.nodeBody[middleElementOffset+4+ip.nodeHeader.keySize:])
}

func (ip *InternalPage) transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, dest *Page) {

	// find the offset and the key of the middle element
	middleElementOffset := ip.getOffset((ip.nodeHeader.numCells - 1) / 2)
	middleElementKey := ip.getKey((ip.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	dest.nodeHeader.numCells = ip.nodeHeader.numCells / 2
	ip.nodeHeader.numCells = (ip.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.nodeHeader.totalBodySize =
		dest.nodeHeader.numCells*(4+dest.nodeHeader.keySize) + 4
	ip.nodeHeader.totalBodySize = ip.nodeHeader.numCells*(4+dest.nodeHeader.keySize) + 4

	ip.nodeHeader.isRoot = false
	ip.nodeHeader.parent = newParentInd
	dest.nodeHeader.parent = newParentInd

	indForKey := newParent.findIndexForKeyInternal(middleElementKey)
	pointerOffset := newParent.getOffsetInternal(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.nodeHeader.totalBodySize += 4 + ip.nodeHeader.keySize + 4
	copy(newParent.nodeBody[pointerOffset+2*4+ip.nodeHeader.keySize:],
		newParent.nodeBody[pointerOffset+4:newParent.nodeHeader.totalBodySize])
	copy(newParent.nodeBody[pointerOffset+4:], middleElementKey)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset:], oldChildInd)
	binary.LittleEndian.PutUint32(newParent.nodeBody[pointerOffset+4+ip.nodeHeader.keySize:], newChildInd)
	newParent.nodeHeader.numCells++

	/**
	 * Transfer data to new pages
	 */
	copy(dest.nodeBody[:],
		ip.nodeBody[middleElementOffset+4+ip.nodeHeader.keySize:])
}
