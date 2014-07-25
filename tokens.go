package wikimediaparser

import (
  "fmt"
)

type item struct {
  typ token
  val string
}

type token int

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
  tokenEOF
)

// https://en.wikipedia.org/wiki/Help:Cheatsheet
const leftTemplate = "{{"
const rightTemplate = "}}"
const leftLink = "[["
const rightLink = "]]"
const pipe = "|"
const eq = "="
const lf = "\n"

var strToToken map[string]token

func init() {
  strToToken = map[string]token{
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
  case tokenEOF:
    return "E"
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
    if len(i.val) > 40 {
      return fmt.Sprintf("%s[...]%s", i.val[:17], i.val[len(i.val)-17:])
    }
    return i.val
  case itemPipe:
    return fmt.Sprintf("%s", desc)
  case controlStruct:
    return fmt.Sprintf("%s %s", desc, i.val)
  case tokenEq:
    return "="
  default:
    return fmt.Sprintf("%s %s", desc, i.val)
  }
}

func (itt token) String() string {
  switch itt {
  case tokenEOF:
    return "E"
  case itemError:
    return "Error"
  case templateStart:
    return "Template start"
  case templateEnd:
    return "Template end"
  case linkStart:
    return "Link start"
  case linkEnd:
    return rightLink
  case itemText:
    return "Text"
  case itemPipe:
    return "|"
  case controlStruct:
    return "Control struct"
  case tokenEq:
    return "="
  default:
    return "??"
  }
}
