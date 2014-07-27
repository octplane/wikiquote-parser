package wikimediaparser

import (
  "fmt"
  "log"
  "os"
)

type parser struct {
  items           []item
  start           int
  pos             int
  logger          *log.Logger
  ignoreNextBlock bool
}

func create_parser(tokens []item) *parser {
  p := &parser{
    items:           tokens,
    start:           0,
    pos:             0,
    logger:          log.New(os.Stdout, "[Parse]\t", log.LstdFlags),
    ignoreNextBlock: false,
  }
  return p
}

func Parse(items []item) (ret Nodes) {
  p := create_parser(items)
  ret = make([]Node, 0)

  defer func() {
    if err := recover(); err != nil {
      ret = p.handleParseError(err, ret)
    }
  }()

  ret = p.Parse(envAlteration{})
  return ret
}

func (p *parser) currentItem() item {
  if p.pos > len(p.items) {
    outOfBoundsPanic(p, p.start)
  }

  return p.items[p.pos]
}

func (p *parser) eatCurrentItem() item {
  ret := p.currentItem()
  p.pos += 1
  return ret
}

// Start at next double LF
func (p *parser) nextBlock() {
  for p.pos < len(p.items) {
    if p.eatCurrentItem().typ == tokenLF {
      if p.eatCurrentItem().typ == tokenLF {
        p.backupI(2)
        return
      }
    }
  }
  // if we arrive here, we are at eof
  outOfBoundsPanic(p, p.start)
}

func (p *parser) next() {
  p.pos += 1
}

func (p *parser) eat(types ...token) {
  it := p.eatCurrentItem()
  exp := make([]string, 0)

  for _, typ := range types {
    if it.typ == token(typ) {
      return
    }
    exp = append(exp, token(typ).String())
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.val, exp, it.typ.String()))

}

func (p *parser) backup() {
  p.backupI(1)
}

func (p *parser) backupI(count int) {
  p.pos -= count
}

func (p *parser) ahead(count int) {
  p.pos += count
}

type envAlteration struct {
  exitTypes    []token
  exitSequence []token
}

func (ev *envAlteration) String() string {
  st := ""
  if len(ev.exitTypes) > 0 {
    st += "Closing types: "
  }
  for _, et := range ev.exitTypes {
    st += et.String() + " "
  }

  if len(ev.exitTypes) > 0 {
    st += "\n"
  }

  if len(ev.exitSequence) > 0 {
    st += "Closing Sequence: "
  }
  for _, es := range ev.exitSequence {
    st += es.String()
  }
  return st
}

func (p *parser) Parse(st envAlteration) Nodes {
  ret := make([]Node, 0)

  for p.pos < len(p.items)-1 {
    it := p.currentItem()

    // If the exit Sequence match, abort immediately
    if len(st.exitSequence) > 0 {
      matching := 0
      for _, ty := range st.exitSequence {
        if p.currentItem().typ == ty {
          matching += 1
          p.next()
        } else {
          break
        }
      }
      if matching == len(st.exitSequence) {
        p.backupI(len(st.exitSequence))
        return ret
      } else {
        p.backupI(matching)
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
    case tokenEOF:
      break
    case tokenEq:
      p.eatCurrentItem()
      if p.currentItem().typ == tokenEq {
        n = p.parseTitle()
      } else {
        p.backup()
        n = Node{typ: nodeText, val: "="}
      }
    }
    if n.typ != nodeInvalid {
      ret = append(ret, n)
    }

    it = p.currentItem()
    for _, typ := range st.exitTypes {
      if it.typ == token(typ) {
        return ret
      }
    }
    p.pos += 1
  }
  return ret
}

func (p *parser) ParseLink() Node {
  ret := Node{typ: nodeLink, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eat(linkStart)
  ret.namedParams["link"] = p.subparse(envAlteration{exitTypes: []token{itemPipe, linkEnd}})
  p.scanSubArgumentsUntil(&ret, linkEnd)

  return ret
}

func (p *parser) ParseTemplate() Node {
  ret := Node{typ: nodeTemplate, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eat(templateStart)
  ret.namedParams["name"] = p.subparse(envAlteration{exitTypes: []token{itemPipe, templateEnd}})
  p.scanSubArgumentsUntil(&ret, templateEnd)

  return ret
}

func (p *parser) parseTitle() Node {
  var ret Node
  exitSequence := make([]token, 0)

  item := p.currentItem()
  level := 0

  for item.typ == tokenEq {
    exitSequence = append(exitSequence, tokenEq)
    item = p.eatCurrentItem()
    level += 1
  }
  defer p.handleTitleError(p.pos, level)

  if level < 2 {
    ret = Node{typ: nodeEq}
    return ret
  }
  // put last seen token in front of here
  p.backup()

  ret = Node{typ: nodeTitle, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}
  ret.namedParams["level"] = Nodes{Node{typ: nodeText, val: fmt.Sprintf("%d", level)}}
  ret.namedParams["title"] = p.subparse(envAlteration{exitSequence: exitSequence})
  p.ahead(level)

  return ret
}

func (p *parser) scanSubArgumentsUntil(node *Node, stop token) {
  cont := true
  for cont {
    elt := p.currentItem()
    switch elt.typ {
    case stop:
      return
    case itemPipe:
      p.next()
    case itemText:
      p.next()
      if p.currentItem().typ == tokenEq {
        k := elt.val
        p.next()
        node.namedParams[k] = p.subparse(envAlteration{exitTypes: []token{itemPipe, stop}})
      } else {
        p.backup()
        node.params = append(node.params, p.subparse(envAlteration{exitTypes: []token{itemPipe, stop}}))
      }
    default:
      node.params = append(node.params, p.subparse(envAlteration{exitTypes: []token{itemPipe, stop}}))
    }
  }

}

func (p *parser) subparse(st envAlteration) Nodes {
  var res Nodes
  res = p.Parse(st)
  return res
}
