package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
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
  glog.V(2).Infof("stringRepresentation for %+v", n)
  switch n.Typ {
  case NodeText, NodeInvalid:
    return n.Val
  case NodeLink:
    if len(n.Params) > 0 {
      return n.Params[0].StringRepresentation()
    } else {
      return n.StringParamOrEmpty("link")
    }
  case NodeTemplate:
    if len(n.Params) > 0 {
      return n.Params[0].StringRepresentation()
    } else {
      return ""
    }
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
  param, ok := n.NamedParams[k]
  if !ok {
    glog.V(2).Infof("Unable to extract parameter \"%s\" for node %s", k, n.String())
  } else {
    return param.StringRepresentation()
  }
  return ""
}

func (n *Node) StringParamOrEmpty(k string) string {
  glog.V(2).Infof("StringParamOrEmpty for %s", k)
  v, ok := n.NamedParams[k]
  if ok {
    ret := v.StringRepresentation()
    return strings.Trim(ret, " \n")
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
  NodeELink
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
  case NodeELink:
    return "ELink"
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
