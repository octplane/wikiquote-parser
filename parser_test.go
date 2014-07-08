package wikiquote_parser

import (
	"bufio"
	"fmt"
	"launchpad.net/xmlpath"
	"log"
	"os"
	"testing"
)

func testMain(*testing.T) {
	pageXPath := xmlpath.MustCompile("/mediawiki/page")
	textXPath := xmlpath.MustCompile(("revision/text"))
	titleXPath := xmlpath.MustCompile("title")

	//fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
	fi, err := os.Open("sample1.xml")

	if err != nil {

		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	// make a read buffer
	r := bufio.NewReader(fi)

	root, err := xmlpath.Parse(r)
	if err != nil {
		log.Fatal(err)
	}
	commands := make([]Command, 0)
	iter := pageXPath.Iter(root)
	for iter.Next() {
		page := iter.Node()
		content, _ := textXPath.String(page)
		title, _ := titleXPath.String(page)
		commands = append(commands, markupExtractor(title, content)...)
	}
	// ExtractStats(commands)
	ExtractQuotes(commands)
}

func TestLinkParser(t *testing.T) {
	lnk := "foo"
	text := "Link to foo "
	tree := Parse(Tokenize(fmt.Sprintf("[[%s|%s]]", lnk, text)))
	if len(tree) != 1 {
		t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
	}
	if tree[0].typ != nodeLink {
		t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
	}
	if tree[0].params["link"] != lnk {
		t.Error("Unexpected link, expected link to %q, got %q", lnk, tree[0].params["link"])
	}
	if tree[0].params["text"] != text {
		t.Error("Unexpected link, expected text link to %q, got %q", text, tree[0].params["text"])
	}
}

func TestTemplate(t *testing.T) {
	temp := "citation"

	tree := Parse(Tokenize(fmt.Sprintf("{{%s}}", temp)))
	if len(tree) != 1 {
		t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
	}

	if tree[0].typ != nodeTemplate {
		t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
	}
	if tree[0].params["name"] != temp {
		t.Errorf("Unexpected name, expected name to %q, got %q", temp, tree[0].params["name"])
	}
}

func TestTemplate2(t *testing.T) {
	temp := "citation"
	txt := "Tant va la cruche à l'eau qu'à la fin tu me les brises."

	tree := Parse(Tokenize(fmt.Sprintf("{{%s|%s}}", temp, txt)))
	if len(tree) != 1 {
		t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
	}

	if tree[0].typ != nodeTemplate {
		t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
	}
	if tree[0].params["name"] != temp {
		t.Errorf("Unexpected name, expected name to %q, got %q", temp, tree[0].params["name"])
	}
	if tree[0].val != txt {
		t.Errorf("Unexpected value to %q, got %q", txt, tree[0].val)
	}
}

func TestTemplate3(t *testing.T) {
	temp := "citation"
	txt := "Tant va la cruche à l'eau qu'à la fin tu me les brises."
	aut := "Les Inconnus"

	tree := Parse(Tokenize(fmt.Sprintf("{{%s|citation=%s|author=%s}}", temp, txt, aut)))
	if len(tree) != 1 {
		t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
	}

	if tree[0].typ != nodeTemplate {
		t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
	}
	if tree[0].params["name"] != temp {
		t.Errorf("Unexpected name, expected name: wanted %q, got %q", temp, tree[0].params["name"])
	}
	if tree[0].params["citation"] != txt {
		t.Errorf("Unexpected citation params: wanted %q, got %q", txt, tree[0].params["citation"])
	}
	if tree[0].params["author"] != aut {
		t.Errorf("Unexpected author param: wanted %q, got %q", aut, tree[0].params["author"])
	}
}
