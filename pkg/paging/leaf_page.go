package paging

import (
	"bytes"
	"encoding/binary"
)

type LeafPage struct {
	PageBase
}

func NewLeafPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint16, totalBodySize uint16) *LeafPage {
	return nil
}

func NewLeafPage() *LeafPage {
	return nil
}

func (lp *LeafPage) getType() NodeType {
	return lp.nodeHeader.nodeType
}

func (lp *LeafPage) getIsRoot() bool {
	return lp.nodeHeader.isRoot
}

func (lp *LeafPage) getParent() uint32 {
	return lp.nodeHeader.parent
}

func (lp *LeafPage) getNumCells() uint16 {
	return lp.nodeHeader.numCells
}

func (lp *LeafPage) getKeySize() uint16 {
	return lp.nodeHeader.keySize
}

func (lp *LeafPage) getTotalBodySize() uint16 {
	return lp.nodeHeader.totalBodySize
}

func (lp *LeafPage) setIsRoot(isRoot bool) {
	lp.nodeHeader.isRoot = isRoot
}

func (lp *LeafPage) setParent(parent uint32) {
	lp.nodeHeader.parent = parent
}

func (lp *LeafPage) setNumCells(numCells uint16) {
	lp.nodeHeader.numCells = numCells
}

func (lp *LeafPage) setTotalBodySize(totalBodySize uint16) {
	lp.nodeHeader.totalBodySize = totalBodySize
}

func (lp *LeafPage) setNodeBody(nodeBodyBytes []byte) {
	copy(lp.nodeBody[:], nodeBodyBytes)
}

func (lp *LeafPage) setNodeBodyRange(nodeBodyBytes []byte, startInd uint16) {
	copy(lp.nodeBody[startInd:], nodeBodyBytes)
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

func (lp *LeafPage) getBody() []byte {
	return lp.nodeBody[:]
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

func (lp *LeafPage) transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent IPage, dest IPage) {
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
	dest.setNumCells((lp.nodeHeader.numCells + 1) / 2)
	lp.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	dest.setTotalBodySize(dest.getNumCells()*OFFSET_SIZE + (lp.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset))

	lp.nodeHeader.totalBodySize = lp.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	lp.nodeHeader.isRoot = false
	lp.nodeHeader.parent = newParentInd
	dest.setParent(newParentInd)

	indForKey := newParent.findIndexForKey(middleElementKey)
	pointerOffset := newParent.getOffset(indForKey)
	// pointerOffset := p.getOffsetInternal(indForKey)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	newParent.setTotalBodySize(newParent.getTotalBodySize() + 4 + lp.nodeHeader.keySize)
	newParent.setNodeBodyRange(newParent.getBody()[pointerOffset+4:newParent.getTotalBodySize()],
		pointerOffset+2*4+lp.nodeHeader.keySize)

	newParent.setNodeBodyRange(middleElementKey, pointerOffset+4)

	newChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newChildIndBytes, newChildInd)
	newParent.setNodeBodyRange(newChildIndBytes, pointerOffset+4+lp.nodeHeader.keySize)

	newParent.setNumCells(newParent.getNumCells() + 1)
	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < dest.getNumCells(); i++ {
		offset := lp.getOffset(lp.nodeHeader.numCells + i)
		offset -= middleElementOffset
		offsetBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(offsetBytes, offset)
		dest.setNodeBodyRange(offsetBytes, i*OFFSET_SIZE)
	}

	/**
	 * Transfer data to new pages
	 */
	dest.setNodeBodyRange(lp.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize],
		dest.getNumCells()*OFFSET_SIZE)

	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(lp.nodeBody[lp.nodeHeader.numCells*OFFSET_SIZE:], lp.nodeBody[oldStartOfCells:lp.nodeHeader.totalBodySize])
}

func (lp *LeafPage) transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent IPage, dest IPage) {

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
	newParent.setNumCells(newParent.getNumCells() + 1)
	dest.setNumCells((lp.nodeHeader.numCells + 1) / 2)
	lp.nodeHeader.numCells /= 2

	/**
	 * update total body size according to new elements added to each page
	 */
	// Size of parent increases by one key size and two pointer sizes (for now)
	// TODO: change 2*4 to 2*CHILD_POINTER_SIZE
	newParent.setTotalBodySize(newParent.getTotalBodySize() + 2*4 + lp.nodeHeader.keySize)
	dest.setTotalBodySize(dest.getNumCells()*OFFSET_SIZE + (lp.nodeHeader.totalBodySize - oldStartOfCells - middleElementOffset))
	lp.nodeHeader.totalBodySize = lp.nodeHeader.numCells*OFFSET_SIZE + middleElementOffset

	lp.nodeHeader.isRoot = false
	lp.nodeHeader.parent = newParentInd
	dest.setParent(newParentInd)

	/**
	 * Update the parent and the new child's offset list
	 */
	// put children pointers and copy middle element key to the new parent
	oldChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(oldChildIndBytes, oldChildInd)
	newParent.setNodeBody(oldChildIndBytes)

	newParent.setNodeBodyRange(middleElementKey, 4)

	newChildIndBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(newChildIndBytes, newChildInd)
	newParent.setNodeBodyRange(newChildIndBytes, 4+lp.nodeHeader.keySize)

	// move offsets from the existing child to the new one, updating them in the process
	for i := uint16(0); i < dest.getNumCells(); i++ {
		offset := lp.getOffset(lp.nodeHeader.numCells + i)
		offset -= middleElementOffset
		offsetBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(offsetBytes, offset)
		dest.setNodeBodyRange(offsetBytes, i*OFFSET_SIZE)
	}

	/**
	 * Transfer data to new pages
	 */
	dest.setNodeBodyRange(lp.nodeBody[oldStartOfCells+middleElementOffset:oldTotalBodySize],
		dest.getNumCells()*OFFSET_SIZE)

	// for the existing node, just shift data left, since the number of offsets is decreased
	copy(lp.nodeBody[lp.nodeHeader.numCells*OFFSET_SIZE:], lp.nodeBody[oldStartOfCells:lp.nodeHeader.totalBodySize])
}

func (lp *LeafPage) hasSufficientSpace(addedSize uint16) bool {
	// TODO: check if there is enough space for new data
	oldSize := NODE_HEADER_SIZE + lp.nodeHeader.totalBodySize
	newSize := oldSize + addedSize + lp.nodeHeader.keySize + DATA_SIZE_SIZE
	return newSize <= PAGE_SIZE
}
