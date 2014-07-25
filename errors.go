package wikimediaparser

import (
  "math"
)

type behaviorOnError int

const (
  abortBehavior = iota
  ignoreSectionBehavior
)

type inspectable struct {
  dumped   bool
  parser   *parser
  delta    int // how many token to dump
  start    int
  err      interface{}
  behavior behaviorOnError
}

func createInspectable(p *parser, e interface{}) inspectable {
  is := inspectable{dumped: false, parser: p, delta: 8, start: p.pos, err: e, behavior: abortBehavior}
  return is
}

func (me inspectable) handle() {
  switch me.err.(type) {
  case inspectable:
    me.dumped = me.err.(inspectable).dumped
  }
  me.handleError()
}

func (me *inspectable) handleError() {
  me.parser.logger.Printf("Syntax error at position %d\n", me.start)
  if me.err.(type) != inspectable {
    me.parser.logger.Printf("Original error was: %+v", me.err)
  }
  me.dumpStream()
}

func (me *inspectable) dumpStream() {
  if !me.dumped {
    me.dumped = true
    me.parser.logger.Println("Inspecting stack during known Exception")
    me.parser.inspectHilight(me.delta*2, me.delta)
  }
}

func (p *parser) handleTitleError(pos int, lvl int) {
  if err := recover(); err != nil {
    myerr := createInspectable(p, err)
    panic(myerr)

    // level := "?"
    // if lvl > 0 {
    //   level = fmt.Sprintf("%d", lvl)
    // }
    // p.syntaxEError(err, pos, "Invalid title-%s format", level)
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
  p.logger.Printf("Next %d elements in stream.\n", ahead)
  pfx := ""
  if hi > 0 {
    pfx = "    "
  }
  prefix := pfx

  if p.pos > len(p.items) {
    p.logger.Printf("Cannot inspect from here: at end of stream (pos=%d)", p.pos)
  } else {
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
  }
}

func (p *parser) inspect(ahead int) {
  p.inspectHilight(ahead, -1)
}

// called by main parser or subparser when something wrong appears
func (p *parser) handleParseError() {
  if err := recover(); err != nil {
    switch err.(type) {
    case inspectable:
      err.(inspectable).handle()
    default:
      p.syntaxEError(err, p.pos, "Error not handled by the parser :-/")
    }
  }
}
