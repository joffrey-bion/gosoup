package gosoup

import (
	"golang.org/x/net/html"
	"strings"
)

// HasAttr returns true if the given node has the specified attribute.
func HasAttr(node *html.Node, attrKey string) bool {
	for _, a := range node.Attr {
		if a.Key == attrKey {
			return true
		}
	}
	return false
}

// GetAttrValue returns the value of the given attribute of the given node.
//
// If the given node does not have the specified attribute, this function panics.
func GetAttrValue(node *html.Node, attrKey string) string {
	for _, a := range node.Attr {
		if a.Key == attrKey {
			return a.Val
		}
	}
	panic("no such attribute '" + attrKey + "'")
}

// GetAttrValue returns the value of the given attribute of the given node.
//
// If the given node does not have the specified attribute, defaultValue is returned.
func GetAttrValueOrDefault(node *html.Node, attrKey, defaultValue string) string {
	for _, a := range node.Attr {
		if a.Key == attrKey {
			return a.Val
		}
	}
	return defaultValue
}

// HasAttrValueContaining returns true if the specified attribute's value
// contains the match string for the given node.
func HasAttrValueContaining(node *html.Node, attrKey, match string) bool {
	return HasAttr(node, attrKey) && strings.Contains(GetAttrValue(node, attrKey), match)
}

func predicateHasAttrValueContaining(attrKey, match string) func(node *html.Node) bool {
	return func(node *html.Node) bool {
		return HasAttrValueContaining(node, attrKey, match)
	}
}

// GetChildrenByAttrValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func GetChildrenByAttrValueContaining(node *html.Node, attrKey, match string) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingChildren(node, predicateHasAttrValueContaining(attrKey, match))
}

// GetChildrenByAttrValueContaining finds the given node's direct children
// that have attributes whose value contains the match string.
//
// The caller should send anything into the exit channel to indicate that no 
// more nodes will be read, unless he finishes the loop.
func GetDescendantsByAttrValueContaining(node *html.Node, attrKey, match string) (output <-chan *html.Node, exit chan interface{}) {
	return GetMatchingDescendants(node, predicateHasAttrValueContaining(attrKey, match))
}