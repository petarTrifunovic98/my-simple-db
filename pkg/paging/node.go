package paging

import (
	"encoding/binary"
	"fmt"
)

type NodeType uint8

const (
	LEAF_NODE NodeType = iota
	INTERNAL_NODE
)

const NODE_HEADER_SIZE = 1 + 1 + 4 //add the sizes of the types used in NodeHeader struct

type NodeHeader struct {
	nodeType NodeType
	isRoot   bool
	parent   uint32
}

func (nh *NodeHeader) Serialize() []byte {
	nodeTypeBytes := byte(nh.nodeType)

	isRootUint8 := uint8(0)
	if nh.isRoot {
		isRootUint8 = 1
	}
	isRootBytes := byte(isRootUint8)
	parentBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(parentBytes, nh.parent)

	nodeHeaderBytes := make([]byte, 0, len(parentBytes)+2)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, nodeTypeBytes, isRootBytes)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, parentBytes...)

	fmt.Println("Header byte len:", len(nodeHeaderBytes))

	return nodeHeaderBytes
}
