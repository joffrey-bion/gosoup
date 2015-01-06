/*
GoSoup is a helper to explore the DOM of an HTML file.

The iteration methods provided here return 2 channels: an output channel, to read
the iterated elements from, and an exit channel.

The caller should send something (the boolean value does not matter) on the exit
channel if he does not exhaust the output channel. This prevents the internal
goroutines from blocking forever if there are more elements to send. For instance:

    children, exit := GetChildren(node)
    for child := range children {
        if (something) {
            // we break early from the loop, notify via exit channel
            exit <- true
            break
        }
        doNormalStuff()
    }

If the loop ends normally, there is no need to send anything on the exit channel.
There is also no need to close the channels. Gosoup will close the output channel
when all elements have been sent in order to end the caller's loop.

The elements are sent to the output channel in an unspecified order, unless
explicitely stated otherwise.

The DOM tree mustn't be modified while reading from the returned channel, because
the iteration is done concurrently.
*/
package gosoup

import (
	"golang.org/x/net/html"
	"strings"
	"sync"
)

func forwardBools(in <-chan bool, out chan bool) {
	for e := range in {
		out <- e
	}
}

func forwardNodes(in <-chan *html.Node, out chan *html.Node) {
	for e := range in {
		out <- e
	}
}

// GetMatchingChildren finds the given node's direct children that match the
// given predicate, and sends them into the output channel.
func GetMatchingChildren(node *html.Node, predicate func(node *html.Node) bool) (<-chan *html.Node, chan bool) {
	out := make(chan *html.Node)
	exit := make(chan bool)
	go func() {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if predicate(child) {
				select {
				case <-exit:
					close(out)
					return
				default:
					out <- child
				}
			}
		}
		close(out)
	}()
	return out, exit
}

// GetChildren finds the given node's direct children, and sends them into the
// output channel.
func GetChildren(node *html.Node) (<-chan *html.Node, chan bool) {
	trueFunc := func(node *html.Node) bool {
		return true
	}
	return GetMatchingChildren(node, trueFunc)
}

// GetMatchingDescendants finds the given node's descendants that match the
// given predicate, and sends them into the output channel.
func GetMatchingDescendants(node *html.Node, predicate func(node *html.Node) bool) (<-chan *html.Node, chan bool) {
	out := make(chan *html.Node, 20)
	exit := make(chan bool, 1)
	go func() {
		var wg sync.WaitGroup
		children, _ := GetChildren(node)
		for child := range children {
			select {
			case <-exit:
				wg.Wait()
				close(out)
				return
			default:
				if predicate(child) {
					out <- child
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					in, subexit := GetMatchingDescendants(child, predicate)
					go forwardBools(exit, subexit)
					forwardNodes(in, out)
				}()
			}
		}
		wg.Wait()
		close(out)
	}()
	return out, exit
}

// GetDescendants finds the given node's descendants, and sends them into the
// output channel.
func GetDescendants(node *html.Node) (<-chan *html.Node, chan bool) {
	trueFunc := func(node *html.Node) bool {
		return true
	}
	return GetMatchingDescendants(node, trueFunc)
}

// GetChildrenByAttributeValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
func GetChildrenByAttributeValueContaining(node *html.Node, attrKey, match string) (<-chan *html.Node, chan bool) {
	trueFunc := func(node *html.Node) bool {
		return HasAttributeValueContaining(node, attrKey, match)
	}
	return GetMatchingChildren(node, trueFunc)
}

// HasAttributeValueContaining returns true if the specified attribute's value
// contains the match string for the given node.
func HasAttributeValueContaining(node *html.Node, attrKey, match string) bool {
	for _, a := range node.Attr {
		if a.Key == attrKey && strings.Contains(a.Val, match) {
			return true
		}
	}
	return false
}

func GetDocCharset(node *html.Node) string {
	for ; node.Parent != nil; node = node.Parent {
	}
	// TODO
	return ""
}
