package wikimediaparser

import (
  "fmt"
  "math"
)

func (p *parser) handleAbort() {
  if err := recover(); err != nil {
    p.logger.Printf("Moved to far away in parsing: %d/%d. Aborting", p.pos, len(p.items))
    panic(err)
  }
}

func (p *parser) onErrorIgnoreBlock(pos int) {
  if err := recover(); err != nil {
    p.pos = pos
    p.ignoreNextBlock = true
  }
}

func (p *parser) subParserErrorIgnoreBlock(ret []Node, pos int, env envAlteration) {
  if err := recover(); err != nil {
    p.pos = pos
    p.ignoreNextBlock = true
    n := Node{typ: nodeText, val: fmt.Sprintf("Error while trying to reach environment condition:\nFor conditions: %s\nAt position %d", env.String(), pos)}
    fmt.Println(n)
    ret = []Node{n}
  }
}

func (p *parser) handleTitleError(pos int, lvl int) {
  if err := recover(); err != nil {
    level := "?"
    if lvl > 0 {
      level = fmt.Sprintf("%d", lvl)
    }
    p.syntaxEError(err, pos, "Invalid title-%s format", level)
  }
}

func (p *parser) syntaxError(pos int, format string, params ...interface{}) {
  if err := recover(); err != nil {
    p.syntaxEError(err, pos, format, params...)
  }
}

func (p *parser) syntaxEError(err interface{}, pos int, format string, params ...interface{}) {
  delta := 4
  p.pos = int(math.Max(0, float64(pos-delta)))
  p.logger.Printf("Syntax error at %q (%d):", p.items[pos].String(), p.pos)
  p.logger.Printf(format, params...)
  p.inspectHilight(delta*2, delta)
  panic(err)
}

func (p *parser) inspectHilight(ahead int, hi int) {
  p.logger.Printf("------ Peeking ahead for %d elements\n", ahead)
  pfx := ""
  if hi > 0 {
    pfx = "    "
  }
  prefix := pfx

  for pos, content := range p.items[p.pos:] {
    if pos > ahead {
      break
    }
    if pos == hi {
      prefix = ">>> "
    } else {
      prefix = pfx
    }
    p.logger.Printf(prefix + content.String())
  }
  p.logger.Println("------ End of Peek")
}

func (p *parser) inspect(ahead int) {
  p.inspectHilight(ahead, -1)
}
