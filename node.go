package gosoup

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

const (
	blank string = " \t\n\r"
)

type Attribute html.Attribute

type NodeType html.NodeType

const (
	ErrorNode NodeType = iota
	TextNode
	DocumentNode
	ElementNode
	CommentNode
	DoctypeNode
)

type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type      NodeType
	DataAtom  atom.Atom
	Data      string
	Namespace string
	Attrs     []Attribute
}

// Root returns the root of the tree containing this node, namely the document node.
func (node *Node) Root() *Node {
	for node.Parent != nil {
		node = node.Parent
	}
	return node
}

// HasAttr returns true if this node has the specified attribute.
func (node *Node) HasAttr(attrKey string) bool {
	for _, a := range node.Attrs {
		if a.Key == attrKey {
			return true
		}
	}
	return false
}

// Attr returns the value of the given attribute.
//
// If this node does not have the specified attribute, this function panics.
func (node *Node) Attr(attrKey string) string {
	for _, a := range node.Attrs {
		if a.Key == attrKey {
			return a.Val
		}
	}
	panic("no such attribute '" + attrKey + "'")
}

// AttrOrDefault returns the value of the given attribute, or defaultValue if this
// node does not have that attribute.
func (node *Node) AttrOrDefault(attrKey, defaultValue string) string {
	for _, a := range node.Attrs {
		if a.Key == attrKey {
			return a.Val
		}
	}
	return defaultValue
}

// AttrValueContains returns true if this node has the given attribute and the value
// of that attribute contains the match string.
func (node *Node) AttrValueContains(attrKey, match string) bool {
	return node.HasAttr(attrKey) && strings.Contains(node.Attr(attrKey), match)
}

// IsTag returns true if this node is a tag with the given name.
func (node *Node) IsTag(name string) bool {
	return node.Type == ElementNode && node.Data == name
}

// TrimTextData trims leading and trailing whitespace if this node is a TextNode.
func (node *Node) TrimTextData() {
	if node.Type == TextNode {
		node.Data = strings.Trim(node.Data, blank)
	}
}

// IsBlankText returns true if this node is a TextNode full of blank space.
func (node *Node) IsBlankText() bool {
	return node.Type == TextNode && strings.Trim(node.Data, blank) == ""
}
