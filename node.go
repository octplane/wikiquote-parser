package wikimediaparser

import (
  "fmt"
  "strings"
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

// Return the Node text content, without any decoration
func (n *Node) StringRepresentation() string {
  switch n.Typ {
  case NodeText, NodeInvalid:
    return n.Val
  case NodeLink:
    return n.StringParamOrEmpty("link")
  default:
    return ""
  }
}

func (n *Node) String() string {
  switch n.Typ {
  case NodeText, NodeInvalid:
    return fmt.Sprintf("%q", n.Val)
  }
  o := fmt.Sprintf("%s: %s", n.Typ.String(), n.Val)
  for ix, p := range n.Params {
    o += fmt.Sprintf("\t%d: %s\n", ix, p.String())
  }

  for k, v := range n.NamedParams {
    o += fmt.Sprintf("\t%s: %s\n", k, v.String())
  }
  return o
}

// StringParam  returns the string value of a given named parameter
func (n *Node) StringParam(k string) string {
  return n.NamedParams[k][0].Val
}

func (n *Node) StringParamOrEmpty(k string) string {
  v, ok := n.NamedParams[k]
  if ok {
    ret := v.StringRepresentation()
    if strings.HasPrefix(ret, "\n") {
      return ret[1:]
    }
    return ret
  }
  return ""
}

func EmptyNode() Node {
  return Node{Typ: NodeEmpty}
}

type nodeType int

const (
  NodeInvalid = nodeType(iota)
  NodeText
  NodeTitle
  NodeLink
  NodeTemplate
  NodePlaceholder
  NodeEq
  NodeUnknown
  NodeEmpty
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
  case NodePlaceholder:
    return " Placeholder "
  case NodeUnknown:
    return "UNK"
  case NodeInvalid:
    return " INV "
  default:
    return "????"
  }
}
