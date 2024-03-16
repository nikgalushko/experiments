package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

func main() {
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

	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func(n *Node) {
			defer wg.Done()
			n.Do(fmt.Sprintf("value-%d", n.ID))
		}(&nodes[i])
	}

	wg.Wait()

	for i := 0; i < size; i++ {
		log.Println("Node", i, "K", nodes[i].K, "State", nodes[i].State.HighestProposeNumber, nodes[i].State.Vote)
	}
}

type (
	ProposeNumber struct {
		k, id int
	}

	State struct {
		mu                   sync.Mutex
		HighestProposeNumber ProposeNumber
		Vote                 Propose
	}

	Propose struct {
		n     ProposeNumber
		value string
	}

	Node struct {
		ID        int
		K         int
		State     *State
		Acceptors []*Node
	}

	UserRequest struct {
		operation string
		value     int
	}
)

func (p ProposeNumber) Compare(p2 ProposeNumber) int {
	if p.k > p2.k {
		return 1
	}
	if p.k < p2.k {
		return -1
	}
	if p.id > p2.id {
		return 1
	}
	if p.id < p2.id {
		return -1
	}
	return 0
}

func (n *Node) Do(val string) {
	phase1 := func() (pn ProposeNumber) {
		isDone := false
		for !isDone {
			n.K++
			pn = ProposeNumber{k: n.K, id: n.ID}

			errsCount := 0
			var proposes []Propose
			for _, acceptor := range n.Acceptors {
				p, err := acceptor.HandlePrepare(pn)
				if err != nil {
					errsCount++
					continue
				}
				proposes = append(proposes, p)
			}

			if errsCount > len(n.Acceptors)/2 {
				log.Println("phase1: errsCount > len(n.Acceptors)/2; repeat")
				continue
			}
			isDone = true

			var maxPropose Propose
			for _, p := range proposes {
				if p.n.Compare(maxPropose.n) == 1 {
					maxPropose = p
				}
			}

			if maxPropose.n.k != 0 {
				val = maxPropose.value
				log.Println("phase1: maxPropose.n.k != 0; change phase2 value to", val)
			}
			log.Println("phase1: maxPropose.n.k == 0; continue to phase2")
		}
		return
	}

	phase2 := func(pn ProposeNumber) error {
		errsCount := 0
		for _, acceptor := range n.Acceptors {
			_, err := acceptor.HandlePropose(Propose{n: pn, value: val})
			if err != nil {
				errsCount++
			}
		}

		if errsCount > len(n.Acceptors)/2 {
			log.Println("phase2: errsCount > len(n.Acceptors)/2")
			return errors.New("phase2 failed")
		}

		log.Println("phase2: success")
		return nil
	}

	for {
		pn := phase1()
		err := phase2(pn)
		if err == nil {
			break
		}
	}
}

func (n *Node) HandlePrepare(p ProposeNumber) (Propose, error) {
	if p.Compare(n.State.HighestProposeNumber) == 1 {
		n.State.mu.Lock()
		n.State.HighestProposeNumber = p
		vote := n.State.Vote
		n.State.mu.Unlock()

		return vote, nil
	}

	return Propose{}, errors.New("reject")
}

func (n *Node) HandlePropose(p Propose) (ProposeNumber, error) {
	if p.n.Compare(n.State.HighestProposeNumber) != -1 {
		n.State.mu.Lock()
		n.State.HighestProposeNumber = p.n
		n.State.Vote = p
		n.State.mu.Unlock()

		return p.n, nil
	}

	return ProposeNumber{}, errors.New("reject")
}
