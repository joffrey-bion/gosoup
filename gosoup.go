// A helper to explore the DOM of an HTML file
package gosoup

import (
	"golang.org/x/net/html"
	"strings"
    "sync"
)

func forward(in <-chan *html.Node, out chan *html.Node) {
	for e := range in {
		out <- e
	}
}

// GetMatchingChildren finds the given node's direct children that match the
// given predicate, and sends them over the returned channel.
// The order is unspecified.
// The DOM tree mustn't be modified while reading from the returned channel,
// because the iteration is done concurrently.
func GetMatchingChildren(node *html.Node, predicate func(node *html.Node) bool) <-chan *html.Node {
	out := make(chan *html.Node)
	go func() {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if predicate(child) {
				out <- child
			}
		}
		close(out)
	}()
	return out
}

// GetChildren finds the given node's direct children, and sends them over the
// returned channel.
// The order is unspecified.
// The DOM tree mustn't be modified while reading from the returned channel,
// because the iteration is done concurrently.
func GetChildren(node *html.Node) <-chan *html.Node {
	trueFunc := func(node *html.Node) bool {
		return true
	}
	return GetMatchingChildren(node, trueFunc)
}

// GetMatchingDescendents finds the given node's descendents that match the
// given predicate, and sends them over the returned channel.
// The order is unspecified.
// The DOM tree mustn't be modified while reading from the returned channel,
// because the iteration is done concurrently.
func GetMatchingDescendents(node *html.Node, predicate func(node *html.Node) bool) <-chan *html.Node {
	out := make(chan *html.Node, 20)
	go func() {
		var wg sync.WaitGroup
		for child := range GetChildren(node) {
			if predicate(child) {
				out <- child
			}
			go func() {
				wg.Add(1)
				defer wg.Done()
				forward(GetMatchingDescendents(child, predicate), out)
			}()
		}
		wg.Wait()
		close(out)
	}()
	return out
}

// GetChildrenByAttributeValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
// The order is unspecified.
// The DOM tree mustn't be modified while reading from the returned channel,
// because the iteration is done concurrently.
func GetChildrenByAttributeValueContaining(node *html.Node, attrKey, match string) <-chan *html.Node {
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

func GetDocCharset(doc *html.Node) string {
	// TODO
	return ""
}
