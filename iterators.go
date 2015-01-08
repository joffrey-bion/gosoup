package gosoup

import (
	"strings"
	"sync"
)

const (
	blank string = " \t\n\r"
	nodeBufferSize int = 20
)

// First retrieves the first node from the given output channel, and takes
// care of the cleaning through the exit channel.
//
// This function can be particularly useful when combined with the iterating
// functions of GoSoup:
//
//     firstChild := gosoup.First(node.Children())
//
//     firstMyClassDescendant := gosoup.First(node.DescendantsByAttributeValueContaining("class", "myClass"))
//
// No need to take care of channels here.
func First(output <-chan *Node, exit chan interface{}) *Node {
	node := <-output
	exit <- true
	return node
}

// Collect gathers all nodes from the given output channel in a slice.
//
// This function can be particularly useful when combined with the iterating
// functions of GoSoup.
func Collect(output <-chan *Node, exit chan interface{}) []*Node {
	var list []*Node
	for node := range output {
		list = append(list, node)
	}
	return list
}

func forward(in <-chan interface{}, out chan interface{}) {
	for e := range in {
		out <- e
	}
}

func forwardNodes(in <-chan *Node, out chan *Node) {
	for e := range in {
		out <- e
	}
}

// notBlank returns true if the node's data is not full of blank space
func notBlank(node *Node) bool {
	return strings.Trim(node.Data, blank) != ""
}

// clean trims leading and trailing whitespace if the node is a TextNode.
func clean(node *Node) *Node {
	if node.Type == TextNode {
		node.Data = strings.Trim(node.Data, blank)
	}
	return node
}

// DescendantsIterator iterates on this node's descendants in depth-first order,
// and send into the output channel the ones that match the given predicate.
//
// If recursive is false, only direct children are considered.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func (node *Node) DescendantsIterator(predicate func(node *Node) bool, recursive bool) (output <-chan *Node, exit chan interface{}) {
	if node == nil {
		panic("iterateOnDescendants: null input node")
	}
	out := make(chan *Node, nodeBufferSize)
	exit = make(chan interface{}, 1)
	go func() {
		var wg sync.WaitGroup
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			select {
			case <-exit:
				// the caller will not read any more nodes, so
				// don't try to send to avoid blocking forever
				wg.Wait()
				close(out)
				return
			default:
				if predicate(child) {
					out <- clean(child)
				}
				if recursive {
					// browse the child's children
					wg.Add(1)
					go func(child *Node) {
						defer wg.Done()
						in, subexit := child.DescendantsIterator(predicate, recursive)
						go forward(exit, subexit)
						forwardNodes(in, out)
					}(child)
				}
			}
		}
		wg.Wait()
		close(out)
	}()
	return out, exit
}

// ChildrenMatching finds this node's direct children that match the
// given predicate, and sends them into the output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func (node *Node) ChildrenMatching(predicate func(node *Node) bool) (output <-chan *Node, exit chan interface{}) {
	return node.DescendantsIterator(predicate, false)
}

// DescendantsMatching finds this node's descendants that match the
// given predicate, and sends them into the output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func (node *Node) DescendantsMatching(predicate func(node *Node) bool) (output <-chan *Node, exit chan interface{}) {
	return node.DescendantsIterator(predicate, true)
}

// Children finds this node's direct children, and sends them into the
// output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func (node *Node) Children() (output <-chan *Node, exit chan interface{}) {
	return node.ChildrenMatching(notBlank)
}

// Descendants finds this node's descendants, and sends them into the
// output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func (node *Node) Descendants() (output <-chan *Node, exit chan interface{}) {
	return node.DescendantsMatching(notBlank)
}

func predicateIsTag(tagName string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.IsTag(tagName)
	}
}

// ChildrenByTag finds the given node's direct children with the specified
// tag name.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func (node *Node) ChildrenByTag(tagName string) (output <-chan *Node, exit chan interface{}) {
	return node.ChildrenMatching(predicateIsTag(tagName))
}

// DescendantsByTag finds the given node's descendants with the specified
// tag name.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func (node *Node) DescendantsByTag(tagName string) (output <-chan *Node, exit chan interface{}) {
	return node.DescendantsMatching(predicateIsTag(tagName))
}

func predicateAttrValueContains(attrKey, match string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.AttrValueContains(attrKey, match)
	}
}

// ChildrenByAttrValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func (node *Node) ChildrenByAttrValueContaining(attrKey, match string) (output <-chan *Node, exit chan interface{}) {
	return node.ChildrenMatching(predicateAttrValueContains(attrKey, match))
}

// DescendantsByAttrValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func (node *Node) DescendantsByAttrValueContaining(attrKey, match string) (output <-chan *Node, exit chan interface{}) {
	return node.DescendantsMatching(predicateAttrValueContains(attrKey, match))
}