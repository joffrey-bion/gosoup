/*
GoSoup is a helper to explore the DOM of an HTML file.

The iteration methods provided here return 2 channels: an output channel, to read
the iterated elements from, and an exit channel.

The caller should send something (the value does not matter) on the exit
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
	"errors"
	"golang.org/x/net/html"
	"strings"
)

// Root returns the root of the document containing the given node.
func Root(node *html.Node) *html.Node {
	for node.Parent != nil {
		node = node.Parent
	}
	return node
}

// First retrieves the first node from the given output channel, and takes
// care of the cleaning through the exit channel.
//
// This function can be particularly useful when combined with the iterating
// functions of GoSoup:
//
//     firstChild := gosoup.First(gosoup.GetChildren(node))
//
//     firstMyClassDescendant := gosoup.First(gosoup.GetDescendantsByAttributeValueContaining(node, "class", "myClass"))
//
// No need to take care of channels here.
func First(output <-chan *html.Node, exit chan interface{}) *html.Node {
	node := <-output
	exit <- true
	return node
}

// Collect gathers all nodes from the given output channel in a slice.
//
// This function can be particularly useful when combined with the iterating
// functions of GoSoup.
func Collect(output <-chan *html.Node, exit chan interface{}) []*html.Node {
	var list []*html.Node
	for node := range output {
		list = append(list, node)
	}
	return list
}

func GetDocContentType(node *html.Node) (string, error) {
	root := Root(node)
	head := First(GetDescendantsByTag(root, "head"))
	if head == nil {
		return "", errors.New("GetDocCharset: head not found")
	}
	meta := First(GetDescendantsByAttrValueContaining(head, "content", "charset="))
	if meta == nil {
		return "", errors.New("GetDocCharset: meta not found")
	}
	return GetAttrValue(meta, "content"), nil
}

func GetDocCharset(node *html.Node) (string, error) {
	content, err := GetDocContentType(node)
	if err != nil {
		return "", err
	}
	charsetWithBullshit := strings.Split(content, "charset=")[1]
	// trim potential trailing stuff
	charsetWithoutBullshit := strings.Split(charsetWithBullshit, ";")[0]
	return strings.Split(charsetWithoutBullshit, " ")[0], nil // just in case
}
