package gosoup

import (
	"golang.org/x/net/html"
	"io"
)

// Parse returns the parse tree for the HTML from the given Reader. The input is
// assumed to be UTF-8 encoded.
func Parse(r io.Reader) (*Node, error) {
	n, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return WrapTree(n), nil
}

// Render renders the parse tree n to the given writer.
//
// Rendering is done on a 'best effort' basis: calling Parse on the output of Render
// will always result in something similar to the original tree, but it is not
// necessarily an exact clone unless the original tree was 'well-formed'.
// 'Well-formed' is not easily specified; the HTML5 specification is complicated.
//
// Calling Parse on arbitrary input typically results in a 'well-formed' parse tree.
// However, it is possible for Parse to yield a 'badly-formed' parse tree. For
// example, in a 'well-formed' parse tree, no <a> element is a child of another <a>
// element: parsing "<a><a>" results in two sibling elements. Similarly, in a
// 'well-formed' parse tree, no <a> element is a child of a <table> element: parsing
// "<p><table><a>" results in a <p> with two sibling children; the <a> is reparented
// to the <table>'s parent. However, calling Parse on "<a><table><a>" does not return
// an error, but the result has an <a> element with an <a> child, and is therefore
// not 'well-formed'.
//
// Programmatically constructed trees are typically also 'well-formed', but it is
// possible to construct a tree that looks innocuous but, when rendered and re-parsed,
// results in a different tree. A simple example is that a solitary text node would
// become a tree containing <html>, <head> and <body> elements. Another example is
// that the programmatic equivalent of "a<head>b</head>c" becomes
// "<html><head><head/><body>abc</body></html>".
func Render(w io.Writer, n *Node) error {
	return html.Render(w, UnwrapTree(n))
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
	n.Type = NodeType(hnode.Type)
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
	h.Type = html.NodeType(node.Type)
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
