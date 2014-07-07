package wikiquote_parser

import (
  "fmt"
  "regexp"
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

/// LEXER

type item struct {
  typ itemType
  val string
}

type itemType int

const (
  itemError = iota
  templateName
  templateEnd
  variableName
  itemText
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
  switch i.typ {
  case itemEOF:
    return "EOF"
  case itemError:
    return i.val
  case templateEnd:
    return fmt.Sprintf("IT: Template end: %s>>\n", i.val)
  case templateName:
    return fmt.Sprintf("IT: Template name: %s\n", i.val)
  case variableName:
    return fmt.Sprintf("IT: Variable: \tvar:%s \n", i.val)
  case itemText:
    if len(i.val) > 10 {
      return fmt.Sprintf("IT: Text: %.10q...\n", i.val)
    }
    return fmt.Sprintf("IT: Text: %q\n", i.val)
  case controlStruct:
    return "IT: control\n"
  }
  return fmt.Sprintf("IT: Unknown item %+v\n", i.val)
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
  l.items <- item{t, l.input[l.start:l.pos]}
  l.start = l.pos
}

func (l *lexer) emitItem(t item) {
  l.items <- t
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

// func lexNumber(l *lexer) stateFn {
//   // Optional leading sign.
//   l.accept("+-")
//   // Is it hex?
//   digits := "0123456789"
//   if l.accept("0") && l.accept("xX") {
//     digits = "0123456789abcdefABCDEF"
//   }
//   l.acceptRun(digits)
//   if l.accept(".") {
//     l.acceptRun(digits)
//   }
//   if l.accept("eE") {
//     l.accept("+-")
//     l.acceptRun("0123456789")
//   }
//   // Is it imaginary?
//   l.accept("i")
//   // Next thing mustn't be alphanumeric.
//   if isAlphaNumeric(l.peek()) {
//     l.next()
//     return l.errorf("bad number syntax: %q",
//       l.input[l.start:l.pos])
//   }
//   l.emit(itemNumber)
//   return lexInsideAction
// }

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
    default:
      l.state = l.state(l)
    }
  }
  panic("not reached")
}

func lexText(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], leftTemplate) {
      if l.pos > l.start {
        l.emit(itemText)
      }
      l.accept("{")
      l.accept("{")
      l.emit(controlStruct)
      return lexTemplateStart // Next state.
    }
    if l.next().typ == eof {
      break
    }
  }
  // Correctly reached EOF.
  if l.pos > l.start {
    l.emit(itemText)
  }
  l.emit(itemEOF) // Useful to make EOF a token.
  return nil      // Stop the run loop.
}

const leftTemplate = "{{"
const rightTemplate = "}}"

func lexTemplateStart(l *lexer) stateFn {
  // eat the template start
  l.refuse("|}")
  l.emit(templateName)
  return inTemplate
}

func inTemplate(l *lexer) stateFn {
  l.refuse("|}")
  if strings.HasPrefix(l.input[l.pos:], rightTemplate) {
    if l.pos > l.start {
      l.emit(templateEnd)
      l.start += 2
    }
    return lexText // Next state.
  } else if strings.HasPrefix(l.input[l.pos:], "|") {
    if l.start < l.pos {
      newLexer := lex("Variable Lexer", l.input[l.start:l.pos])
      halt := false
      for !halt {
        it := newLexer.nextItem()
        switch it.typ {
        case itemEOF:
          halt = true
        }

        if halt != true {
          l.emitItem(it)
        }
      }
      l.start = l.pos
    }
    l.accept("|")
    // l.accept("=")
    // fmt.Println("Starting sub parser for", l.input[l.start:l.pos])
    // newLexer := lex("Variable Lexer", l.input[l.start:l.pos])
    // halt := false
    // for !halt {
    // 	it := newLexer.nextItem()
    // 	switch it.typ {
    // 	case itemEOF:
    // 		halt = true
    // 	}

    // 	if halt != true {
    // 		l.emitItem(it)
    // 	}
    // }
    return inTemplate
  }
  return nil
}

// const leftMeta = "{{"

// func lexText(l *lexer) stateFn {
//   for {
//     if strings.HasPrefix(l.input[l.pos:], leftMeta) {
//       if l.pos > l.start {
//         l.emit(itemText)
//       }
//       return lexLeftMeta // Next state.
//     }
//     if l.next().typ == eof {
//       break
//     }
//   }
//   // Correctly reached EOF.
//   if l.pos > l.start {
//     l.emit(itemText)
//   }
//   l.emit(itemEOF) // Useful to make EOF a token.
//   return nil      // Stop the run loop.
// }

// func lexLeftMeta(l *lexer) stateFn {
//   l.pos += len(leftMeta)
//   l.emit(itemLeftMeta)
//   return lexInsideAction // Now inside {{ }}.
// }

// func lexInsideAction(l *lexer) stateFn {
//   // Either number, quoted string, or identifier.
//   // Spaces separate and are ignored.
//   // Pipe symbols separate and are emitted.
//   for {
//     if strings.HasPrefix(l.input[l.pos:], rightMeta) {
//       return lexRightMeta
//     }
//     switch r := l.next(); {
//     case r == eof || r == '\n':
//       return l.errorf("unclosed action")
//     case isSpace(r):
//       l.ignore()
//     case r == '|':
//       l.emit(itemPipe)
//     case r == '"':
//       return lexQuote
//     case r == '`':
//       return lexRawQuote
//     case r == '+' || r == '-' || '0' <= r && r <= '9':
//       l.backup()
//       return lexNumber
//     case isAlphaNumeric(r):
//       l.backup()
//       return lexIdentifier
//     }
//   }
// }
// func lexNumber(l *lexer) stateFn {
//     // Optional leading sign.
//     l.accept("+-")
//     // Is it hex?
//     digits := "0123456789"
//     if l.accept("0") && l.accept("xX") {
//         digits = "0123456789abcdefABCDEF"
//     }
//     l.acceptRun(digits)
//     if l.accept(".") {
//         l.acceptRun(digits)
//     }
//     if l.accept("eE") {
//         l.accept("+-")
//         l.acceptRun("0123456789")
//     }

////

func Parse(body string) []item {
  ret := make([]item, 1000)
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

func markupExtractor(title string, body string) []Command {

  l := lex(title, body)
  go l.run()

  var it item
  halt := false
  for !halt {
    it = l.nextItem()
    fmt.Printf(it.String())
    switch it.typ {
    case itemEOF:
      halt = true
      fmt.Println("EOF !!")
      break
    }
  }
  panic("Hammer time!")

  markup := regexp.MustCompile("(?s){{[^}]+}}")
  param := regexp.MustCompile("([^=]+)=(.*)")

  strCommands := markup.FindAllString(body, -1)

  commands := make([]Command, len(strCommands))
  pos := 0

  for _, cmd := range strCommands {
    cmd = cmd[2 : len(cmd)-2]

    arguments := make([]string, 10000)
    argumentsIndex := 0
    kvArguments := make(map[string]string, 1000)

    for _, arg := range strings.Split(cmd, "|") {
      kv := param.FindStringSubmatch(arg)
      if len(kv) == 3 {
        key := strings.TrimSpace(strings.ToLower(kv[1]))
        val, exists := kvArguments[key]
        if exists && val != kv[0] {
          // FIXME: handle the issue
          // panic(fmt.Sprintf("Parameter %s already exists with value \"%s\", here wants : \"%s\"", key, val, kv[0]))
        } else {
          kvArguments[key] = kv[2]
        }
      } else {
        arguments[argumentsIndex] = arg
        argumentsIndex += 1
      }
    }

    cmd := strings.TrimSpace(strings.ToLower(arguments[0]))

    // Parse special "defaultsort:", "if:", "msg:" commands
    if strings.Index(cmd, ":") != -1 {
      cmd = cmd[0:strings.Index(cmd, ":")]
      // FIXME inject arguments in Command anyway
    }

    // Ignore the empty command
    if cmd != "" {
      commands[pos] = Command{Cmd: cmd, PageTitle: title, Arguments: arguments[1:], NamedArguments: kvArguments}
      fmt.Printf("%+v\n", commands[pos])
      pos += 1
    }
  }
  return commands
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
