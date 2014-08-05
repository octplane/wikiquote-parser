package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
)

type parser struct {
  name            string
  items           []item
  start           int
  pos             int
  consumed        int
  ignoreNextBlock bool
}

func create_parser(name string, tokens []item) *parser {
  p := &parser{
    name:            name,
    items:           tokens,
    start:           0,
    pos:             0,
    consumed:        0,
    ignoreNextBlock: false,
  }

  return p
}

func (p *parser) currentItem() item {
  if p.pos > len(p.items)-1 {
    return item{typ: tokenEOF}
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
    if it.typ == token(typ) {
      return
    }
    exp = append(exp, token(typ).String())
  }
  panic(fmt.Sprintf("Syntax Error at %q\nExpected any of %q, got %q", it.val, exp, it.typ.String()))
}

func (p *parser) scanSubArgumentsUntil(node *Node, stop token) {
  cont := true
  for cont {
    elt := p.currentItem()
    switch elt.typ {
    case stop:
      return
    case itemPipe:
      p.consume(1)
    case itemText:
      p.consume(1)
      if p.currentItem().typ == tokenEq {
        k := elt.val
        p.consume(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Param for %s", p.name, k), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
        node.namedParams[k] = params
        p.consume(consumed)
      } else {
        p.backup(1)
        params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous parameter", p.name), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
        node.params = append(node.params, params)
        p.consume(consumed)
      }
    default:
      params, consumed := ParseWithEnv(fmt.Sprintf("%s::Anonymous Complex parameter", p.name), p.items[p.pos:], envAlteration{exitTypes: []token{itemPipe, stop}})
      node.params = append(node.params, params)
      p.consume(consumed)
    }
  }
}

// Start at next double LF
func (p *parser) nextBlock() {
  for p.pos < len(p.items) {
    if p.currentItem().typ == tokenLF {
      p.consume(1)
      if p.currentItem().typ == tokenLF {
        p.consume(1)
        return
      }
    }
    p.consume(1)
  }
  p.pos = len(p.items) - 1
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
  glog.V(2).Infof("%s: Creating Parser (%s)\n", title, env.String())
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
      ret = p.handleParseError(err, ret)
    }
  }()

  glog.V(2).Infof("Exit sequence: %s\n", env.String())
  p.pos = 0
  it := p.currentItem()

  for it.typ != tokenEOF {
    glog.V(2).Infof("token %s\n", it.String())
    // If the exit Sequence match, abort immediately
    if len(env.exitSequence) > 0 {
      matching := 0
      for _, ty := range env.exitSequence {
        if p.currentItem().typ == ty {
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

    var n Node = Node{typ: nodeInvalid}
    switch it.typ {
    case itemText:
      n = Node{typ: nodeText, val: it.val}
    case linkStart:
      n = p.ParseLink()
    case templateStart:
      n = p.ParseTemplate()
    case tokenEq:
      n = Node{typ: nodeText, val: "="}
      if p.pos == 0 {
        n = p.parseTitle()
      }
    case tokenLF:
      glog.V(2).Info("LF")

      p.consume(1)
      if p.currentItem().typ == tokenEq {
        n = p.parseTitle()
      } else {
        p.backup(1)
        n = Node{typ: nodeText, val: "\n"}
      }
    default:
      glog.V(2).Infof("UNK", it.String())
      n = Node{typ: nodeUnknown, val: it.val}
    }
    glog.V(2).Infoln("Appending", n.String())
    ret = append(ret, n)

    it = p.currentItem()
    for _, typ := range env.exitTypes {
      if it.typ == token(typ) {
        return ret
      }
    }

    p.consume(1)
    it = p.currentItem()

  }

  if (len(env.exitSequence) > 0 || len(env.exitTypes) > 0) && p.currentItem().typ == tokenEOF {
    outOfBoundsPanic(p, len(p.items))
  }

  return ret
}
