package wikiquote_parser

import (
  "fmt"
  "sort"
  "strings"
  "unicode/utf8"
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

type Book struct {
  Title  string
  Author string
  Editor string
  Year   string
  Page   string
  Isbn   string
}

type Quote struct {
  Text   string
  Author string
  Book   Book
}

type BookQuote struct {
  Book  Book
  Quote Quote
}

func (bq *BookQuote) ToString() string {
  var out string
  out = fmt.Sprintf("Quote: \"%s\", by %s, from %s, page, %d published in %d", bq.Quote.Text, bq.Quote.Author, bq.Book.Title, bq.Book.Page, bq.Book.Year)
  return out
}

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

//// Parser

/// LEXER

type item struct {
  typ itemType
  val string
}

type itemType int

const (
  itemError = iota
  templateName
  templateStart
  templateEnd
  linkStart
  linkEnd
  variableName
  itemText
  itemPipe
  controlStruct
  itemEOF
)

type sRune struct {
  typ runeType
  val rune
}

type runeType int

const (
  eof = iota
  ruune
)

func (i item) String() string {
  desc := i.typ.String()
  switch i.typ {
  case itemEOF:
    return "EOF"
  case itemError:
    return i.val
  case templateStart:
    return fmt.Sprintf("%s", desc)
  case templateEnd:
    return fmt.Sprintf("%s", desc)
  case linkStart:
    return fmt.Sprintf("%s", desc)
  case linkEnd:
    return fmt.Sprintf("%s", desc)
  case itemText:
    if len(i.val) > 10 {
      return fmt.Sprintf("%s: \"%.10s...\"", desc, i.val)
    }
    return fmt.Sprintf("%s: %q", desc, i.val)
  case itemPipe:
    return fmt.Sprintf("%s", desc)
  case controlStruct:
    return fmt.Sprintf("%s %s", desc, i.val)
  }
  return fmt.Sprintf("Unknown item %+v", desc, i.val)
}

func (itt itemType) String() string {
  switch itt {
  case itemEOF:
    return "EOF"
  case itemError:
    return "Error"
  case templateStart:
    return "Template start"
  case templateEnd:
    return "Template end"
  case linkStart:
    return "Link start"
  case linkEnd:
    return "Link end"
  case itemText:
    return "Text"
  case itemPipe:
    return "Pipe"
  case controlStruct:
    return "Control struct"
  }
  return "??"
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
  name  string    // used only for error reports.
  input string    // the string being scanned.
  start int       // start position of this item.
  pos   int       // current position in the input.
  width int       // width of last rune read from input.
  items chan item // channel of scanned items.
  state stateFn
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
  var it item

  if l.start > len(l.input) {
    it = item{t, ""}
  } else {
    it = item{t, l.input[l.start:l.pos]}
  }

  l.items <- it
  l.start = l.pos
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
  for state := lexText; state != nil; {
    state = state(l)
  }
  close(l.items) // No more tokens will be delivered.
}

// next returns the next rune in the input.
func (l *lexer) next() (r sRune) {
  var ru rune
  if l.pos >= len(l.input) {
    l.width = 0
    return sRune{typ: eof}
  }
  ru, l.width =
    utf8.DecodeRuneInString(l.input[l.pos:])
  l.pos += l.width
  return sRune{val: ru, typ: ruune}
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
  l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
  l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() sRune {
  r := l.next()
  l.backup()
  return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
  if strings.IndexRune(valid, l.next().val) >= 0 {
    return true
  }
  l.backup()
  return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
  for strings.IndexRune(valid, l.next().val) >= 0 {
  }
  l.backup()
}

func (l *lexer) refuse(valid string) {
  for strings.IndexRune(valid, l.next().val) == -1 {
  }
  l.backup()
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
  l.items <- item{
    itemError,
    fmt.Sprintf(format, args...),
  }
  return nil
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
  l := &lexer{
    name:  name,
    input: input,
    state: lexText,
    items: make(chan item, 20), // Max number of stacked items in the parse chain
  }
  return l
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
  for {
    select {
    case item := <-l.items:
      return item
    }
  }
  panic("not reached")
}

const leftTemplate = "{{"
const rightTemplate = "}}"
const leftLink = "[["
const rightLink = "]]"
const pipe = "|"

var strToToken map[string]itemType

func init() {
  strToToken = map[string]itemType{
    leftTemplate:  templateStart,
    rightTemplate: templateEnd,
    leftLink:      linkStart,
    rightLink:     linkEnd,
    pipe:          itemPipe,
  }
}
func lexText(l *lexer) stateFn {
  for {
    for st, it := range strToToken {
      if strings.HasPrefix(l.input[l.pos:], st) {
        if l.pos > l.start {
          l.emit(itemText)
          return lexText
        }
        l.pos = l.start + len(st)
        l.emit(it)
        return lexText
      }
    }
    if l.next().typ == eof {
      break
    }
    return lexText
  }
  l.emit(itemEOF) // Useful to make EOF a token.
  return nil      // Stop the run loop.
}

func Tokenize(body string) tokens {
  ret := make([]item, 0)
  l := lex("", body)
  go l.run()

  var it item
  halt := false
  for !halt {
    it = l.nextItem()
    ret = append(ret, it)
    switch it.typ {
    case itemEOF:
      halt = true
      break
    }
  }
  return ret
}

type tokens []item

func (its tokens) inspect() {
  fmt.Println("Tokens:")
  for _, it := range its {
    fmt.Println(it.String())
  }
  fmt.Println("...")
}

type parser struct {
  items []item
  start int
  pos   int
}

func create_parser(items []item) *parser {
  p := &parser{
    items: items,
    start: 0,
    pos:   0,
  }
  return p
}

func Parse(items []item) Nodes {
  p := create_parser(items)
  return p.Parse(envAlteration{})
}

func (p *parser) CurrentItem() item {
  return p.items[p.pos]
}

func (p *parser) nextItem() item {
  ret := p.items[p.pos]
  p.pos += 1
  return ret
}

func (p *parser) next() {
  p.pos += 1
}

func (p *parser) nextItemOfTypesOrSyntaxError(types ...itemType) item {
  it := p.nextItem()
  exp := make([]string, 0)

  for _, typ := range types {
    if it.typ == itemType(typ) {
      return it
    }
    exp = append(exp, itemType(typ).String())
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.val, exp, it.typ.String()))
}

func (p *parser) eat(types ...itemType) {
  it := p.nextItem()
  exp := make([]string, 0)

  for _, typ := range types {
    if it.typ == itemType(typ) {
      return
    }
    exp = append(exp, itemType(typ).String())
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.val, exp, it.typ.String()))

}

func (p *parser) backup() {
  p.pos -= 1
}

type envAlteration struct {
  firstToken *string
  exitTypes  []itemType
}

func (e *envAlteration) empty() bool {
  return e.firstToken == nil && len(e.exitTypes) == 0
}

func (p *parser) Parse(st envAlteration) Nodes {
  overrideText := false

  /* envAlteration can replace the current text node by another and can also specify exit constraint of the parser */
  if st.firstToken != nil {
    overrideText = true
  }

  ret := make([]Node, 0)

  for p.pos < len(p.items) {
    it := p.CurrentItem()
    if overrideText {
      // fmt.Printf("Overriding first item from %s to %s\n", it.val, *st.firstToken)
      if it.typ == itemText {
        it.val = *st.firstToken
        overrideText = false
      } else {
        panic(fmt.Sprintf("Have an override for the first token %q when item is not an itemText %+v ", *st.firstToken, it))
      }
    }

    var n Node = Node{typ: nodeInvalid}
    switch it.typ {
    case itemText:
      n = Node{typ: nodeText, val: it.val}
    case linkStart:
      n = p.ParseLink()
    case templateStart:
      n = p.ParseTemplate()
    case itemEOF:
      break
    }
    if n.typ != nodeInvalid {
      ret = append(ret, n)
    }
    it = p.CurrentItem()
    for _, typ := range st.exitTypes {
      if it.typ == itemType(typ) {
        return ret
      }
    }
    p.pos += 1
  }
  return ret
}

func (p *parser) inspect(ahead int) {
  fmt.Printf("------ Peeking ahead for %d elements\n", ahead)
  for pos, content := range p.items[p.pos:] {
    if pos > ahead {
      break
    }
    fmt.Printf(content.String())
  }
  fmt.Println("------ End of Peek")
}

func (p *parser) ParseLink() Node {
  ret := Node{typ: nodeLink, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eat(linkStart)
  ret.namedParams["link"] = p.subparse(envAlteration{exitTypes: []itemType{itemPipe, linkEnd}})
  p.scanSubArgumentsUntil(&ret, linkEnd)

  return ret
}

func (p *parser) ParseTemplate() Node {
  ret := Node{typ: nodeTemplate, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eat(templateStart)
  ret.namedParams["name"] = p.subparse(envAlteration{exitTypes: []itemType{itemPipe, templateEnd}})
  p.scanSubArgumentsUntil(&ret, templateEnd)

  return ret
}

func (p *parser) scanSubArgumentsUntil(node *Node, stop itemType) {
  cont := true
  for cont {
    elt := p.CurrentItem()
    switch elt.typ {
    case stop:
      return
    case itemPipe:
      p.next()
    case itemText:
      ix := strings.Index(elt.val, "=")
      if ix != -1 {
        k := elt.val[:ix]
        v := elt.val[ix+1:]
        node.namedParams[k] = p.subparse(envAlteration{firstToken: &v, exitTypes: []itemType{itemPipe, stop}})
      } else {
        node.params = append(node.params, p.subparse(envAlteration{exitTypes: []itemType{itemPipe, stop}}))
      }
    default:
      node.params = append(node.params, p.subparse(envAlteration{exitTypes: []itemType{itemPipe, stop}}))
    }
  }

}

func (p *parser) subparse(st envAlteration) []Node {
  return p.Parse(st)
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

func BuildQuote(qCommand Command, reference Command) {
  quote, hasQuote := qCommand.NamedArguments["citation"]
  if !hasQuote {
    quote = qCommand.Arguments[0]
    hasQuote = true
  }

  author := reference.ArgOrEmpty("auteur")
  title := reference.ArgOrEmpty("titre")
  // page := reference.ArgOrEmpty("page")
  // editor := reference.ArgOrEmpty("éditeur")
  // year := reference.ArgOrEmpty("année")
  isbn := reference.ArgOrEmpty("isbn")

  if quote == "" || author == "" || title == "" || isbn == "" {
  } else {
    if len(quote) < 130 {
      fmt.Printf("%s\t%s\t%s\t%s\t%s\n", qCommand.PageTitle, cleanup(quote), cleanup(author), title, isbn)
    }
  }

}

func ExtractQuotes(commands []Command) {
  var buffer *Command = nil
  for ix, cmd := range commands {
    if len(cmd.Cmd) > 3 && cmd.Cmd[0:4] == "réf" {
      if buffer != nil {
        BuildQuote(*buffer, cmd)
      } else {
        // fmt.Println("Found ref without quote")
        // fmt.Println(cmd)
      }
      buffer = nil
    } else if strings.Index(cmd.Cmd, "citation") == 0 {
      buffer = &commands[ix]
    } else {
      // fmt.Println("Unknown command:", cmd.Cmd)
      buffer = nil
    }
  }
}

func ExtractStats(commands []Command) {
  commandPopularity := make(map[string]int, 1000)
  argsPopularity := make(map[string]map[string]int, 100)

  var count int
  var hasCommand bool
  var aCount int
  var hasArg bool

  for _, cmd := range commands {
    count, hasCommand = commandPopularity[cmd.Cmd]
    if hasCommand {
      commandPopularity[cmd.Cmd] = count + 1
    } else {
      commandPopularity[cmd.Cmd] = 1
      argsPopularity[cmd.Cmd] = make(map[string]int, 100)
    }

    for k := range cmd.NamedArguments {
      aCount, hasArg = argsPopularity[cmd.Cmd][k]
      if hasArg {
        argsPopularity[cmd.Cmd][k] = aCount + 1
      } else {
        argsPopularity[cmd.Cmd][k] = 1
      }
    }
  }

  var maxArgsOcc int
  for _, pl := range sortMapByValue(commandPopularity) {
    fmt.Printf("### %s (%d occ.)\n", pl.Key, pl.Value)
    total := pl.Value
    maxArgsOcc = 0
    for _, ar := range sortMapByValue(argsPopularity[pl.Key]) {
      if ar.Value > maxArgsOcc {
        maxArgsOcc = ar.Value
      }
      if ar.Value > maxArgsOcc/10 && 100*ar.Value/total > 1 {
        fmt.Printf("- %s (%d%%)\n", ar.Key, 100*ar.Value/total)
      }
    }
  }
}
