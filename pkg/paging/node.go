package paging

type NodeType uint8

const (
	LEAF_NODE NodeType = iota
	INTERNAL_NODE
)

type Node struct {
	nodeType NodeType
	isRoot   bool
	parent   *Node
}
