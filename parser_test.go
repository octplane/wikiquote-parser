package wikiquote_parser

import (
  "bufio"
  "fmt"
  "io/ioutil"
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

func TestSample1(t *testing.T) {
  content, err := ioutil.ReadFile("./tests/sample1.txt")
  if err != nil {
    panic(err)
  }
  tokens := Tokenize(string(content))
  ex1 := "==== [[Léon Blum]] ====\n"
  if tokens[0].val != ex1 && tokens[0].typ != itemText {
    t.Errorf("Unexpected text block. Expected Text: %q, got %s", ex1, tokens[0].String())
  }

  tree := Parse(tokens)
  for _, n := range tree {
    fmt.Println(n.String())
  }
}
