package wikiquote_parser

import (
  "fmt"
)

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
  tokenEq
  controlStruct
  itemEOF
)

// https://en.wikipedia.org/wiki/Help:Cheatsheet
const leftTemplate = "{{"
const rightTemplate = "}}"
const leftLink = "[["
const rightLink = "]]"
const pipe = "|"
const eq = "="
const lf = "\n"

var strToToken map[string]itemType

func init() {
  strToToken = map[string]itemType{
    leftTemplate:  templateStart,
    rightTemplate: templateEnd,
    leftLink:      linkStart,
    rightLink:     linkEnd,
    pipe:          itemPipe,
    eq:            tokenEq,
  }
}

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
  case tokenEq:
    return " EQ "
  default:
    return fmt.Sprintf("%s %s", desc, i.val)
  }
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
  case tokenEq:
    return " = "
  default:
    return "??"
  }
}
