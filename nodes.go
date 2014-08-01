package wikimediaparser

// Nodes : a simple list of Nodes
type Nodes []Node

func (ns Nodes) String() string {
  output := "["
  for _, node := range ns {
    segment := node.String()
    if segment != "\"\"" {
      output += "::" + segment
    }
  }
  output += "]"
  return output
}
