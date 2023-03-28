package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

type tree struct {
	nodes   [][]*node
	current *node
	longest int

	run  func(n *node)
	undo func(n *node)
}

type node struct {
	Slot         int
	Transactions []string
	Length       int

	parent   *node
	children []*node
}

func makeTree(run func(n *node), undo func(n *node)) tree {
	h := make([]*node, 1)
	genisis := &node{
		Slot:   0,
		Length: 0,
	}
	h[0] = genisis
	nodes := [][]*node{h}
	return tree{
		nodes:   nodes,
		current: genisis,

		run:  run,
		undo: undo,
	}
}

func (t *tree) insert(slot int, transactions []string, parentSlot int, parentHash []byte) {
	parent := t.findParent(parentSlot, parentHash)
	n := &node{
		Slot:         slot,
		Transactions: transactions,
		Length:       parent.Length + 1,
		parent:       parent,
		children:     []*node{},
	}
	parent.children = append(parent.children, n)
	t.append(n)

	if n.Length > t.longest {
		t.longest = n.Length
		t.goTo(n)
	} else if n.Length == t.longest {
		t.goTo(tieBreaker(t.current, n))
	}
}

func (t *tree) insertNext(slot int, transactions []string) *node {
	n := &node{
		Slot:         slot,
		Transactions: transactions,
		Length:       t.current.Length + 1,
		parent:       t.current,
		children:     []*node{},
	}

	setParentChild(t.current, n)
	t.append(n)

	t.run(n)
	t.current = n

	return n
}

func tieBreaker(nodes ...*node) *node {
	best := nodes[0]
	bestHash := best.hash()
	for _, x := range nodes[1:] {
		if bytes.Compare(x.hash(), bestHash) == 1 {
			best = x
		}
	}
	return best
}

func (t *tree) append(n *node) {
	for len(t.nodes) <= n.Slot {
		t.nodes = append(t.nodes, []*node{})
	}
	t.nodes[n.Slot] = append(t.nodes[n.Slot], n)
}

func (t *tree) goTo(n *node) {
	var last *node
	for !t.current.containsExcluded(n, last) {
		t.undo(t.current)
		last = t.current
		t.current = t.current.parent
	}

	x := n
	path := make([]*node, 0)
	for x != t.current {
		path = append(path, x)
		last = x
		x = x.parent
	}

	for i := len(path) - 1; i >= 0; i-- {
		t.run(path[i])
		for _, child := range t.current.children {
			if child == path[i] {
				t.current = child
			}
		}
	}

	t.current = n
}

func (n *node) containsExcluded(n2 *node, excluded *node) bool {
	if n == n2 {
		return true
	}
	for _, x := range n.children {
		if x == excluded {
			continue
		}

		if x.contains(n2) {
			return true
		}
	}
	return false
}

func (n *node) contains(n2 *node) bool {
	if n == n2 {
		return true
	}
	for _, x := range n.children {
		if x.contains(n2) {
			return true
		}
	}
	return false
}

func setParentChild(parent *node, child *node) {
	parent.children = append(parent.children, child)
	child.parent = parent
}

func (t *tree) findParent(parentSlot int, H []byte) *node {
	for _, x := range t.nodes[parentSlot] {
		if bytes.Equal(x.hash(), H) {
			return x
		}
	}
	log.Fatal("invalid parent")
	return nil
}

func (n *node) hash() []byte {
	return hashObject(n)
}

func (t *tree) print(alias string) {
	logger := log.New(os.Stdout, "", 0)
	str := "-------------------------\n" + alias + "\n"
	for _, slot := range t.nodes {
		for _, node := range slot {
			if node.parent != nil {
				str += fmt.Sprint("(", node.hash()[:2], " ", node.Slot, " ", node.parent.hash()[:2], len(node.Transactions), ")")
				if t.current == node {
					str += "* "
				} else {
					str += "  "
				}
			} else {
				str += fmt.Sprint("(", node.hash()[:2], " ", node.Slot, " ", "nil ", len(node.Transactions), ")")
			}
		}
		if len(slot) > 0 {
			str += "\n"
		}
	}
	str += "-------------------------"
	logger.Print(str)
}
