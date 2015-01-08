/*
GoSoup is a helper to explore the DOM of an HTML file. It wraps the golang.org/x/net/html
package, providing helpful methods.

The iteration methods provided here return 2 channels: an output channel, to read
the iterated elements from, and an exit channel.

The caller should send something (the value does not matter) on the exit
channel if he does not exhaust the output channel. This prevents the internal
goroutines from blocking forever if there are more elements to send. For instance:

    children, exit := node.Children()
    for child := range children {
        if (something) {
            // we break early from the loop, notify GoSoup via the exit channel
            exit <- true
            break
        }
        doStuffWith(child)
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
	"strings"
)

func GetDocContentType(node *Node) (string, error) {
	root := node.Root()
	head := First(root.DescendantsByTag("head"))
	if head == nil {
		return "", errors.New("GetDocCharset: head not found")
	}
	meta := First(head.DescendantsByAttrValueContaining("content", "charset="))
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
