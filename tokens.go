package wikimediaparser

import (
  "fmt"
)

type item struct {
  Typ token
  Val string
}

type token int

const (
  itemError = token(iota)
  placeholderStart
  placeholderEnd
  templateName
  templateStart
  templateEnd
  linkStart
  linkEnd
  tokenELnkStart
  tokenELnkEnd
  variableName
  itemText
  itemPipe
  tokenEq
  tokenSp
  tokenLF
  controlStruct
  tokenNowikiStart
  tokenNowikiEnd
  tokenEOF
)

// https://en.wikipedia.org/wiki/Help:Cheatsheet
const leftTemplate = "{{"
const rightTemplate = "}}"
const leftPlaceholder = "{{{"
const rightPlaceholder = "}}}"
const leftExtLink = "["
const rightExtLink = "]"
const leftLink = "[["
const rightLink = "]]"
const pipe = "|"
const eq = "="
const sp = " "
const lf = "\n"
const nowikistart = "<nowiki>"
const nowikiend = "</nowiki>"

var strToToken map[string]token
var tokensAsString []string

func init() {
  strToToken = map[string]token{
    leftPlaceholder:  placeholderStart,
    leftTemplate:     templateStart,
    rightTemplate:    templateEnd,
    rightPlaceholder: placeholderEnd,
    leftLink:         linkStart,
    leftExtLink:      tokenELnkStart,
    rightExtLink:     tokenELnkEnd,
    rightLink:        linkEnd,
    pipe:             itemPipe,
    eq:               tokenEq,
    sp:               tokenSp,
    lf:               tokenLF,
    nowikistart:      tokenNowikiStart,
    nowikiend:        tokenNowikiEnd,
  }
  tokensAsString = []string{
    leftPlaceholder,
    leftTemplate,
    rightTemplate,
    rightPlaceholder,
    leftLink,
    leftExtLink,
    rightLink,
    rightExtLink,
    pipe,
    eq,
    sp,
    lf,
    nowikistart,
    nowikiend,
  }
}

func (i item) String() string {
  desc := i.Typ.String()
  switch i.Typ {
  case tokenEOF:
    return "E"
  case itemError:
    return i.Val
  case templateStart, templateEnd, linkStart, linkEnd, placeholderStart, placeholderEnd,
    tokenNowikiStart, tokenNowikiEnd, tokenELnkStart, tokenELnkEnd, tokenSp, tokenEq:
    return fmt.Sprintf("%s", desc)
  case itemText:
    if len(i.Val) > 40 {
      return fmt.Sprintf("%s[...]%s", i.Val[:17], i.Val[len(i.Val)-17:])
    }
    return i.Val
  case tokenLF:
    return "\\n"
  case itemPipe:
    return fmt.Sprintf("%s", desc)
  case controlStruct:
    return fmt.Sprintf("%s %s", desc, i.Val)
  default:
    return fmt.Sprintf("%s %s", desc, i.Val)
  }
}

func (itt token) String() string {
  switch itt {
  case tokenEOF:
    return "<EOF>"
  case itemError:
    return "Error"
  case templateStart:
    return "{{"
  case templateEnd:
    return "}}"
  case linkStart:
    return leftLink
  case linkEnd:
    return rightLink
  case placeholderStart:
    return "{{{"
  case placeholderEnd:
    return "}}}"
  case tokenELnkStart:
    return "["
  case tokenELnkEnd:
    return "]"
  case itemText:
    return "Text"
  case itemPipe:
    return "|"
  case controlStruct:
    return "Control struct"
  case tokenEq:
    return "="
  case tokenSp:
    return " "
  case tokenNowikiStart:
    return "<noWiki>"
  case tokenNowikiEnd:
    return "</noWiki>"
  default:
    return "??"
  }
}
