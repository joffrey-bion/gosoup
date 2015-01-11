/*
GoSoup allows to parse HTML content and browse the produced tree. It wraps the
golang.org/x/net/html package, providing helpful methods.

Iterators

The most interesting functions provided by GoSoup are the iterator functions.
These functions return a NodeIterator object, which can then be filtered and/or
mapped and the nodes can then be collected into a slice.

Read about the NodeIterator type and its methods to get an idea of how flexible
and powerful it is. Should that power not suffice, you can have full control of
what happens by directly using the internals of the NodeIterator type.

Raw Use Of Iterators

The NodeIterator objects contains a read-only channel to read the nodes from.
The DOM tree mustn't be modified while reading from the iterator.

The caller should close the iterator if he does not exhaust the output channel.
This prevents the internal goroutines from blocking forever if there are more
elements to send. For instance:

    it := node.Children()
    for child := range it.Nodes {
        if (something) {
            // we break early from the loop, close the iterator
            it.Close()
            break
        }
        doStuffWith(child)
    }

If the loop ends normally, there is no need to close the iterator.
Gosoup will close the output channel when all elements have been sent in order
to end the caller's loop.
*/
package gosoup

import (
	"errors"
	"strings"
)

// GetDocContentType returns the content-type string description taken from a <meta>
// element in the <head> part of the HTML tree.
func GetDocContentType(node *Node) (string, error) {
	root := node.Root() // moves to the document node
	head := root.DescendantsByTag("head").First()
	if head == nil {
		return "", errors.New("GetDocContentType: head not found")
	}
	meta := head.DescendantsByAttrValueContaining("content", "charset=").First()
	if meta == nil {
		return "", errors.New("GetDocContentType: no head child with 'content' attribute containing 'charset'")
	}
	return meta.Attr("content"), nil
}

// GetDocCharset returns the charset contained in the content-type string description
// taken from a <meta> element in the <head> part of the HTML tree.
func GetDocCharset(node *Node) (string, error) {
	content, err := GetDocContentType(node)
	if err != nil {
		return "", err
	}
	charsetWithBullshit := strings.Split(content, "charset=")[1]
	// trim potential trailing stuff
	charsetWithoutBullshit := strings.Split(charsetWithBullshit, ";")[0]
	return strings.Split(charsetWithoutBullshit, " ")[0], nil // just in case
}

func notBlank(n *Node) bool {
	return !n.IsBlankText()
}

func clean(n *Node) *Node {
	n.TrimTextData()
	return n
}

// Children returns an iterator on this node's direct children.
func (node *Node) Children() NodeIterator {
	return node.TreeIterator(false).Filter(notBlank).Map(clean)
}

// Descendants returns an iterator on this node's descendants, in depth-first order.
func (node *Node) Descendants() NodeIterator {
	return node.TreeIterator(true).Filter(notBlank).Map(clean)
}

// ChildrenMatching returns an iterator on this node's direct children that match
// the given predicate.
func (node *Node) ChildrenMatching(predicate func(node *Node) bool) NodeIterator {
	return node.Children().Filter(predicate)
}

// DescendantsMatching returns an iterator on this node's descendants that match
// the given predicate, in depth-first order.
func (node *Node) DescendantsMatching(predicate func(node *Node) bool) NodeIterator {
	return node.Descendants().Filter(predicate)
}

func predicateIsTag(tagName string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.IsTag(tagName)
	}
}

// ChildrenByTag returns an iterator on this node's direct children with the specified
// tag name.
func (node *Node) ChildrenByTag(tagName string) NodeIterator {
	return node.ChildrenMatching(predicateIsTag(tagName))
}

// DescendantsByTag returns an iterator on this node's descendants with the specified
// tag name, in depth-first order.
func (node *Node) DescendantsByTag(tagName string) NodeIterator {
	return node.DescendantsMatching(predicateIsTag(tagName))
}

func predicateAttrValueContains(attrKey, match string) func(node *Node) bool {
	return func(node *Node) bool {
		return node.AttrValueContains(attrKey, match)
	}
}

// ChildrenByAttrValueContaining returns an iterator on this node's direct children
// that have attributes whose value contains the match string.
func (node *Node) ChildrenByAttrValueContaining(attrKey, match string) NodeIterator {
	return node.ChildrenMatching(predicateAttrValueContains(attrKey, match))
}

// DescendantsByAttrValueContaining returns an iterator on this node's descendants
// that have attributes whose value contains the match string, in depth-first order.
func (node *Node) DescendantsByAttrValueContaining(attrKey, match string) NodeIterator {
	return node.DescendantsMatching(predicateAttrValueContains(attrKey, match))
}
