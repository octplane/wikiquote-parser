package wikimediaparser

import (
  "fmt"
  "testing"
)

func assertEqual(t *testing.T, desc string, expected interface{}, provided interface{}) {
  if expected != provided {
    t.Errorf("Unexpected %s: expected %s, got %s", desc, expected, provided)
  }
}

func assertException(t *testing.T, desc string, cls inspectableClass) {
  if err := recover(); err != nil {
    is, ok := err.(inspectable)
    if ok {
      if is.class == cls {
        return
      } else {
        t.Errorf("Expected exception %s, got exception %s", cls.String(), is.class.String())
      }
    } else {
      t.Errorf("Expected exception %s, got %+v", cls.String(), err)
    }
  }
  t.Errorf("Expected exception %s, got nothing", desc)
}

func TestTokenizer1(t *testing.T) {
  text := "Bar baz baz"
  toks := Tokenize(fmt.Sprintf("%s\n", text))
  assertEqual(t, "number of tokens", 3, len(toks))
}

func TestTokenizer2(t *testing.T) {
  text := "Bar baz baz"
  toks := Tokenize(fmt.Sprintf("====%s====\n", text))
  assertEqual(t, "number of tokens", 11, len(toks))
}

func TestEqual(t *testing.T) {
  text := "Bar=baz"
  nodes := Parse(Tokenize(text))
  fmt.Println(nodes)
  assertEqual(t, "number of nodes", 3, len(nodes))
}

func TestTitle(t *testing.T) {
  text := "Bar baz baz"
  nodes := Parse(Tokenize(fmt.Sprintf("====%s====\n", text)))
  fmt.Println(nodes.String())
  assertEqual(t, "type for Node", nodeType(nodeTitle).String(), nodes[0].typ.String())
  assertEqual(t, "Title Level for Node", "4", nodes[0].namedParams["level"][0].val)
}

func TestTitle2(t *testing.T) {
  nodes := Parse(Tokenize("====Titre 4====\n===Titre 3===\n"))
  fmt.Println(nodes)
  assertEqual(t, "type for Node", nodeType(nodeTitle).String(), nodes[0].typ.String())
  assertEqual(t, "Title Level for Node", "3", nodes[1].namedParams["level"][0].val)
}

func TestLinkParser(t *testing.T) {
  lnk := "foo"
  text := "Link to foo "
  tree := Parse(Tokenize(fmt.Sprintf("[[%s|%s]]", lnk, text)))

  assertEqual(t, "Node count", 1, len(tree))
  assertEqual(t, "Node type", nodeType(nodeLink).String(), tree[0].typ.String())
  assertEqual(t, "Link", lnk, tree[0].StringParam("link"))
  assertEqual(t, "Text link", text, tree[0].params[0][0].val)
}

func TestTokenize(t *testing.T) {
  s := "|{{" // [[]]}}|"
  ts := Tokenize(s)

  assertEqual(t, "number of tokens", 3, len(ts))
  assertEqual(t, "Token", token(itemPipe).String(), ts[0].typ.String())
  assertEqual(t, "Token", token(templateStart).String(), ts[1].typ.String())
  assertEqual(t, "Token", token(tokenEOF).String(), ts[2].typ.String())
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

  if tree[0].namedParams["author"][0].namedParams["link"][0].val != aut {
    t.Errorf("Unexpected author param: wanted %q, got %q", aut, tree[0].namedParams["author"][1].namedParams["link"][0].val)
  }
}

func TestComplexTemplate(t *testing.T) {
  match := "un jour peu s'en faut ma mère viendra"
  txt := fmt.Sprintf("{{Citation|%s|thumb|author=Nobody}}", match)
  tree := Parse(Tokenize(txt))

  if tree[0].params[0][0].val != match {
    t.Errorf("Invalid parameter, got %q, expected %q", match, tree[0].params[0][0].val)
  }
}

func TestComplexLink(t *testing.T) {
  linkText := "&lt;center&gt;''Le Printemps''&lt;br /&gt;Pierre Auguste Cot, 1873&lt;/center&gt;"
  txt := fmt.Sprintf("[[File:1873 Pierre Auguste Cot - Spring.jpg|thumb|upright=1.8|%s]]", linkText)

  tree := Parse(Tokenize(txt))
  if tree[0].params[1][0].val != linkText {
    t.Errorf("Invalid parameter, got %q, expected %q", linkText, tree[0].params[1][0].val)
  }

}

func TestNextBlockParser(t *testing.T) {
  txt := "this is some text\n\nthis is another block\n"
  toks := Tokenize(txt)
  parser := create_parser("top", toks)
  parser.nextBlock()
  assertEqual(t, "Next block position", 3, parser.pos)
}

func TestNoNextBlockParser(t *testing.T) {
  txt := "There is no next block\n"
  toks := Tokenize(txt)
  parser := create_parser("top", toks)
  parser.nextBlock()
  assertEqual(t, "Next block position is a EOF", len(parser.items)-1, parser.pos)
}

func TestSyntaxError(t *testing.T) {
  doc := "Some line\n==== Malformed title===\n\nAnother Block"

  p := Parse(Tokenize(doc))
  fmt.Println(p)
  assertEqual(t, "Node count", 4, len(p))

}

func TestSyntaxError2(t *testing.T) {
  fmt.Println("HAHA")
  // = pipo == is actually not a broken title
  doc := "Some line\n==== Malformed title===\n\nAnother Block\n\n== This title is broken = \n\nEnd of block"

  p := Parse(Tokenize(doc))
  fmt.Println(p)
  assertEqual(t, "Node count", 7, len(p))

}
