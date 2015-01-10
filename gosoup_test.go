package gosoup

import (
	//	"fmt"
	"strings"
	"testing"
)

const (
	HTML string = `<html>
	<head>
		<title>Your Title Here</title>
	</head>
	<body BGCOLOR="FFFFFF">
		<aside>
			<img src="clouds.jpg" align="BOTTOM">
		</aside>
		<hr>
		<a href="http://somegreatsite.com">Link Name</a> is a link to another nifty site
		<h1>This is a Header</h1>
		<h2>This is a Medium Header</h2>
		Send me mail at <a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.
		<p>
		This is a new paragraph!
		<p>
		<b>This is a new paragraph!</b>
		<br>
		<b><i>This is a new sentence without a paragraph break, in bold italics.</i></b>
		<hr>
	</body>
</html>`

	// HTML_CLEANED is what is actually read by the html package
	HTML_CLEANED string = `<html>
	<head>
		<title>Your Title Here</title>
	</head>
	<body bgcolor="FFFFFF">
		<aside>
			<img src="clouds.jpg" align="BOTTOM" />
		</aside>
		<hr />
		<a href="http://somegreatsite.com">Link Name</a>is a link to another nifty site
		<h1>This is a Header</h1>
		<h2>This is a Medium Header</h2>
		Send me mail at	<a href="mailto:support@yourcompany.com">support@yourcompany.com</a>.
		<p>This is a new paragraph!</p>
		<p>
			<b>This is a new paragraph!</b> <br /> 
			<b><i>This is a new sentence without a paragraph break, in bold italics.</i></b>
		</p>
		<hr />
	</body>
</html>`
)

func assert(t *testing.T, test bool, msg ...interface{}) {
	if !test {
		t.Fatal(msg)
	}
}

func assertEqualsWithMsg(t *testing.T, expected interface{}, value interface{}, msg ...interface{}) {
	assert(t, value == expected, msg)
}

func assertEquals(t *testing.T, expected interface{}, value interface{}) {
	assertEqualsWithMsg(t, value == expected, "assert failed: value is ", value, " expected: ", expected)
}

func assertNodeWithData(t *testing.T, ch <-chan *Node, data string) *Node {
	node, ok := <-ch
	assert(t, ok, "no node, expected '", data, "'")
	assertEqualsWithMsg(t, data, node.Data, "expected node '"+data+"', got '"+node.Data+"'")
	return node
}

func assertNoMoreNodes(t *testing.T, ch <-chan *Node) {
	_, ok := <-ch
	assert(t, !ok, "too many nodes")
}

func TestChildren(t *testing.T) {
	doc, err := Parse(strings.NewReader(HTML))
	if err != nil {
		t.Fatal(err)
	}

	ch := doc.Children().Nodes
	htmlNode := assertNodeWithData(t, ch, "html")
	assertNoMoreNodes(t, ch)

	ch = htmlNode.Children().Nodes
	head := assertNodeWithData(t, ch, "head")
	body := assertNodeWithData(t, ch, "body")
	assertNoMoreNodes(t, ch)

	ch = head.Children().Nodes
	title := assertNodeWithData(t, ch, "title")
	assertNoMoreNodes(t, ch)

	ch = body.Children().Nodes
	aside := assertNodeWithData(t, ch, "aside")
	assertNodeWithData(t, ch, "hr")
	anchor := assertNodeWithData(t, ch, "a")
	assertNodeWithData(t, ch, "is a link to another nifty site")
	assertNodeWithData(t, ch, "h1")
	assertNodeWithData(t, ch, "h2")
	assertNodeWithData(t, ch, "Send me mail at")
	assertNodeWithData(t, ch, "a")
	assertNodeWithData(t, ch, ".")
	assertNodeWithData(t, ch, "p")
	assertNodeWithData(t, ch, "p")
	assertNodeWithData(t, ch, "hr")
	assertNoMoreNodes(t, ch)

	ch = title.Children().Nodes
	assertNodeWithData(t, ch, "Your Title Here")
	assertNoMoreNodes(t, ch)

	ch = aside.Children().Nodes
	assertNodeWithData(t, ch, "img")
	assertNoMoreNodes(t, ch)

	ch = anchor.Children().Nodes
	assertNodeWithData(t, ch, "Link Name")
	assertNoMoreNodes(t, ch)
}

func TestDescendants(t *testing.T) {
	doc, err := Parse(strings.NewReader(HTML))
	if err != nil {
		t.Fatal(err)
	}

	ch := doc.Descendants().Nodes
	assertNodeWithData(t, ch, "html")
	assertNodeWithData(t, ch, "head")
	assertNodeWithData(t, ch, "title")
	assertNodeWithData(t, ch, "Your Title Here")
	assertNodeWithData(t, ch, "body")
	assertNodeWithData(t, ch, "aside")
	assertNodeWithData(t, ch, "img")
	assertNodeWithData(t, ch, "hr")
	assertNodeWithData(t, ch, "a")
	assertNodeWithData(t, ch, "Link Name")
	assertNodeWithData(t, ch, "is a link to another nifty site")
	assertNodeWithData(t, ch, "h1")
	assertNodeWithData(t, ch, "This is a Header")
	assertNodeWithData(t, ch, "h2")
	assertNodeWithData(t, ch, "This is a Medium Header")
	assertNodeWithData(t, ch, "Send me mail at")
	assertNodeWithData(t, ch, "a")
	assertNodeWithData(t, ch, "support@yourcompany.com")
	assertNodeWithData(t, ch, ".")
	assertNodeWithData(t, ch, "p")
	assertNodeWithData(t, ch, "This is a new paragraph!")
	assertNodeWithData(t, ch, "p")
	assertNodeWithData(t, ch, "b")
	assertNodeWithData(t, ch, "This is a new paragraph!")
	assertNodeWithData(t, ch, "br")
	assertNodeWithData(t, ch, "b")
	assertNodeWithData(t, ch, "i")
	assertNodeWithData(t, ch, "This is a new sentence without a paragraph break, in bold italics.")
	assertNodeWithData(t, ch, "hr")
	assertNoMoreNodes(t, ch)
}

func TestChildrenMatching(t *testing.T) {
	doc, err := Parse(strings.NewReader(HTML))
	if err != nil {
		t.Fatal(err)
	}
	
	containsH := func(n *Node) bool {
		return strings.ContainsAny(n.Data, "hH")
	}

	ch := doc.Children().Nodes
	htmlNode := assertNodeWithData(t, ch, "html")
	assertNoMoreNodes(t, ch)

	ch = htmlNode.Children().Nodes
	head := assertNodeWithData(t, ch, "head")
	body := assertNodeWithData(t, ch, "body")
	assertNoMoreNodes(t, ch)

	ch = head.ChildrenMatching(containsH).Nodes
	assertNoMoreNodes(t, ch)

	ch = body.ChildrenMatching(containsH).Nodes
	assertNodeWithData(t, ch, "hr")
	assertNodeWithData(t, ch, "is a link to another nifty site")
	assertNodeWithData(t, ch, "h1")
	assertNodeWithData(t, ch, "h2")
	assertNodeWithData(t, ch, "hr")
	assertNoMoreNodes(t, ch)
}

func TestDescendantsMatching(t *testing.T) {
	doc, err := Parse(strings.NewReader(HTML))
	if err != nil {
		t.Fatal(err)
	}
	
	containsH := func(n *Node) bool {
		return strings.ContainsAny(n.Data, "hH")
	}

	ch := doc.DescendantsMatching(containsH).Nodes
	assertNodeWithData(t, ch, "html")
	assertNodeWithData(t, ch, "head")
	assertNodeWithData(t, ch, "Your Title Here")
	assertNodeWithData(t, ch, "hr")
	assertNodeWithData(t, ch, "is a link to another nifty site")
	assertNodeWithData(t, ch, "h1")
	assertNodeWithData(t, ch, "This is a Header")
	assertNodeWithData(t, ch, "h2")
	assertNodeWithData(t, ch, "This is a Medium Header")
	assertNodeWithData(t, ch, "This is a new paragraph!")
	assertNodeWithData(t, ch, "This is a new paragraph!")
	assertNodeWithData(t, ch, "This is a new sentence without a paragraph break, in bold italics.")
	assertNodeWithData(t, ch, "hr")
	assertNoMoreNodes(t, ch)
}
