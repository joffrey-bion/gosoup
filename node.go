package gosoup

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"strings"
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

// Parse returns the parse tree for the HTML from the given Reader. The input is
// assumed to be UTF-8 encoded.
func Parse(r io.Reader) (*Node, error) {
	hnode, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return WrapTree(hnode), nil
}

// WrapTree converts the given html.Node and all its children into an equivalent
// tree of gosoup.Node.
//
// The returned node has a nil Parent field, and nil NextSibling and PrevSibling
// fields. However, FirstChildren and LastChildren are populated, and the children
// themselves are completely linked with all their fields.
func WrapTree(hnode *html.Node) *Node {
	if hnode == nil {
		return nil
	}
	n := new(Node)

	// copy data
	n.DataAtom = hnode.DataAtom
	n.Data = hnode.Data
	n.Namespace = hnode.Namespace
	n.Attrs = make([]Attribute, len(hnode.Attr))
	for _, hattr := range hnode.Attr {
		n.Attrs = append(n.Attrs, Attribute(hattr))
	}

	// link to children nodes
	if hnode.FirstChild == nil {
		// the node has no children
		return n
	}

	n.FirstChild = WrapTree(hnode.FirstChild)
	n.FirstChild.Parent = n
	n.LastChild = n.FirstChild
	for hchild := hnode.FirstChild.NextSibling; hchild != nil; hchild = hchild.NextSibling {
		newChild := WrapTree(hchild)
		newChild.Parent = n
		newChild.PrevSibling = n.LastChild
		n.LastChild.NextSibling = newChild
		n.LastChild = newChild
	}
	return n
}

// WrapTree converts the given gosoup.Node and all its children into an equivalent
// tree of html.Node.
//
// The returned node has a nil Parent field, and nil NextSibling and PrevSibling
// fields. However, FirstChildren and LastChildren are populated, and the children
// themselves are completely linked with all their fields.
func UnwrapTree(node *Node) *html.Node {
	if node == nil {
		return nil
	}
	h := new(html.Node)

	// copy data
	h.DataAtom = node.DataAtom
	h.Data = node.Data
	h.Namespace = node.Namespace
	h.Attr = make([]html.Attribute, len(node.Attrs))
	for _, attr := range node.Attrs {
		h.Attr = append(h.Attr, html.Attribute(attr))
	}

	// link to children nodes
	if node.FirstChild == nil {
		// the node has no children
		return h
	}

	h.FirstChild = UnwrapTree(node.FirstChild)
	h.FirstChild.Parent = h
	h.LastChild = h.FirstChild
	for child := node.FirstChild.NextSibling; child != nil; child = child.NextSibling {
		newChild := UnwrapTree(child)
		newChild.Parent = h
		newChild.PrevSibling = h.LastChild
		h.LastChild.NextSibling = newChild
		h.LastChild = newChild
	}
	return h
}

// Root returns the root of the document containing this node.
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

// AttrOrDefault returns the value of the given attribute if this node
// has that attribute, otherwise returns defaultValue.
//
// If this node does not have the specified attribute, defaultValue is returned.
func (node *Node) AttrOrDefault(attrKey, defaultValue string) string {
	for _, a := range node.Attrs {
		if a.Key == attrKey {
			return a.Val
		}
	}
	return defaultValue
}

// HasAttrContaining returns true if the specified attribute's value
// contains the match string.
func (node *Node) HasAttrContaining(attrKey, match string) bool {
	return node.HasAttr(attrKey) && strings.Contains(node.Attr(attrKey), match)
}

// IsTag returns true if this node is a tag with the given name.
func (node *Node) IsTag(name string) bool {
	return node.Type == ElementNode && node.Data == name
}
