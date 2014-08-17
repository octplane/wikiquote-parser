package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
)

type markup int

const (
  title = markup(iota)
  link
  template
)

func (p *parser) parseTitle() (ret Node) {
  exitSequence := make([]token, 0)

  glog.V(2).Infoln("Parsing title 1", p.items[p.pos:])
  item := p.eatCurrentItem()
  level := 0

  for item.Typ == tokenEq {
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
  glog.V(2).Infoln("Parsing title", p.items[p.pos:])
  ret = Node{Typ: NodeTitle, NamedParams: make(map[string]Nodes), Params: make([]Nodes, 0)}
  ret.NamedParams["level"] = Nodes{Node{Typ: NodeText, Val: fmt.Sprintf("%d", level)}}
  content, consumed := ParseWithEnv(fmt.Sprintf("%s::Title(%d)", p.name, level),
    p,
    p.items[p.pos:],
    nil,
    exitSequence,
    ignoreLineBehavior,
  )
  ret.NamedParams["title"] = content
  // Only one invalid node in return, insert instead of element here
  if len(content) == 1 && content[0].Typ == NodeInvalid {
    ret = content[0]
    glog.V(2).Infoln("will insert an invalid node", ret)
  }

  // Be cool with next reader, give him the LF
  p.consume(consumed - 2)
  glog.V(2).Infoln("And now", p.items[p.pos:], consumed)

  return ret
}

func (p *parser) ParseLink() Node {
  ret := Node{Typ: NodeLink, NamedParams: make(map[string]Nodes), Params: make([]Nodes, 0)}

  p.eatUntil(linkStart)
  linkObject, consumed := ParseWithEnv(fmt.Sprintf("%s::Link", p.name),
    p,
    p.items[p.pos:],
    []token{itemPipe, linkEnd}, nil, ignoreSectionBehavior)
  ret.NamedParams["link"] = linkObject
  p.consume(consumed)

  consumed = ScanSubArgumentsUntil("link", p, &ret, linkEnd)
  p.consume(consumed)
  return ret
}

func (p *parser) ParseTemplate() Node {
  name := p.name + ".innerTemplateParser"
  actualParser := p.CreateParser(name, p.items[p.pos:], nil, nil, ignoreSectionBehavior)
  node, consumed := actualParser.innerTemplateParser()
  p.consume(consumed)
  return node
}

func (p *parser) innerTemplateParser() (ret Node, consumed int) {
  ret = Node{Typ: NodeTemplate, NamedParams: make(map[string]Nodes), Params: make([]Nodes, 0)}
  consumed = 0
  glog.V(2).Infoln("Parsing a template")

  p.eatUntil(templateStart)
  name, cons1 := ParseWithEnv(fmt.Sprintf("%s::Template", p.name),
    p,
    p.items[p.pos:],
    []token{itemPipe, templateEnd}, nil, ignoreSectionBehavior)
  ret.NamedParams["name"] = name
  p.consume(cons1)
  consumed += cons1
  glog.V(2).Infof("Found template %s, now scanning sub arguments from %s", name.String(), p.items[p.pos:])
  cons1 = ScanSubArgumentsUntil("template", p, &ret, templateEnd)
  p.consume(cons1)

  consumed += cons1
  glog.V(2).Infoln("Parsed a template")
  return ret, consumed
}

func (p *parser) ParsePlaceholder() Node {
  ret := Node{Typ: NodePlaceholder, NamedParams: make(map[string]Nodes), Params: make([]Nodes, 0)}

  p.eatUntil(placeholderStart)
  name, consumed := ParseWithEnv(fmt.Sprintf("%s::Placeholder", p.name),
    p,
    p.items[p.pos:],
    []token{placeholderEnd}, nil, ignoreSectionBehavior)

  ret.NamedParams["content"] = name
  p.consume(consumed)

  return ret
}
