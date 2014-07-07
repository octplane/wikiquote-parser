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

func TestSample1(*testing.T) {
  content, err := ioutil.ReadFile("./tests/sample1.txt")
  if err != nil {
    panic(err)
  }
  cmds := Parse(string(content))
  fmt.Println(cmds)
}
