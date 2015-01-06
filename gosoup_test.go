package gosoup

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

const (
	HTML string = `<html>
	<head>
		<title>Your Title Here</title>
	</head>
	<body BGCOLOR="FFFFFF">
		<center>
			<img src="clouds.jpg" align="BOTTOM">
		</center>
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
)

func TestGetChildren(t *testing.T) {
	fmt.Println(HTML)
	doc, _ := html.Parse(strings.NewReader(HTML))
	descendants, _ := GetMatchingDescendants(doc, func(n *html.Node) bool { return strings.Contains(n.Data, "h") })
	i := 0
	fmt.Println("Descendants:")
	for child := range descendants {
		fmt.Printf("%d - '%s'\n", i, child.Data)
		i++
	}
	t.SkipNow()
}
