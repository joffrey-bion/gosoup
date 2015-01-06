package gosoup

import (
	"golang.org/x/net/html"
)

// IsTag returns true if the given node is a tag with the given name.
func IsTag(node *html.Node, name string) bool {
	return node.Type == html.ElementNode && node.Data == name
}

func predicateIsTag(tagName string) func(node *html.Node) bool {
	return func(node *html.Node) bool {
		return IsTag(node, tagName)
	}
}

// GetChildrenByTag finds the given node's direct children with the specified
// tag name.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func GetChildrenByTag(node *html.Node, tagName string) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingChildren(node, predicateIsTag(tagName))
}

// GetDescendantsByTag finds the given node's descendants with the specified
// tag name.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func GetDescendantsByTag(node *html.Node, tagName string) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingDescendants(node, predicateIsTag(tagName))
}