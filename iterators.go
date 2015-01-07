package gosoup

import (
	"golang.org/x/net/html"
	"strings"
	"sync"
)

const (
	BLANK string = " \t\n\r"
)

func forward(in <-chan interface{}, out chan interface{}) {
	for e := range in {
		out <- e
	}
}

func forwardNodes(in <-chan *html.Node, out chan *html.Node) {
	for e := range in {
		out <- e
	}
}

// notBlank returns true if the node's data is not full of blank space
func notBlank(node *html.Node) bool {
	return strings.Trim(node.Data, BLANK) != ""
}

func clean(node *html.Node) *html.Node {
	if node.Type == html.TextNode {
		node.Data = strings.Trim(node.Data, BLANK)
	}
	return node
}

// iterateOnDescendants finds the given node's descendants that match the
// given predicate, and sends them into the output channel.
//
// If recursive is false, only direct children are considered.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func iterateOnDescendants(node *html.Node, predicate func(node *html.Node) bool, recursive bool) (output <-chan *html.Node, exit chan interface{}) {
	if node == nil {
		panic("iterateOnDescendants: null input node")
	}
	out := make(chan *html.Node, 20)
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
					capturedChild := child
					wg.Add(1)
					go func() {
						defer wg.Done()
						in, subexit := iterateOnDescendants(capturedChild, predicate, recursive)
						go forward(exit, subexit)
						forwardNodes(in, out)
					}()
				}
			}
		}
		wg.Wait()
		close(out)
	}()
	return out, exit
}

// GetMatchingChildren finds the given node's direct children that match the
// given predicate, and sends them into the output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func GetMatchingChildren(node *html.Node, predicate func(node *html.Node) bool) (output <-chan *html.Node, exit chan interface{}) {
	return iterateOnDescendants(node, predicate, false)
}

// GetMatchingDescendants finds the given node's descendants that match the
// given predicate, and sends them into the output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func GetMatchingDescendants(node *html.Node, predicate func(node *html.Node) bool) (output <-chan *html.Node, exit chan interface{}) {
	return iterateOnDescendants(node, predicate, true)
}

// GetChildren finds the given node's direct children, and sends them into the
// output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func GetChildren(node *html.Node) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingChildren(node, notBlank)
}

// GetDescendants finds the given node's descendants, and sends them into the
// output channel.
//
// The caller should send anything into the exit channel to indicate that no
// more nodes will be read, unless he finishes the loop.
func GetDescendants(node *html.Node) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingDescendants(node, notBlank)
}
