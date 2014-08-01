package wikimediaparser

import (
  "fmt"
  "log"
  "os"
)

type parser struct {
  name            string
  items           []item
  start           int
  pos             int
  consumed        int
  logger          *log.Logger
  ignoreNextBlock bool
}

func create_parser(name string, tokens []item) *parser {
  p := &parser{
    name:            name,
    items:           tokens,
    start:           0,
    pos:             0,
    consumed:        0,
    logger:          log.New(os.Stdout, "[Parse]\t", log.LstdFlags),
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
    if p.eatCurrentItem().typ == tokenLF {
      if p.eatCurrentItem().typ == tokenLF {
        return
      }
    }
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
  fmt.Printf("%s: Creating Parser (%s)\n", title, env.String())
  ret = make([]Node, 0)

  ret = p.parse(env)
  fmt.Printf("%s: Consumed %d / %d\n", title, p.consumed, len(p.items))
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

  fmt.Println("Exit sequence", env.String())

  for p.pos < len(p.items) {
    it := p.eatCurrentItem()
    fmt.Printf("token %s\n", it.String())
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
        fmt.Println("Found exit sequence", p.items[p.pos])
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
      p.backup(1)
      n = p.ParseLink()
    case templateStart:
      p.backup(1)
      n = p.ParseTemplate()
    case tokenEOF:
      fmt.Println("EOF...")
      continue
    case tokenEq:
      n = Node{typ: nodeText, val: "="}
      if p.pos == 1 {
        p.backup(1)

        n = p.parseTitle()
      }
    case tokenLF:
      fmt.Println("LF")
      fmt.Println(p.currentItem())

      p.consume(1)
      if p.currentItem().typ == tokenEq {
        p.backup(1)
        n = p.parseTitle()
      } else {
        fmt.Println(p.currentItem())
        p.backup(1)
        n = Node{typ: nodeText, val: "\n"}
      }
    default:
      fmt.Printf("UNK", it.String())
      n = Node{typ: nodeUnknown, val: it.val}
    }
    fmt.Println("Appending", n.String())
    ret = append(ret, n)

    it = p.currentItem()
    for _, typ := range env.exitTypes {
      if it.typ == token(typ) {
        return ret
      }
    }

  }

  if (len(env.exitSequence) > 0 || len(env.exitTypes) > 0) && p.currentItem().typ == tokenEOF {
    outOfBoundsPanic(p, len(p.items))
  }

  return ret
}
