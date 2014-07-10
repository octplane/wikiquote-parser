package main

import (
  "bufio"
  "fmt"
  "github.com/octplane/wikiquote-parser"
  "launchpad.net/xmlpath"
  "log"
  "os"
)

func main() {
  pageXPath := xmlpath.MustCompile("/mediawiki/page")
  textXPath := xmlpath.MustCompile(("revision/text"))
  titleXPath := xmlpath.MustCompile("title")

  //fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
  fi, err := os.Open("sample.xml")

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
  // ns := make(wikiquote_parser.Nodes, 0)
  iter := pageXPath.Iter(root)
  for iter.Next() {
    page := iter.Node()
    content, _ := textXPath.String(page)
    title, _ := titleXPath.String(page)
    fmt.Println(title)
    fmt.Println(wikiquote_parser.Parse(wikiquote_parser.Tokenize(content)))
  }
}
