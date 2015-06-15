package main

import (
  "bufio"
  "crypto/sha1"
  "encoding/csv"
  "flag"
  "fmt"
  "github.com/golang/glog"
  . "github.com/octplane/wikiquote-parser"
  "github.com/octplane/wikiquote-parser/domenech/internals"
  "io"
  "launchpad.net/xmlpath"
  "os"
  "strconv"
  "strings"
  "time"
)

type Command struct {
  Cmd            string
  PageTitle      string
  Arguments      []string
  NamedArguments map[string]string
}

func (cmd *Command) ArgOrEmpty(key string) string {
  return cmd.ArgOrElse(key, "")
}

func (cmd *Command) ArgOrElse(key string, def string) string {
  ret, has := cmd.NamedArguments[key]
  if !has {
    return def
  }
  return ret
}

type nodeType int

const (
  quote = nodeType(iota)
  source
  link
  unknownType
)

func normalizedType(n Node) nodeType {
  if n.Typ == NodeTemplate {
    switch n.StringParam("name") {
    case "citation", "Citation":
      return quote
    case "Réf Livre", "Réf Pub", "Réf Article", "Réf Film":
      return source

    default:
      return unknownType
    }
  } else if n.Typ == NodeLink {
    if strings.Index(n.StringParamOrEmpty("link"), "Catégorie:") == 0 {
      return link
    }
  }
  return unknownType
}

type QuoteNode struct {
  source Node
  quote  Node
}

func (q *QuoteNode) nonEmpty() bool {
  return q.source.Typ != NodeEmpty && q.quote.Typ != NodeEmpty
}

func (q *QuoteNode) title() string {
  return q.source.StringParamOrEmpty("titre")
}

func (q *QuoteNode) isbn() string {
  return q.source.StringParamOrEmpty("ISBN")
}

func (q *QuoteNode) authorText() string {
  return q.source.StringParamOrEmpty("auteur")
}

var MAX_QUOTE_LENGTH = 100

func (q *QuoteNode) valid() bool {
  valid := q.nonEmpty() && q.authorText() != "" && len(q.quoteString()) < MAX_QUOTE_LENGTH
  if !valid {
    reason := ""
    if !q.nonEmpty() {
      reason = "quote is empty"
    } else if q.authorText() == "" {
      reason = "author text is empty"
    } else if len(q.quoteString()) >= MAX_QUOTE_LENGTH {
      reason = fmt.Sprintf("quote has length %d\n", len(q.quoteString()))
    }
    glog.V(2).Infof("%s is invalid because %s", q.StringRepresentation(""), reason)

  }
  return valid
}

func (q *QuoteNode) quoteString() string {
  if len(q.quote.Params) == 1 {
    return q.quote.Params[0].StringRepresentation()
  } else {
    return q.quote.StringParamOrEmpty("citation")
  }
  return ""
}

func (q *QuoteNode) StringRepresentation(category string) string {

  var authortext string

  quotetext := q.quoteString()
  authortext = q.authorText()
  title := q.title()
  isbn := q.isbn()

  return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", category, isbn, authortext, title, quotetext)
}

func ExtractQuoteNodes(nodes Nodes, theme string, id int) {
  var q QuoteNode = QuoteNode{source: EmptyNode(), quote: EmptyNode()}
  count := 0

  for _, node := range nodes {
    switch normalizedType(node) {
    case link:
      catName := node.StringParamOrEmpty("link")[11:]
      categ := internals.Category{Text: catName}
      _ = categ
    case quote:
      q.quote = node
    case source:
      q.source = node
    }
    if q.nonEmpty() {
      count += 1

      quote := q.quoteString()
      author := q.authorText()
      title := q.title()
      isbn := q.isbn()

      h := sha1.New()

      io.WriteString(h, quote)
      sha1 := fmt.Sprintf("%x", h.Sum(nil))

      quoteWriter.Write([]string{sha1, strconv.Itoa(id), theme, quote, author, title, isbn})
    }
  }
}

var (
  pageXPath   *xmlpath.Path
  pageIdXPath *xmlpath.Path
  textXPath   *xmlpath.Path
  titleXPath  *xmlpath.Path
)

var pageChannel = make(chan *xmlpath.Node)

var quoteWriter *csv.Writer

const layout = "20060102_1504"

func main() {

  flag.Parse()
  pageXPath = xmlpath.MustCompile("/mediawiki/page")
  pageIdXPath = xmlpath.MustCompile("id")
  textXPath = xmlpath.MustCompile(("revision/text"))
  titleXPath = xmlpath.MustCompile("title")

  // fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
  fi, err := os.Open("sample6.xml")
  //fi, err := os.Open("sample.xml")
  //fi, err := os.Open("sample2.xml")

  if err != nil {

    panic(err)
  }
  // close fi on exit and check for its returned error
  defer func() {
    if err := fi.Close(); err != nil {
      panic(err)
    }
  }()

  t := time.Now()
  prefix := t.Format(layout)
  fname := fmt.Sprintf("quotes-fr-%s.csv", prefix)

  // open output file
  glog.V(1).Infof("Creating file %s", fname)
  quoteCSV, err := os.Create(fname)
  if err != nil {
    panic(err)
  }
  // close fo on exit and check for its returned error
  defer func() {
    if err := quoteCSV.Close(); err != nil {
      panic(err)
    }
  }()

  quoteWriter = csv.NewWriter(quoteCSV)

  // make a read buffer
  r := bufio.NewReader(fi)

  root, err := xmlpath.Parse(r)
  if err != nil {
    glog.Fatal(err)
  }
  iter := pageXPath.Iter(root)

  go Multiplex()
  go Multiplex()
  go Multiplex()
  go Multiplex()

  for iter.Next() {
    page := iter.Node()
    pageChannel <- page

  }
}

func Multiplex() {
  for page := range pageChannel {
    extractAndTokenize(page)
  }
}

func extractAndTokenize(page *xmlpath.Node) {

  content, _ := textXPath.String(page)
  id, _ := pageIdXPath.String(page)
  title, _ := titleXPath.String(page)

  if strings.Index(title, "Modèle:") == -1 &&
    strings.Index(title, "Catégorie:") == -1 &&
    strings.Index(title, "MediaWiki:") == -1 &&
    strings.Index(title, "Aide:") == -1 {
    glog.V(1).Infof("Entering %s", title)
    tokens := Tokenize(content)
    i, _ := strconv.Atoi(id)
    ExtractQuoteNodes(Parse(tokens), title, i)
  }
}
