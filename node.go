package wikiquote_parser

import (
  "fmt"
)

type Node struct {
  typ         nodeType
  val         string
  namedParams map[string]Nodes
  params      []Nodes
}

func (n *Node) String() string {
  switch n.typ {
  case nodeText:
    return fmt.Sprintf("%q", n.val)
  }
  o := fmt.Sprintf("Node: %s\n", n.typ.String())
  for ix, p := range n.params {
    o += fmt.Sprintf("\t%d : %s\n", ix, p.String())
  }

  for k, v := range n.namedParams {
    o += fmt.Sprintf("\t%s : %s\n", k, v.String())
  }
  return o
}

func (n *Node) StringParam(k string) string {
  return n.namedParams[k][0].val
}

type nodeType int

const (
  nodeInvalid = iota
  nodeText
  nodeTitle
  nodeLink
  nodeTemplate
  nodeEq
)

func (n nodeType) String() string {
  switch n {
  case nodeText:
    return "Text"
  case nodeLink:
    return "Link"
  case nodeTemplate:
    return "Template"
  case nodeEq:
    return " EQ "
  case nodeTitle:
    return " Title "
  default:
    return "????"
  }
}
