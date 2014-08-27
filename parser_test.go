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
  assertEqual(t, "number of tokens", 7, len(toks))
}

func TestTokenizer2(t *testing.T) {
  text := "Bar baz baz"
  toks := Tokenize(fmt.Sprintf("====%s====\n", text))
  assertEqual(t, "number of tokens", 15, len(toks))
}

func TestEqual(t *testing.T) {
  text := "Bar=baz"
  nodes := Parse(Tokenize(text))
  assertEqual(t, "number of nodes", 3, len(nodes))
}

func TestTitle(t *testing.T) {
  text := "Bar baz baz"
  nodes := Parse(Tokenize(fmt.Sprintf("====%s====\n", text)))
  assertEqual(t, "type for Node", nodeType(NodeTitle).String(), nodes[0].Typ.String())
  assertEqual(t, "Title Level for Node", "4", nodes[0].NamedParams["level"][0].Val)
}

func TestTitle2(t *testing.T) {
  nodes := Parse(Tokenize("====Titre 4====\n===Titre 3===\n"))
  assertEqual(t, "type for Node", nodeType(NodeTitle).String(), nodes[0].Typ.String())
  assertEqual(t, "Title Level for Node", "3", nodes[1].NamedParams["level"][0].Val)
}

func TestLinkParser(t *testing.T) {
  lnk := "foo"
  text := "Link to foo "
  tree := Parse(Tokenize(fmt.Sprintf("[[%s|%s]]", lnk, text)))

  assertEqual(t, "Node count", 1, len(tree))
  assertEqual(t, "Node type", nodeType(NodeLink).String(), tree[0].Typ.String())
  assertEqual(t, "Link", lnk, tree[0].StringParam("link"))
  assertEqual(t, "Text link", text, tree[0].Params[0].StringRepresentation())
}

func TestTokenize(t *testing.T) {
  s := "|{{" // [[]]}}|"
  ts := Tokenize(s)

  assertEqual(t, "number of tokens", 3, len(ts))
  assertEqual(t, "Token", token(itemPipe).String(), ts[0].Typ.String())
  assertEqual(t, "Token", token(templateStart).String(), ts[1].Typ.String())
  assertEqual(t, "Token", token(tokenEOF).String(), ts[2].Typ.String())
}

func TestTokenizeplaceHolder(t *testing.T) {
  s := "|{{{"
  ts := Tokenize(s)

  assertEqual(t, "number of tokens", 3, len(ts))
  assertEqual(t, "Token", token(itemPipe).String(), ts[0].Typ.String())
  assertEqual(t, "Token", token(placeholderStart).String(), ts[1].Typ.String())
  assertEqual(t, "Token", token(tokenEOF).String(), ts[2].Typ.String())
}

func TestTokenize2(t *testing.T) {
  s := "{{{{"
  ts := Tokenize(s)

  if len(ts) != 3 {
    t.Errorf("Unexpected item count, expected 3, got %d", len(ts))
  }

  if ts[0].Typ != placeholderStart {
    t.Errorf("Unexpected token, got %s when wanting templateStart", ts[0].Typ.String())
  }
}

func TestTokenize3(t *testing.T) {
  s := "aaa[["
  ts := Tokenize(s)

  if len(ts) != 3 {
    t.Errorf("Unexpected item count, expected 3, got %d", len(ts))
  }

  if ts[1].Typ != linkStart {
    t.Errorf("Unexpected token, got %s when wanting linkStart", ts[1].Typ.String())
  }
}

func TestTemplate(t *testing.T) {
  temp := "citation"

  tree := Parse(Tokenize(fmt.Sprintf("{{%s}}", temp)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].Typ != NodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].Typ.String())
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

  if tree[0].Typ != NodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].Typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name to %q, got %q", temp, tree[0].StringParam("name"))
  }
  assertEqual(t, "txt value", txt, tree[0].Params[0].StringRepresentation())
}

func TestTemplate3(t *testing.T) {
  temp := "citation"
  txt := "Si six scies scient six cyprès..."
  aut := "Ane onyme"

  tree := Parse(Tokenize(fmt.Sprintf("{{%s|citation=%s|author=%s}}", temp, txt, aut)))
  if len(tree) != 1 {
    t.Errorf("Unexpected node count, expected 1 node, got %d nodes.", len(tree))
  }

  if tree[0].Typ != NodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].Typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name: wanted %q, got %q", temp, tree[0].StringParam("name"))
  }
  assertEqual(t, "Template citation param", txt, tree[0].StringParam("citation"))

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

  if tree[0].Typ != NodeTemplate {
    t.Errorf("Unexpected node type, expected nodeLink, got %q.", tree[0].Typ.String())
  }
  if tree[0].StringParam("name") != temp {
    t.Errorf("Unexpected name, expected name: wanted %q, got %q", temp, tree[0].StringParam("name"))
  }
  assertEqual(t, "Citation Parameter", txt, tree[0].StringParamOrEmpty("citation"))
  assertEqual(t, "Author Parameter", aut, tree[0].StringParamOrEmpty("author"))
}

func TestComplexTemplate(t *testing.T) {
  match := "un jour peu s'en faut ma mère viendra"
  txt := fmt.Sprintf("{{Citation|%s|thumb|author=Nobody}}", match)
  tree := Parse(Tokenize(txt))

  assertEqual(t, "Anonymous Body", match, tree[0].Params[0].StringRepresentation())

  // if tree[0].Params[0][0].Val != match {
  //   t.Errorf("Invalid parameter, got %q, expected %q", match, tree[0].Params[0][0].Val)
  // }
}

func TestComplexLink(t *testing.T) {
  linkText := "&lt;center&gt;''Le Printemps''&lt;br /&gt;Pierre Auguste Cot, 1873&lt;/center&gt;"
  txt := fmt.Sprintf("[[File:1873 Pierre Auguste Cot - Spring.jpg|thumb|upright=1.8|%s]]", linkText)

  tree := Parse(Tokenize(txt))
  assertEqual(t, "Parameter", linkText, tree[0].Params[1].StringRepresentation())
}

func TestTemplateInLink(t *testing.T) {
  linkText := "{{citation|jamais sans mon poney}}"
  txt := fmt.Sprintf("[[File: secret service|%s]]", linkText)
  tree := Parse(Tokenize(txt))

  assertEqual(t, "citation first parameter", "jamais sans mon poney", tree[0].Params[0].StringRepresentation())

}

func TestNextBlockParser(t *testing.T) {
  txt := "this is some text\n\nthis is another block\n"
  toks := Tokenize(txt)
  parser := create_parser("top", toks, nil, nil, abortBehavior)
  parser.nextBlock()
  assertEqual(t, "Next block position", 9, parser.pos)
}

func TestNoNextBlockParser(t *testing.T) {
  txt := "There is no next block\n"
  toks := Tokenize(txt)
  parser := create_parser("top", toks, nil, nil, abortBehavior)
  parser.nextBlock()
  assertEqual(t, "Next block position is a EOF", len(parser.items)-1, parser.pos)
}

func TestSyntaxError(t *testing.T) {
  doc := "Some line\n==== Malformed title===\n\nAnother Block"

  p := Parse(Tokenize(doc))
  assertEqual(t, "Node count", 9, len(p))

}

func TestSyntaxError2(t *testing.T) {
  // = pipo == is actually not a broken title
  doc := "Some line\n==== Malformed title===\n\nAnother Block\n\n== This title is broken = \n\nEnd of block"

  p := Parse(Tokenize(doc))
  assertEqual(t, "Node count", 18, len(p))
}

func TestSyntaxError3(t *testing.T) {
  // = pipo == is actually not a broken title
  doc := "Nice valid text\n{{Template Name|and this complex parameter will never be closed"

  p := Parse(Tokenize(doc))
  assertEqual(t, "Node count", 7, len(p))
}

func TestSyntaxError4(t *testing.T) {
  // = pipo == is actually not a broken title
  doc := "===broken title==\nNice valid text"

  p := Parse(Tokenize(doc))
  assertEqual(t, "Node count", 7, len(p))
}

func TestNowikiMarkup(t *testing.T) {
  doc := "<nowiki>{{</nowiki>"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Nowiki ignored", 1, len(p))
}

func TestPacha(t *testing.T) {
  doc := "{{Réf Film|titre=Le Pacha|auteur=[[Michel Audiard]]|date={{œuvre|de 1968|1968}}|acteur={{w|Jean Gabin}}}}"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Node count", 1, len(p))
}

func TestAccueil(t *testing.T) {
  //  doc := "{{TemplateName|nice text{{début cadre|rouge}} {{{!}} }} Remaining text"
  doc := "{{TemplateName|before{{{!}}after}} Remaining text"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Error node present", 1, len(p[0].Params))

}

func TestTemplateRendering(t *testing.T) {
  doc := "{{Citation|citation={{Personnage|Hello}} world{{tab}}}}"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Template is rendered correctly as string", "Hello world", p[0].StringParamOrEmpty("citation"))

}

func TestLinkRendering(t *testing.T) {
  doc := "[[w:pipo|Text]]"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Link is rendered correctly as string", "Text", p.StringRepresentation())

}

func TestLink2Rendering(t *testing.T) {
  doc := "[http://url.com/ \"Minor Threat\" by Aaron Gell in ''GQ'' (March 2006)]"
  p := Parse(Tokenize(doc))
  assertEqual(t, "Parse external url", "http://url.com/", p[0].StringParamOrEmpty("link"))
  assertEqual(t, "Parse external link title", "\"Minor Threat\" by Aaron Gell in ''GQ'' (March 2006)", p[0].Params[0].StringRepresentation())

}
