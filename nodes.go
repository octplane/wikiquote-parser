package wikimediaparser

// Nodes : a simple list of Nodes
type Nodes []Node

func (ns Nodes) StringRepresentation() string {
  output := ""
  for _, node := range ns {
    output += node.StringRepresentation()
  }
  return output
}

func (ns Nodes) String() string {
  output := "["
  for _, node := range ns {
    if node.Val != "\"\"" {
      output += "::" + node.String()
    }
  }
  output += "]"
  return output
}

type NodesList []Nodes

func (ns NodesList) StringRepresentation() string {
  output := ""
  for _, nodes := range ns {
    output += nodes.StringRepresentation()
  }
  return output
}
