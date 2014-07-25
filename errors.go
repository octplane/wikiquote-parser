package wikimediaparser

import (
  "fmt"
  "math"
)

type behaviorOnError int

const (
  abortBehavior = iota
  ignoreSectionBehavior
)

type inspectable struct {
  parser   *parser
  pos      int
  message  string
  delta    int // how many token to dump
  start    int
  err      interface{}
  behavior behaviorOnError
}

func (p *parser) defaultInspectable(i *inspectable) {
  i.parser = p
  i.pos = p.pos
  i.delta = 8
  i.behavior = abortBehavior
}

func createInspectable(p *parser, pos int, m string, e interface{}) inspectable {
  is := inspectable{parser: p, pos: pos - 1, message: m, delta: 8, start: p.pos, err: e, behavior: abortBehavior}
  return is
}

func (me inspectable) handle() {
  switch me.err.(type) {
  case inspectable:
    me.err.(inspectable).handleError()
  }
  me.handleError()
}

func (me inspectable) handleError() {
  me.parser.logger.Println(me.message)
  _, ok := me.err.(inspectable)
  if !ok {
    if me.err != nil {
      me.parser.logger.Printf("Original error was: %+v", me.err)
    }
    me.dumpStream()
  }
}

func (me *inspectable) dumpStream() {
  me.parser.logger.Println("Inspecting stack during known Exception")
  // rewind in stream
  // TOKENSTOKENS
  //   ˆ
  //   Error position
  // Start from:
  // max ( error position - delta, 0)
  // highlight at:
  // min ( me.delta, error position)
  me.parser.pos = int(math.Max(float64(me.pos-me.delta), 0))
  me.parser.inspectHilight(me.delta*2, int(math.Min(float64(me.delta), float64(me.pos))))
}

func (p *parser) handleTitleError(pos int, lvl int) {
  if err := recover(); err != nil {

    level := "?"
    if lvl > 0 {
      level = fmt.Sprintf("%d", lvl)
    }
    msg := fmt.Sprintf("Invalid title-%s format at position %d", level, pos)

    myerr := createInspectable(p, pos, msg, err)
    panic(myerr)
  }
}

func outOfBoundsPanic(p *parser) {
  msg := "Went too far, out of bounds"
  myerr := inspectable{}
  p.defaultInspectable(&myerr)
  myerr.message = msg
  myerr.behavior = ignoreSectionBehavior
  panic(myerr)
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
  p.logger.Printf("Error Context:\n")
  prefix := " "
  line := "\""

  if p.pos > len(p.items) {
    p.logger.Printf("Cannot inspect from here: at end of stream (pos=%d)", p.pos)
  } else {
    for pos, content := range p.items[p.pos:] {
      if pos > ahead {
        break
      }
      if pos == hi {
        prefix += "ˆ"
      } else {
        prefix += " " // * len(content.String())
      }
      line = line + content.String()
    }
  }
  p.logger.Println(line + "\"")
  p.logger.Println(prefix)

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
      insp := err.(inspectable)
      ok := true
      var behavior behaviorOnError
      for ok {
        behavior = insp.behavior
        insp, ok = insp.err.(inspectable)
      }
      switch behavior {
      case abortBehavior:
        panic("Abort")
      case ignoreSectionBehavior:
        panic("Ignore")
      }
    default:
      p.syntaxEError(err, p.pos, "Error not handled by the parser :-/")
    }
    // Go down in the error hierarchy

  }
}
