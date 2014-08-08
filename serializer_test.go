package wikimediaparser

import (
  "fmt"
  "testing"
)

func TestNodeRepresentation(t *testing.T) {
  linkText := "this is a link"
  tree := Parse(Tokenize(fmt.Sprintf("[[%s]]", linkText)))

  assertEqual(t, "Link representation", linkText, tree.StringRepresentation())
}

func TestNodeRepresentation2(t *testing.T) {
  linkText := "Marc Lévy"
  tree := Parse(Tokenize(fmt.Sprintf("{{Réf Livre|titre=Et si c'était vrai...|auteur=[[%s]]}}", linkText)))

  assertEqual(t, "Link representation ", linkText, tree[0].StringParamOrEmpty("auteur"))
}
