package paging

import (
	"bytes"
	"encoding/binary"
)

type InternalPage struct {
	PageBase
}

func NewInternalPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint16, totalBodySize uint16) *InternalPage {
	p := &InternalPage{
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

	return p
}

func NewInternalPage() *InternalPage {
	return nil
}

func (ip *InternalPage) getHeader() *NodeHeader {
	return &ip.nodeHeader
}

func (ip *InternalPage) getType() NodeType {
	return ip.nodeHeader.nodeType
}

func (ip *InternalPage) getIsRoot() bool {
	return ip.nodeHeader.isRoot
}

func (ip *InternalPage) getParent() uint32 {
	return ip.nodeHeader.parent
}

func (ip *InternalPage) getNumCells() uint16 {
	return ip.nodeHeader.numCells
}

func (ip *InternalPage) getKeySize() uint16 {
	return ip.nodeHeader.keySize
}

func (ip *InternalPage) getTotalBodySize() uint16 {
	return ip.nodeHeader.totalBodySize
}

func (ip *InternalPage) setIsRoot(isRoot bool) {
	ip.nodeHeader.isRoot = isRoot
}

func (ip *InternalPage) setParent(parent uint32) {
	ip.nodeHeader.parent = parent
}

func (ip *InternalPage) setNumCells(numCells uint16) {
	ip.nodeHeader.numCells = numCells
}

func (ip *InternalPage) setTotalBodySize(totalBodySize uint16) {
	ip.nodeHeader.totalBodySize = totalBodySize
}

func (ip *InternalPage) setNodeBody(nodeBodyBytes []byte) {
	copy(ip.nodeBody[:], nodeBodyBytes)
}

func (ip *InternalPage) setNodeBodyRange(nodeBodyBytes []byte, startInd uint16) {
	copy(ip.nodeBody[startInd:], nodeBodyBytes)
}

func (ip *InternalPage) getOffset(ind uint16) uint16 {
	return ind * (4 + ip.nodeHeader.keySize)
}

func (ip *InternalPage) getKey(ind uint16) []byte {
	return ip.nodeBody[4+ind*(4+ip.nodeHeader.keySize) : 4+ind*(4+ip.nodeHeader.keySize)+ip.nodeHeader.keySize]
}

func (ip *InternalPage) getBody() []byte {
	return ip.nodeBody[:]
}

func (ip *InternalPage) getPointer(ind uint16) uint32 {
	pointerBytes := ip.nodeBody[ind*(4+ip.nodeHeader.keySize) : ind*(4+ip.nodeHeader.keySize)+4]
	return binary.LittleEndian.Uint32(pointerBytes)
}

func (ip *InternalPage) findIndexForKey(key []byte) (ind uint16, exists bool) {
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
			return currentIndex, true
		}

		currentIndex = (leftIndex + rightIndex) / 2
	}

	return currentIndex, false
}

func (ip *InternalPage) transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent IPage, dest IPage) {

	// find the offset and the key of the middle element
	middleElementOffset := ip.getOffset((ip.nodeHeader.numCells - 1) / 2)
	middleElementKey := ip.getKey((ip.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	dest.setNumCells(ip.nodeHeader.numCells / 2)
	ip.nodeHeader.numCells = (ip.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.setTotalBodySize(dest.getNumCells()*(4+dest.getKeySize()) + 4)
	ip.nodeHeader.totalBodySize = ip.nodeHeader.numCells*(4+dest.getKeySize()) + 4

	ip.nodeHeader.isRoot = false
	ip.nodeHeader.parent = newParentInd
	dest.setParent(newParentInd)

	indForKey, _ := newParent.findIndexForKey(middleElementKey)
	pointerOffset := newParent.getOffset(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.setTotalBodySize(newParent.getTotalBodySize() + 4 + ip.nodeHeader.keySize)
	newParent.setNodeBodyRange(newParent.getBody()[pointerOffset+4:newParent.getTotalBodySize()],
		pointerOffset+2*4+ip.nodeHeader.keySize)

	newParent.setNodeBodyRange(middleElementKey, pointerOffset+4)

	oldChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(oldChildIndBytes, oldChildInd)
	newParent.setNodeBodyRange(oldChildIndBytes, pointerOffset)

	newChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newChildIndBytes, newChildInd)
	newParent.setNodeBodyRange(newChildIndBytes, pointerOffset+4+ip.nodeHeader.keySize)

	newParent.setNumCells(newParent.getNumCells() + 1)

	/**
	 * Transfer data to new pages
	 */
	dest.setNodeBody(ip.nodeBody[middleElementOffset+4+ip.nodeHeader.keySize:])
}

func (ip *InternalPage) transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent IPage, dest IPage) {

	// find the offset and the key of the middle element
	middleElementOffset := ip.getOffset((ip.nodeHeader.numCells - 1) / 2)
	middleElementKey := ip.getKey((ip.nodeHeader.numCells - 1) / 2)

	/**
	 * update number of cells; newParent gets one new element,
	 * while children are left with half of the previous number
	 * of elements
	 */
	dest.setNumCells(ip.nodeHeader.numCells / 2)
	ip.nodeHeader.numCells = (ip.nodeHeader.numCells - 1) / 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.setTotalBodySize(dest.getNumCells()*(4+dest.getKeySize()) + 4)
	ip.nodeHeader.totalBodySize = ip.nodeHeader.numCells*(4+dest.getKeySize()) + 4

	ip.nodeHeader.isRoot = false
	ip.nodeHeader.parent = newParentInd
	dest.setParent(newParentInd)

	indForKey, _ := newParent.findIndexForKey(middleElementKey)
	pointerOffset := newParent.getOffset(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.setTotalBodySize(newParent.getTotalBodySize() + 4 + ip.nodeHeader.keySize + 4)
	newParent.setNodeBodyRange(newParent.getBody()[pointerOffset+4:newParent.getTotalBodySize()],
		pointerOffset+2*4+ip.nodeHeader.keySize)
	newParent.setNodeBodyRange(middleElementKey, pointerOffset+4)

	oldChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(oldChildIndBytes, oldChildInd)
	newParent.setNodeBodyRange(oldChildIndBytes, pointerOffset)

	newChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newChildIndBytes, newChildInd)
	newParent.setNodeBodyRange(newChildIndBytes, pointerOffset+4+ip.nodeHeader.keySize)

	newParent.setNumCells(newParent.getNumCells() + 1)

	/**
	 * Transfer data to new pages
	 */

	dest.setNodeBody(ip.nodeBody[middleElementOffset+4+ip.nodeHeader.keySize:])
}

func (ip *InternalPage) hasSufficientSpace(addedSize uint16) bool {
	// oldSize := NODE_HEADER_SIZE + p.nodeHeader.totalBodySize
	// newSize := oldSize + addedSize
	// return newSize <= PAGE_SIZE

	// For testing purposes. The commented code above is the correct code.
	return ip.nodeHeader.numCells < 3

}
