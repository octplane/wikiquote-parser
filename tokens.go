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
  variableName
  itemText
  itemPipe
  tokenEq
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
const leftLink = "[["
const rightLink = "]]"
const pipe = "|"
const eq = "="
const lf = "\n"
const nowikistart = "<nowiki>"
const nowikiend = "</nowiki>"

var strToToken map[string]token
var tokensAsString []string

func init() {
  strToToken = map[string]token{
    leftPlaceholder:  placeholderStart,
    leftTemplate:     templateStart,
    rightPlaceholder: placeholderEnd,
    rightTemplate:    templateEnd,
    leftLink:         linkStart,
    rightLink:        linkEnd,
    pipe:             itemPipe,
    eq:               tokenEq,
    lf:               tokenLF,
    nowikistart:      tokenNowikiStart,
    nowikiend:        tokenNowikiEnd,
  }
  tokensAsString = []string{
    leftPlaceholder,
    leftTemplate,
    rightPlaceholder,
    rightTemplate,
    leftLink,
    rightLink,
    pipe,
    eq,
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
    tokenNowikiStart, tokenNowikiEnd:
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
  case tokenEq:
    return "="
  default:
    return fmt.Sprintf("%s %s", desc, i.Val)
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
  case placeholderStart:
    return "Placeholder start"
  case placeholderEnd:
    return "Placeholder end"
  case itemText:
    return "Text"
  case itemPipe:
    return "|"
  case controlStruct:
    return "Control struct"
  case tokenEq:
    return "="
  case tokenNowikiStart:
    return "<noWiki>"
  case tokenNowikiEnd:
    return "</noWiki>"
  default:
    return "??"
  }
}
