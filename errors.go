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

func (be behaviorOnError) String() string {
  switch be {
  case abortBehavior:
    return "Aborting Exception"
  case ignoreSectionBehavior:
    return "Ignore current section Exception"
  }
  panic(fmt.Sprintf("Unknown behavior %d", be))
}

const (
  EOFException = iota
  RuntimeException
)

type inspectableClass int

func (is inspectableClass) String() string {
  switch is {
  case EOFException:
    return "End of file Exception"
  case RuntimeException:
    return "Runtime exception"
  }

  panic(fmt.Sprintf("Unexpected inspectable class %d", is))
}

type inspectable struct {
  parser   *parser
  pos      int
  message  string
  delta    int // how many token to dump
  start    int
  err      interface{}
  behavior behaviorOnError
  class    inspectableClass
}

func (p *parser) defaultInspectable(i *inspectable) {
  i.parser = p
  i.pos = p.pos
  i.delta = 8
  i.behavior = abortBehavior
  i.class = RuntimeException
}

func createInspectable(p *parser, pos int, m string, e interface{}) inspectable {
  is := inspectable{}
  p.defaultInspectable(&is)
  is.pos = pos - 1
  is.message = m
  is.start = p.pos
  is.err = e

  return is
}

func (me inspectable) handle() {
  rL := report
  if me.behavior == ignoreSectionBehavior {
    rL = noReport
  }

  switch me.err.(type) {
  case inspectable:
    rL = me.err.(inspectable).handleError(rL)
  }
  if rL == report {
    me.handleError(rL)
  }
}

type reportLevel int

const (
  noReport = reportLevel(iota)
  report
)

func (rl reportLevel) String() string {
  switch rl {
  case noReport:
    return "No report"
  case report:
    return "Report"
  }
  panic("Unknown report level")
}

func (me inspectable) handleError(parentRL reportLevel) reportLevel {
  rl := parentRL
  if me.behavior == ignoreSectionBehavior {
    rl = noReport
  } else {
    me.parser.logger.Printf("parentrl : %s, message: %s", parentRL.String(), me.message)
  }
  _, ok := me.err.(inspectable)
  if !ok && rl == report {
    if me.err != nil {
      me.parser.logger.Printf("Original error was: %+v", me.err)
    }
    me.dumpStream()
  }
  return rl
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

func outOfBoundsPanic(p *parser, s int) {
  msg := fmt.Sprintf("Went too far, out of bounds (from %d)", s)
  myerr := inspectable{}
  p.defaultInspectable(&myerr)
  myerr.message = msg
  myerr.behavior = ignoreSectionBehavior
  myerr.class = EOFException
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
func (p *parser) handleParseError(err interface{}, ret Nodes) Nodes {
  switch err.(type) {
  case inspectable:
    err.(inspectable).handle()
    insp := err.(inspectable)
    last_valid_inspectable := insp
    ok := true
    var behavior behaviorOnError
    for ok {
      last_valid_inspectable = insp
      behavior = insp.behavior
      insp, ok = insp.err.(inspectable)
    }
    switch behavior {
    case abortBehavior:
      panic(last_valid_inspectable)
    case ignoreSectionBehavior:
      // We were told to ignore the syntax error. We will move on until we meet 2 consecutives \n
      // and start parsing again
      fmt.Println("Last inspectable", last_valid_inspectable)
      // Reset parser internal state
      p.pos = 0
      p.consumed = 0
      p.nextBlock()
      fmt.Printf("Now at position %d\n", p.pos)
      ret = make([]Node, 0)

      ret = append(ret, Node{typ: nodeInvalid, val: fmt.Sprintf("> (ignored until %d )<", p.pos)})
      return ret
    }
  default:
    p.syntaxEError(err, p.pos, "Error not handled by the parser :-/")
  }
  // Go down in the error hierarchy
  return ret
}
