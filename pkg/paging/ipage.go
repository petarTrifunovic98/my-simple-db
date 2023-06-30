package paging

type IPage interface {
	getOffset(ind uint16) uint16
	getKey(ind uint16) []byte
	findIndexForKey(key []byte) uint16
	transferCellsNotRoot(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page)
	transferCells(newParentInd uint32, oldChildInd uint32, newChildInd uint32, newParent *Page, destination *Page)
}

type PageBase struct {
	nodeHeader NodeHeader
	nodeBody   [PAGE_SIZE - NODE_HEADER_SIZE]byte
}
