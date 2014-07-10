package wikiquote_parser

import (
  "fmt"
  "testing"
)

func TestLinkParser(t *testing.T) {
  lnk := "foo"
  text := "Link to foo "
  tree := Parse(Tokenize(fmt.Sprintf("[[%s|%s]]", lnk, text)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }
  if tree[0].typ != nodeLink {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
  }
  if tree[0].StringParam("link") != lnk {
    t.Error("Unexpected link, expected link to %q, got %q", lnk, tree[0].StringParam("link"))
  }
  if tree[0].params[0][0].val != text {
    t.Error("Unexpected link, expected text link to %q, got %q", text, tree[0].params[0][0].val)
  }
}

func TestTokenize(t *testing.T) {
  s := "|{{" // [[]]}}|"
  ts := Tokenize(s)

  if len(ts) != 3 {
    t.Errorf("Unexpected item count, expected 3, got %d", len(ts))
  }

  if ts[0].typ != itemPipe {
    t.Errorf("Unexpected token, got %s when wanting itemPipe", ts[0].typ.String())
  }
  if ts[1].typ != templateStart {
    t.Errorf("Unexpected token, got %s when wanting leftTemplate", ts[1].typ.String())
  }

  if ts[2].typ != itemEOF {
    t.Errorf("Unexpected token, got %s when wanting itemEOF", ts[2].typ.String())
  }
}

func TestTokenize2(t *testing.T) {
  s := "{{{{"
  ts := Tokenize(s)

  if len(ts) != 3 {
    t.Errorf("Unexpected item count, expected 3, got %d", len(ts))
  }

  if ts[1].typ != templateStart {
    t.Errorf("Unexpected token, got %s when wanting templateStart", ts[1].typ.String())
  }
}

func TestTokenize3(t *testing.T) {
  s := "aaa[["
  ts := Tokenize(s)

  if len(ts) != 3 {
    t.Errorf("Unexpected item count, expected 3, got %d", len(ts))
  }

  if ts[1].typ != linkStart {
    t.Errorf("Unexpected token, got %s when wanting linkStart", ts[1].typ.String())
  }
}

func TestTemplate(t *testing.T) {
  temp := "citation"

  tree := Parse(Tokenize(fmt.Sprintf("{{%s}}", temp)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].typ != nodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name to %q, got %q", temp, tree[0].StringParam("name"))
  }
}

func TestTemplate2(t *testing.T) {
  temp := "citation"
  txt := "Tant va la cruche à l'eau qu'à la fin tu me les brises."

  tree := Parse(Tokenize(fmt.Sprintf("{{%s|%s}}", temp, txt)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].typ != nodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name to %q, got %q", temp, tree[0].StringParam("name"))
  }
  if tree[0].params[0][0].val != txt {
    t.Errorf("Unexpected value to %q, got %q", txt, tree[0].val)
  }
}

func TestTemplate3(t *testing.T) {
  temp := "citation"
  txt := "Si six scies scient six cyprès..."
  aut := "Ane onyme"

  tree := Parse(Tokenize(fmt.Sprintf("{{%s|citation=%s|author=%s}}", temp, txt, aut)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].typ != nodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name: wanted %q, got %q", temp, tree[0].StringParam("name"))
  }
  if tree[0].StringParam("citation") != txt {
    t.Errorf("Unexpected citation namedParams: wanted %q, got %q", txt, tree[0].StringParam("citation"))
  }
  if tree[0].StringParam("author") != aut {
    t.Errorf("Unexpected author param: wanted %q, got %q", aut, tree[0].StringParam("author"))
  }
}

func TestTemplate4(t *testing.T) {
  temp := "citation"
  txt := "Tant va la cruche à l'eau qu'à la fin tu me les brises."
  aut := "Les Inconnus"

  source := fmt.Sprintf("{{%s|citation=%s|author=[[%s]]}}", temp, txt, aut)
  toks := Tokenize(source)

  tree := Parse(toks)
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].typ != nodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name: wanted %q, got %q", temp, tree[0].StringParam("name"))
  }
  if tree[0].StringParam("citation") != txt {
    t.Errorf("Unexpected citation namedParams: wanted %q, got %q", txt, tree[0].StringParam("citation"))
  }

  if tree[0].namedParams["author"][1].namedParams["link"][0].val != aut {
    t.Errorf("Unexpected author param: wanted %q, got %q", aut, tree[0].namedParams["author"][1].namedParams["link"][0].val)
  }
}

func TestComplexTemplate(t *testing.T) {
  match := "un jour peu s'en faut ma mère viendra"
  txt := fmt.Sprintf("{{Citation|%s|thumb|author=Nobody}}", match)
  tree := Parse(Tokenize(txt))

  if tree[0].params[0][0].val != match {
    t.Error("Invalid parameter, got %q, expected %q", match, tree[0].params[0][0].val)
  }
}

func TestComplexLink(t *testing.T) {
  linkText := "&lt;center&gt;''Le Printemps''&lt;br /&gt;Pierre Auguste Cot, 1873&lt;/center&gt;"
  txt := fmt.Sprintf("[[File:1873 Pierre Auguste Cot - Spring.jpg|thumb|upright=1.8|%s]]", linkText)

  tree := Parse(Tokenize(txt))
  if tree[0].params[1][0].val != linkText {
    t.Error("Invalid parameter, got %q, expected %q", linkText, tree[0].params[1][0].val)
  }

}
