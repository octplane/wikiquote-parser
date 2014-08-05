package wikimediaparser

import (
  "fmt"
)

// Node as it is emitted by the parser
//    - contains a NodeType for clear identification
//    - a string val Val
//    - a list of named parameters which are actually Node Lists
//    -a list of anonymous parameters, a Node list again
type Node struct {
  Typ         nodeType
  Val         string
  NamedParams map[string]Nodes
  Params      []Nodes
}

func (n *Node) String() string {
  switch n.Typ {
  case NodeText, NodeInvalid:
    return fmt.Sprintf("%q", n.Val)
  }
  o := fmt.Sprintf("Node: %s\n", n.Typ.String())
  for ix, p := range n.Params {
    o += fmt.Sprintf("\t%d : %s\n", ix, p.String())
  }

  for k, v := range n.NamedParams {
    o += fmt.Sprintf("\t%s : %s\n", k, v.String())
  }
  return o
}

// StringParam  returns the string value of a given named parameter
func (n *Node) StringParam(k string) string {
  return n.NamedParams[k][0].Val
}

type nodeType int

const (
  NodeInvalid = nodeType(iota)
  NodeText
  NodeTitle
  NodeLink
  NodeTemplate
  NodeEq
  NodeUnknown
)

func (n nodeType) String() string {
  switch n {
  case NodeText:
    return "Text"
  case NodeLink:
    return "Link"
  case NodeTemplate:
    return "Template"
  case NodeEq:
    return " EQ "
  case NodeTitle:
    return " Title "
  case NodeUnknown:
    return " UNK "
  case NodeInvalid:
    return " INV "
  default:
    return "????"
  }
}
