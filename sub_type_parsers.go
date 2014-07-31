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

  item := p.eatCurrentItem()
  level := -1

  for item.typ == tokenEq {
    exitSequence = append(exitSequence, tokenEq)
    item = p.eatCurrentItem()
    level += 1
  }

  // this is an actual title
  // put last seen token in front of here
  p.backup(1)

  ret = Node{typ: nodeTitle, namedParams: make(map[string]Nodes), params: make([]Nodes, 0)}
  ret.namedParams["level"] = Nodes{Node{typ: nodeText, val: fmt.Sprintf("%d", level)}}
  content, consumed := ParseWithEnv(fmt.Sprintf("%s::Title(%d)", p.name, level),
    p.items[p.pos:],
    envAlteration{
      exitSequence:    exitSequence,
      forbiddenMarkup: []markup{title}})
  ret.namedParams["title"] = content
  p.consume(consumed)

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

  return ret
}
