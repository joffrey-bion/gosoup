// A helper to explore the DOM of an HTML file
package gosoup

import (
	"golang.org/x/net/html"
	"strings"
)

// GetMatchingDirectChildren iterates over the given node's children, and sends the ones that match the given predicate over the returned channel.
func GetMatchingDirectChildren(node *html.Node, predicate func(node *html.Node) bool) chan *html.Node {
	channel := make(chan *html.Node)
	go func() {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if predicate(child) {
				channel <- child
			}
		}
	}()
	return channel
}

func GetDirectChildren(node *html.Node) chan *html.Node {
	trueFunc := func(node *html.Node) bool {
		return true
	}
	return GetMatchingDirectChildren(node, trueFunc)
}

func GetChildrenByAttributeValueContaining(node *html.Node, attrKey, attrValuePart string) chan *html.Node {
	trueFunc := func(node *html.Node) bool {
		return HasAttributeValueContaining(node, attrKey, attrValuePart)
	}
	return GetMatchingDirectChildren(node, trueFunc)
}

func HasAttributeValueContaining(node *html.Node, attrKey, attrValuePart string) bool {
	for _, a := range node.Attr {
		if a.Key == attrKey && strings.Contains(a.Val, attrValuePart) {
			return true
		}
	}
	return false
}

func GetDocCharset(doc *html.Node) string {
	// TODO
	return ""
}
