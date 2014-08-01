package wikimediaparser

import (
  "fmt"
)

type markup int

const (
  title = markup(iota)
  link
  template
)

func (p *parser) parseTitle() (ret Node) {
  exitSequence := make([]token, 0)

  fmt.Println("Parsing title 1", p.items[p.pos:])
  item := p.eatCurrentItem()
  level := 0

  for item.typ == tokenEq {
    exitSequence = append(exitSequence, tokenEq)
    item = p.eatCurrentItem()
    level += 1
  }

  exitSequence = append(exitSequence, tokenLF)

  // this is an actual title
  // put last seen token in front of here
  if level > 0 {
    p.backup(1)
  }
  fmt.Println("Parsing title", p.items[p.pos:])
  ret = Node{typ: nodeTitle, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}
  ret.namedParams["level"] = Nodes{Node{typ: nodeText, val: fmt.Sprintf("%d", level)}}
  content, consumed := ParseWithEnv(fmt.Sprintf("%s::Title(%d)", p.name, level),
    p.items[p.pos:],
    envAlteration{
      exitSequence:    exitSequence,
      forbiddenMarkup: []markup{title}})
  ret.namedParams["title"] = content
  // Only one invalid node in return, insert instead of element here
  if len(content) == 1 && content[0].typ == nodeInvalid {
    ret = content[0]
    fmt.Printf("will insert an invalid node", ret)
  }

  // Be cool with next reader, give him the LF
  p.consume(consumed - 2)
  fmt.Println("And now", p.items[p.pos:], consumed)

  return ret
}

func (p *parser) ParseLink() Node {
  ret := Node{typ: nodeLink, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eatUntil(linkStart)
  linkObject, consumed := ParseWithEnv(fmt.Sprintf("%s::Link", p.name), p.items[p.pos:],
    envAlteration{
      exitTypes:       []token{itemPipe, linkEnd},
      forbiddenMarkup: []markup{link}})
  ret.namedParams["link"] = linkObject
  p.consume(consumed)

  p.scanSubArgumentsUntil(&ret, linkEnd)
  // eat Link End
  p.consume(1)

  return ret
}

func (p *parser) ParseTemplate() Node {
  ret := Node{typ: nodeTemplate, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}

  p.eatUntil(templateStart)
  name, consumed := ParseWithEnv(fmt.Sprintf("%s::Template", p.name), p.items[p.pos:],
    envAlteration{
      exitTypes:       []token{itemPipe, templateEnd},
      forbiddenMarkup: []markup{template}})
  ret.namedParams["name"] = name
  p.consume(consumed)
  p.scanSubArgumentsUntil(&ret, templateEnd)
  p.consume(1)

  return ret
}
