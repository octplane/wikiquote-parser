package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
)

type parser struct {
  name         string
  parent       *parser
  items        []item
  start        int
  pos          int
  consumed     int
  exitTypes    []token
  exitSequence []token
  onError      behaviorOnError
}

func create_parser(name string, tokens []item, exitTypes []token, exitSequence []token, onError behaviorOnError) *parser {
  p := &parser{
    name:         name,
    items:        tokens,
    start:        0,
    pos:          0,
    consumed:     0,
    exitTypes:    exitTypes,
    exitSequence: exitSequence,
    onError:      onError,
  }

  return p
}

func (p *parser) currentItem() item {
  if p.pos > len(p.items)-1 {
    return item{Typ: tokenEOF}
  }

  return p.items[p.pos]
}

func (p *parser) consume(count int) {
  p.pos += count
  p.consumed += count
}

func (p *parser) backup(count int) {
  p.consume(-1 * count)
}

func (p *parser) eatCurrentItem() item {
  ret := p.currentItem()
  p.pos += 1
  p.consumed += 1
  return ret
}

func (p *parser) eat(typ token) {
  it := p.eatCurrentItem()
  exp := make([]string, 0)

  if it.Typ == token(typ) {
    return
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.Val, exp, it.Typ.String()))
}

func (p *parser) eatUntil(typ token) {
  it := p.currentItem()

  for it.Typ != token(typ) {
    it = p.eatCurrentItem()
  }
}

func (p *parser) log(format string, params ...interface{}) {
  parms := []interface{}{p.name}
  parms = append(parms, params...)
  glog.V(2).Infof("[%s] "+format, parms...)
}

func Cleanup(source string) string {
  return Parse(Tokenize(source)).StringRepresentation()
}

func Parse(items []item) (ret Nodes) {
  nodes, _ := ParseWithConsumed(items)
  return nodes
}

func (p *parser) CreateParser(name string, tokens []item, exitTypes []token, exitSequence []token, onError behaviorOnError) *parser {
  ret := create_parser(name, tokens, exitTypes, exitSequence, onError)
  ret.parent = p
  return ret
}

func ScanSubArgumentsUntil(name string, parent *parser, node *Node, stop token) (consumed int) {
  var p *parser
  name = name + "-subArgs"
  if parent != nil {
    p = parent.CreateParser(name, parent.items[parent.pos:], nil, nil, ignoreSectionBehavior)
  } else {
    p = create_parser(name, parent.items[parent.pos:], nil, nil, ignoreSectionBehavior)
  }
  p.log("%s: Creating Argument Parser (%s)\n", name, p.EnvironmentString())
  p.scanSubArgumentsUntil(node, stop)
  return p.consumed
}

func (p *parser) scanSubArgumentsUntil(node *Node, stop token) {
  p.log("Sub-scanning until %s", stop.String())
  ret := make([]Node, 0)

  defer func() {
    if err := recover(); err != nil {
      ret = p.handleParseError(err, ret)
      node.Params = append(node.Params, ret)
    }
  }()

  elt := p.currentItem()
  for elt.Typ != tokenEOF {
    elt = p.currentItem()
    p.log("Element is now %s, scanning until %s\n", elt.String(), stop.String())
    switch elt.Typ {
    case stop, tokenEOF:
      p.consume(1)
      p.log("Finished sub-scanning at token %s, remains: %s", elt.Typ.String(), p.items[p.pos:])
      return
    case itemPipe:
      p.consume(1)
    case itemText:
      p.consume(1)
      if p.currentItem().Typ == tokenEq {
        k := elt.Val
        p.consume(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Param for %s", p.name, k), p, p.items[p.pos:], []token{itemPipe, stop}, nil, abortBehavior)
        node.NamedParams[k] = params
        p.consume(consumed)
      } else {
        p.backup(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous parameter", p.name), p, p.items[p.pos:], []token{itemPipe, stop}, nil, abortBehavior)
        node.Params = append(node.Params, params)
        p.consume(consumed)
      }
    default:
      params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous Complex parameter", p.name), p, p.items[p.pos:], []token{itemPipe, stop}, nil, abortBehavior)
      node.Params = append(node.Params, params)
      p.consume(consumed)
    }
  }
}

func (p *parser) nextLine() {
  p.log("Will now attempt to find next line for %s\n", p.items[p.pos:])
  for p.pos < len(p.items) {
    if p.currentItem().Typ == tokenLF {
      p.consume(1)
      return
    }
    p.consume(1)
  }
  if p.items[len(p.items)-1].Typ == tokenEOF {
    p.pos = len(p.items) - 1
  } else {
    p.pos = len(p.items)
  }
}

// Start at next double LF
func (p *parser) nextBlock() {
  p.log("Parser: %s Will now attempt to find next block in %s\n", p.name, p.items[p.pos:])
  for p.pos < len(p.items) {
    if p.currentItem().Typ == tokenLF {
      p.consume(1)
      if p.currentItem().Typ == tokenLF {
        p.consume(1)
        return
      }
    }
    p.consume(1)
  }
  if p.items[len(p.items)-1].Typ == tokenEOF {
    p.pos = len(p.items) - 1
  } else {
    p.pos = len(p.items)
  }
}

func (ev *parser) EnvironmentString() string {
  st := ""
  if len(ev.exitTypes) > 0 {
    st += "Closing types: "
  }
  for _, et := range ev.exitTypes {
    st += et.String() + " "
  }

  if len(ev.exitSequence) > 0 {
    st += ". Closing Sequence: "
  }
  for _, es := range ev.exitSequence {
    st += es.String()
  }

  st += ". Error does: " + ev.onError.String()
  return st
}

func ParseWithEnv(name string, parent *parser, items []item, exitTypes []token, exitSequence []token, onError behaviorOnError) (ret Nodes, consumed int) {
  var p *parser
  if parent != nil {
    p = parent.CreateParser(name, items, exitTypes, exitSequence, onError)
  } else {
    p = create_parser(name, items, exitTypes, exitSequence, onError)
  }
  p.log("%s: Creating Parser (%s) with %d items: %+v\n", name, p.EnvironmentString(), len(items), items)
  ret = make([]Node, 0)

  ret = p.parse()
  p.log("%s: Consumed %d / %d\n", name, p.consumed, len(p.items))
  return ret, p.consumed
}

func ParseWithConsumed(items []item) (ret Nodes, consumed int) {
  return ParseWithEnv("top-level", nil, items, nil, nil, ignoreSectionBehavior)
}

func (p *parser) parse() (ret Nodes) {
  ret = make([]Node, 0)

  defer func() {
    if err := recover(); err != nil {
      ret = p.handleParseError(err, ret)
    }
  }()

  p.log("Starting parsing with exit sequence: %s\n", p.EnvironmentString())
  p.pos = 0
  it := p.currentItem()

  for it.Typ != tokenEOF {
    p.log("token %s\n", it.String())
    // If the exit Sequence match, abort immediately
    if len(p.exitSequence) > 0 {
      matching := 0
      for _, ty := range p.exitSequence {
        if p.currentItem().Typ == ty {
          matching += 1
          p.consume(1)
        } else {
          break
        }
      }
      if matching == len(p.exitSequence) {
        p.log("Found exit sequence: %s\n", p.items[p.pos])
        return ret
      } else {
        p.backup(matching)
      }
    }

    it = p.currentItem()
    for _, typ := range p.exitTypes {
      if it.Typ == token(typ) {
        return ret
      }
    }

    var n Node = Node{Typ: NodeInvalid}
    switch it.Typ {
    case itemText:
      n = Node{Typ: NodeText, Val: it.Val}
    case linkStart:
      n = p.ParseLink()
    case tokenELnkStart:
      n = p.ParseELink()
    case templateStart:
      n = p.ParseTemplate()
    case placeholderStart:
      n = p.ParsePlaceholder()
    case tokenEq:
      n = Node{Typ: NodeText, Val: "="}
      if p.pos == 0 {
        n = p.parseTitle()
      }
    case tokenSp:
      n = Node{Typ: NodeText, Val: " "}
    case tokenLF:
      p.consume(1)
      if p.currentItem().Typ == tokenEq {
        n = p.parseTitle()
      } else {
        p.backup(1)
        n = Node{Typ: NodeText, Val: "\n"}
      }
    case tokenNowikiStart:
      n = p.ParseNowiki()
    default:
      n = Node{Typ: NodeUnknown, Val: it.Val}
    }
    p.log("[%s] Appending %s (remains %s), until %s", p.name, n.String(), p.items[p.pos:], p.EnvironmentString())
    ret = append(ret, n)

    if p.currentItem().Typ != tokenEOF {
      p.consume(1)
    }
    it = p.currentItem()

  }

  if (len(p.exitSequence) > 0 || len(p.exitTypes) > 0) && p.currentItem().Typ == tokenEOF {
    outOfBoundsPanic(p, len(p.items))
  }

  return ret
}
