package paging

type IPage interface {
	getType() NodeType
	getIsRoot() bool
	getParent() uint32
	getNumCells() uint16
	getKeySize() uint16
	getTotalBodySize() uint16
	setIsRoot(bool)
	setParent(uint32)
	setNumCells(uint16)
	setTotalBodySize(uint16)
	setNodeBody([]byte)
	setNodeBodyRange([]byte, uint16)
	getOffset(uint16) uint16
	getKey(uint16) []byte
	getBody() []byte
	findIndexForKey([]byte) uint16
	transferCellsNotRoot(uint32, uint32, uint32, IPage, IPage)
	transferCells(uint32, uint32, uint32, IPage, IPage)
	hasSufficientSpace(uint16) bool
}

type PageBase struct {
	nodeHeader NodeHeader
	nodeBody   [PAGE_SIZE - NODE_HEADER_SIZE]byte
}

func NewIPageWithParams(nodeType NodeType, isRoot bool, parent uint32, numCells uint16, totalBodySize uint16) IPage {
	if nodeType == LEAF_NODE {
		return NewLeafPageWithParams(nodeType, isRoot, parent, numCells, totalBodySize)
	} else {
		return NewInternalPageWithParams(nodeType, isRoot, parent, numCells, totalBodySize)
	}
}
