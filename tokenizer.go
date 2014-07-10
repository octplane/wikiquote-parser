package wikiquote_parser

import (
  "fmt"
  "strings"
  "unicode/utf8"
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
  if l.pos > l.start {
    l.emit(itemText)
  }
  l.emit(itemEOF)
  return nil // Stop the run loop.
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

func (its tokens) String() string {
  out := ""
  for _, it := range its {
    out += it.String() + "\n"
  }
  return out
}
