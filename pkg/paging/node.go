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

const NODE_HEADER_SIZE = 1 + 1 + 4 + 4 //add the sizes of the types used in NodeHeader struct

type NodeHeader struct {
	nodeType NodeType
	isRoot   bool
	parent   uint32
	numCells uint32
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

	numCellsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numCellsBytes, nh.numCells)

	nodeHeaderBytes := make([]byte, 0, len(parentBytes)+len(numCellsBytes)+2)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, nodeTypeBytes, isRootBytes)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, parentBytes...)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, numCellsBytes...)

	fmt.Println("Header byte len:", len(nodeHeaderBytes))

	return nodeHeaderBytes
}

func (nh *NodeHeader) Deserialize(nodeHeaderBytes []byte) {
	nh.nodeType = NodeType(nodeHeaderBytes[0])
	isRootBytes := nodeHeaderBytes[1]
	nh.isRoot = true
	if isRootBytes == 0 {
		nh.isRoot = false
	}

	nh.parent = binary.LittleEndian.Uint32(nodeHeaderBytes[2:6])
	nh.numCells = binary.LittleEndian.Uint32(nodeHeaderBytes[6:10])
}

func (nh *NodeHeader) Print() {
	if nh.nodeType == LEAF_NODE {
		fmt.Println("leaf node with", nh.numCells, "cells")
	} else {
		fmt.Println("internal node with", nh.numCells, "cells")
	}
}
