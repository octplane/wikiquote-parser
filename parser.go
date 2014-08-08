package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
)

type parser struct {
  name     string
  items    []item
  start    int
  pos      int
  consumed int
}

func create_parser(name string, tokens []item) *parser {
  p := &parser{
    name:     name,
    items:    tokens,
    start:    0,
    pos:      0,
    consumed: 0,
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

func (p *parser) eatUntil(types ...token) {
  it := p.eatCurrentItem()
  exp := make([]string, 0)

  for _, typ := range types {
    if it.Typ == token(typ) {
      return
    }
    exp = append(exp, token(typ).String())
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.Val, exp, it.Typ.String()))
}

func (p *parser) scanSubArgumentsUntil(node *Node, stop token) {
  glog.V(2).Infof("Sub-scanning until %s", stop.String())
  elt := p.currentItem()
  for elt.Typ != tokenEOF {
    elt = p.currentItem()
    glog.V(2).Infof("Element is now %s, scanning until %s\n", elt.String(), stop.String())
    switch elt.Typ {
    case stop, tokenEOF:
      glog.V(2).Infof("Finished sub-scanning at token %s", elt.Typ.String())
      return
    case itemPipe:
      p.consume(1)
    case itemText:
      p.consume(1)
      if p.currentItem().Typ == tokenEq {
        k := elt.Val
        p.consume(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Param for %s", p.name, k), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
        node.NamedParams[k] = params
        p.consume(consumed)
      } else {
        p.backup(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous parameter", p.name), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
        node.Params = append(node.Params, params)
        p.consume(consumed)
      }
    default:
      params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous Complex parameter", p.name), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
      node.Params = append(node.Params, params)
      p.consume(consumed)
    }
  }
}

// Start at next double LF
func (p *parser) nextBlock() {
  glog.V(2).Infof("Will now attempt to find next block for %s\n", p.items[p.pos:])
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

type envAlteration struct {
  exitTypes       []token
  exitSequence    []token
  forbiddenMarkup []markup
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

func ParseWithEnv(title string, items []item, env envAlteration) (ret Nodes, consumed int) {
  p := create_parser(title, items)
  glog.V(2).Infof("%s: Creating Parser (%s) with %d items\n", title, env.String(), len(items))
  ret = make([]Node, 0)

  ret = p.parse(env)
  glog.V(2).Infof("%s: Consumed %d / %d\n", title, p.consumed, len(p.items))
  return ret, p.consumed
}

func Parse(items []item) (ret Nodes) {
  nodes, _ := ParseWithConsumed(items)
  return nodes
}

func ParseWithConsumed(items []item) (ret Nodes, consumed int) {
  return ParseWithEnv("top-level", items, envAlteration{})
}

func (p *parser) parse(env envAlteration) (ret Nodes) {
  ret = make([]Node, 0)

  defer func() {
    if err := recover(); err != nil {
      ret = p.handleParseError(err, env, ret)
    }
  }()

  glog.V(2).Infof("Exit sequence: %s\n", env.String())
  p.pos = 0
  it := p.currentItem()

  for it.Typ != tokenEOF {
    glog.V(2).Infof("token %s\n", it.String())
    // If the exit Sequence match, abort immediately
    if len(env.exitSequence) > 0 {
      matching := 0
      for _, ty := range env.exitSequence {
        if p.currentItem().Typ == ty {
          matching += 1
          p.consume(1)
        } else {
          break
        }
      }
      if matching == len(env.exitSequence) {
        glog.V(2).Infof("Found exit sequence: %s\n", p.items[p.pos])
        return ret
      } else {
        p.backup(matching)
      }
    }

    it = p.currentItem()
    for _, typ := range env.exitTypes {
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
    case templateStart:
      n = p.ParseTemplate()
    case placeholderStart:
      n = p.ParsePlaceholder()
    case tokenEq:
      n = Node{Typ: NodeText, Val: "="}
      if p.pos == 0 {
        n = p.parseTitle()
      }
    case tokenLF:
      glog.V(2).Info("LF")

      p.consume(1)
      if p.currentItem().Typ == tokenEq {
        n = p.parseTitle()
      } else {
        p.backup(1)
        n = Node{Typ: NodeText, Val: "\n"}
      }
    default:
      glog.V(2).Infof("UNK", it.String())
      n = Node{Typ: NodeUnknown, Val: it.Val}
    }
    glog.V(2).Infof("Appending %s (remains %s), until", n.String(), p.items[p.pos:], env.String())
    ret = append(ret, n)

    p.consume(1)
    it = p.currentItem()

  }

  if (len(env.exitSequence) > 0 || len(env.exitTypes) > 0) && p.currentItem().Typ == tokenEOF {
    outOfBoundsPanic(p, len(p.items))
  }

  return ret
}
