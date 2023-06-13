package paging

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type NodeType uint8

const (
	LEAF_NODE NodeType = iota
	INTERNAL_NODE
)

const NODE_HEADER_SIZE = 1 + 1 + 4 + 2 + 2 + 2 //add the sizes of the types used in NodeHeader struct

type NodeHeader struct {
	parent        uint32
	numCells      uint16
	totalBodySize uint16
	keySize       uint16
	nodeType      NodeType
	isRoot        bool
}

func (nh *NodeHeader) Serialize() []byte {
	parentBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(parentBytes, nh.parent)

	numCellsBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(numCellsBytes, nh.numCells)

	totalBodySizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalBodySizeBytes, nh.totalBodySize)

	keySizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(keySizeBytes, nh.keySize)

	fmt.Println(unsafe.Sizeof(*nh))
	nodeTypeBytes := byte(nh.nodeType)

	isRootUint8 := uint8(0)
	if nh.isRoot {
		isRootUint8 = 1
	}
	isRootBytes := byte(isRootUint8)

	nodeHeaderBytes := make([]byte, 0, NODE_HEADER_SIZE)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, parentBytes...)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, numCellsBytes...)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, totalBodySizeBytes...)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, keySizeBytes...)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))
	nodeHeaderBytes = append(nodeHeaderBytes, nodeTypeBytes, isRootBytes)
	fmt.Println("Header byte len:", len(nodeHeaderBytes))

	return nodeHeaderBytes
}

func (nh *NodeHeader) Deserialize(nodeHeaderBytes []byte) {
	nh.parent = binary.LittleEndian.Uint32(nodeHeaderBytes[0:4])
	nh.numCells = binary.LittleEndian.Uint16(nodeHeaderBytes[4:6])
	nh.totalBodySize = binary.LittleEndian.Uint16(nodeHeaderBytes[6:8])
	nh.keySize = binary.LittleEndian.Uint16(nodeHeaderBytes[8:10])
	nh.nodeType = NodeType(nodeHeaderBytes[10])
	isRootBytes := nodeHeaderBytes[11]
	nh.isRoot = true
	if isRootBytes == 0 {
		nh.isRoot = false
	}
}

func (nh *NodeHeader) Print() {
	if nh.nodeType == LEAF_NODE {
		fmt.Println("leaf node with", nh.numCells, "cells")
	} else {
		fmt.Println("internal node with", nh.numCells, "cells")
	}
}
