package main

import "testing"

func FuzzSingleDecreePaxos(f *testing.F) {
	const size = 5
	nodes := []Node{}
	for i := 0; i < size; i++ {
		nodes = append(nodes, Node{ID: i, K: 0, State: &State{}})
	}

	for i := 0; i < size; i++ {
		acceptors := []*Node{}
		for j := 0; j < size; j++ {
			if i != j {
				acceptors = append(acceptors, &nodes[j])
			}
		}
		nodes[i].Acceptors = acceptors
	}

	f.Fuzz(func(t *testing.T, in []byte) {

	})
}
