/*
GoSoup allows to parse HTML content and browse the produced tree. It wraps the
golang.org/x/net/html package, providing helpful methods.

Iterators

The iteration methods provided here return an Iterator object, containing an
output channel, to read the iterated elements from.
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

func GetDocContentType(node *Node) (string, error) {
	root := node.Root()
	head := root.DescendantsByTag("head").First()
	if head == nil {
		return "", errors.New("GetDocCharset: head not found")
	}
	meta := head.DescendantsByAttrValueContaining("content", "charset=").First()
	if meta == nil {
		return "", errors.New("GetDocCharset: meta not found")
	}
	return meta.Attr("content"), nil
}

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
