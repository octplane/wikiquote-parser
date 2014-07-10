package wikiquote_parser

type Nodes []Node

func (ns Nodes) String() string {
  output := ""
  for _, node := range ns {
    segment := node.String()
    if segment != "\"\"" {
      output += " " + node.String()
    }
  }
  return output
}
