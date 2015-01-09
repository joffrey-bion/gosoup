package gosoup

import (
	"strings"
)

const (
	blank          string = " \t\n\r"
	nodeBufferSize int    = 20
)

// NodeIterator is useful to iterate over a set of nodes without storing all
// references in a slice.
//
// It also allows chaining methods to filter the nodes in some way.
//
// The caller should close the iterator via the Close() method when no more nodes
// are going to be read, unless he exhausts the iterator's Nodes channel. This
// unblocks internal goroutines and allows their garbage collection.
type NodeIterator struct {
	Nodes  <-chan *Node
	exit   chan interface{}
	closed bool
}

// Close notifies this NodeIterator that no more nodes will be read from it.
// This prevents the internal goroutines from hanging forever.
//
// This function should be called when the caller stops reading nodes while
// the channel is not exhausted. When the channel is exhausted, there is no
// need to call Close().
func (i NodeIterator) Close() {
	if !i.closed {
		i.exit <- true
		i.closed = true
	}
}

// First retrieves the first node of this iterator and closes it.
func (i NodeIterator) First() *Node {
	node, ok := <-i.Nodes
	if !ok {
		// no nodes at all
		return nil
	}
	// at least one node, notify that we break early
	i.Close()
	return node
}

// All retrieves all nodes from this iterator and returns them as a slice.
func (i NodeIterator) All() []*Node {
	var list []*Node
	for node := range i.Nodes {
		list = append(list, node)
	}
	return list
}

// Apply applies the given function to all Nodes of this iterator.
func (i NodeIterator) Apply(f func(n *Node)) {
	for node := range i.Nodes {
		f(node)
	}
}

// Filter returns a new iterator that only iterates on the Node that
// match the given predicate.
func (i NodeIterator) Filter(predicate func(*Node) bool) NodeIterator {
	c := make(chan *Node)
	filtered := NodeIterator{c, i.exit, false}
	go func() {
		for node := range i.Nodes {
			if predicate(node) {
				c <- node
			}
		}
		close(c)
	}()
	return filtered
}

// Limit returns a new iterator that automatically stops if it has
// read the given maximum number of Nodes.
func (i NodeIterator) Limit(max int) NodeIterator {
	c := make(chan *Node)
	limited := NodeIterator{c, i.exit, false}
	go func() {
		count := 0
		for node := range i.Nodes {
			c <- node
			count++
			if count > max {
				i.Close()
				break
			}
		}
		close(c)
	}()
	return limited
}

// clean trims leading and trailing whitespace if the node is a TextNode.
//
// Calling this method actually mutates the given node.
// It returns the same -although mutated- node for convenience.
func clean(node *Node) *Node {
	if node.Type == TextNode {
		node.Data = strings.Trim(node.Data, blank)
	}
	return node
}

// notBlank returns true if the node's data is not full of blank space
func notBlank(node *Node) bool {
	return strings.Trim(node.Data, blank) != ""
}

// TreeIterator returns an iterator on this node's descendants in depth-first order.
// If recursive is false, only direct children are considered.
func (node *Node) TreeIterator(recursive bool) NodeIterator {
	if node == nil {
		panic("iterateOnDescendants: null input node")
	}
	out := make(chan *Node, nodeBufferSize)
	exit := make(chan interface{}, 1)
	go func() {
		node.recIterateOnDescendants(recursive, out, exit)
		close(out)
	}()
	return NodeIterator{out, exit, false}
}

func (node *Node) recIterateOnDescendants(recursive bool, out chan<- *Node, exit chan interface{}) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		select {
		case <-exit:
			// the caller will not read any more nodes, so
			// don't try to send to avoid blocking forever
			exit <- true // to exit all calls in the recursive stack
			return
		default:
			out <- clean(child)
			if recursive {
				// browse the child's children
				child.recIterateOnDescendants(recursive, out, exit)
			}
		}
	}
}

// Children returns an iterator on this node's direct children.
func (node *Node) Children() NodeIterator {
	return node.TreeIterator(false).Filter(notBlank)
}

// Descendants returns an iterator on this node's descendants, in depth-first
// order.
func (node *Node) Descendants() NodeIterator {
	return node.TreeIterator(true).Filter(notBlank)
}

// ChildrenMatching returns an iterator on this node's direct children that match
// the given predicate.
func (node *Node) ChildrenMatching(predicate func(node *Node) bool) NodeIterator {
	return node.Children().Filter(predicate)
}

// DescendantsMatching returns an iterator on this node's descendants that match
// the given predicate, in depth-first order.
func (node *Node) DescendantsMatching(predicate func(node *Node) bool) NodeIterator {
	return node.Descendants().Filter(predicate)
}

func predicateIsTag(tagName string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.IsTag(tagName)
	}
}

// ChildrenByTag returns an iterator on this node's direct children with the specified
// tag name.
func (node *Node) ChildrenByTag(tagName string) NodeIterator {
	return node.ChildrenMatching(predicateIsTag(tagName))
}

// DescendantsByTag returns an iterator on this node's descendants with the specified
// tag name, in depth-first order.
func (node *Node) DescendantsByTag(tagName string) NodeIterator {
	return node.DescendantsMatching(predicateIsTag(tagName))
}

func predicateAttrValueContains(attrKey, match string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.AttrValueContains(attrKey, match)
	}
}

// ChildrenByAttrValueContaining returns an iterator on this node's direct children
// that have attributes whose value contains the match string.
func (node *Node) ChildrenByAttrValueContaining(attrKey, match string) NodeIterator {
	return node.ChildrenMatching(predicateAttrValueContains(attrKey, match))
}

// DescendantsByAttrValueContaining returns an iterator on this node's descendants
// that have attributes whose value contains the match string, in depth-first order.
func (node *Node) DescendantsByAttrValueContaining(attrKey, match string) NodeIterator {
	return node.DescendantsMatching(predicateAttrValueContains(attrKey, match))
}
