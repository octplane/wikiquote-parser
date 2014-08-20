package main

import (
  "bufio"
  "flag"
  "fmt"
  "github.com/golang/glog"
  . "github.com/octplane/wikiquote-parser"
  "launchpad.net/xmlpath"
  "os"
  "sort"
  "strings"
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

// type Book struct {
//   Title  string
//   Author string
//   Editor string
//   Year   string
//   Page   string
//   Isbn   string
// }

// type Quote struct {
//   Text   string
//   Author string
//   Book   Book
// }

// Thank you andrew https://groups.google.com/d/msg/golang-nuts/FT7cjmcL7gw/Gj4_aEsE_IsJ
// A data structure to hold a key/value pair.
type Pair struct {
  Key   string
  Value int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it.
func sortMapByValue(m map[string]int) PairList {
  p := make(PairList, len(m))
  i := 0
  for k, v := range m {
    p[i] = Pair{k, v}
    i += 1
  }
  sort.Sort(sort.Reverse(p))
  return p
}

func cleanup(in string) string {
  return strings.Replace(
    strings.Replace(
      strings.Replace(
        strings.Replace(
          strings.Replace(in, "[[", "", -1), "]]", "", -1),
        "<poem>", "", -1),
      "</poem>", "", -1),
    "\n", " ", -1)
}

type nodeType int

const (
  quote = nodeType(iota)
  source
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
  } else {
    return unknownType
  }
}

type Quote struct {
  source Node
  quote  Node
}

func (q *Quote) nonEmpty() bool {
  return q.source.Typ != NodeEmpty && q.quote.Typ != NodeEmpty
}

func (q *Quote) title() string {
  return q.source.StringParamOrEmpty("titre")
}

func (q *Quote) isbn() string {
  return q.source.StringParamOrEmpty("ISBN")
}

func (q *Quote) authorText() string {
  return q.source.StringParamOrEmpty("auteur")
}

var MAX_QUOTE_LENGTH = 100

func (q *Quote) valid() bool {
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

func (q *Quote) quoteString() string {
  if len(q.quote.Params) == 1 {
    return q.quote.Params[0].StringRepresentation()
  } else {
    return q.quote.StringParamOrEmpty("citation")
  }
  return ""
}

func (q *Quote) StringRepresentation(category string) string {

  var authortext string

  quotetext := q.quoteString()
  authortext = q.authorText()
  title := q.title()
  isbn := q.isbn()

  return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", category, isbn, authortext, title, quotetext)
}

func ExtractQuotes(nodes Nodes, theme string) {
  var q Quote = Quote{source: EmptyNode(), quote: EmptyNode()}
  count := 0

  for _, node := range nodes {
    switch normalizedType(node) {
    case quote:
      q.quote = node
    case source:
      q.source = node
    }
    if q.nonEmpty() {
      if q.valid() {
        count += 1
        fmt.Println(q.StringRepresentation(theme))
      } else {
        glog.V(2).Infoln("Ignoring quote")
      }
      q = Quote{source: EmptyNode(), quote: EmptyNode()}
    }
  }
}

func main() {
  flag.Parse()
  pageXPath := xmlpath.MustCompile("/mediawiki/page")
  textXPath := xmlpath.MustCompile(("revision/text"))
  titleXPath := xmlpath.MustCompile("title")

  //fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
  fi, err := os.Open("sample4.xml")
  //fi, err := os.Open("sample.xml")
  //fi, err := os.Open("sample1.xml")

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
    glog.Fatal(err)
  }
  iter := pageXPath.Iter(root)
  for iter.Next() {
    page := iter.Node()
    content, _ := textXPath.String(page)
    title, _ := titleXPath.String(page)

    tokens := Tokenize(content)
    glog.V(2).Infof("Entering %s", title)
    ExtractQuotes(Parse(tokens), title)
  }
}
