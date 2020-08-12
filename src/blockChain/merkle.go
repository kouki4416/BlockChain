package blockChain

import "crypto/sha256"

type MerkleTree struct{
	RootNode *MerkleNode
}

type MerkleNode struct{
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}

func NewMerkleNode(left, right *MerkleNode, data[]byte) *MerkleNode{
	node := MerkleNode{}

	if left == nil && right == nil{
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data ... )
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}


//Just concatnate all nodes to create merkle tree
func NewMerkleTree(data [][]byte) *MerkleTree{
	var nodes []MerkleNode

	if len(data) % 2 != 0 { // check if nodes are even o/w copy last one
		data = append(data, data[len(data) - 1])
	}

	for _, dat := range data{
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	//connect nodes into a tree shape
	for  i := 0; i < len(data)/2; i++{
		var level []MerkleNode

		for j := 0; j < len(nodes); j+=2{
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node)
		}

		nodes = level
	}

	tree := MerkleTree{&nodes[0]}
	return &tree
}