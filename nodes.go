package wikimediaparser

// Nodes : a simple list of Nodes
type Nodes []Node

func (ns Nodes) String() string {
  output := "["
  for _, node := range ns {
    if node.Val != "\"\"" {
      output += "::" + node.Val
    }
  }
  output += "]"
  return output
}
