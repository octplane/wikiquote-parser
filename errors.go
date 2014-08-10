package wikimediaparser

import (
  "fmt"
  "github.com/golang/glog"
  "math"
)

type behaviorOnError int

const (
  abortBehavior = behaviorOnError(iota)
  ignoreSectionBehavior
  ignoreLineBehavior
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
  i.behavior = p.onError
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
  if me.behavior != abortBehavior {
    rl = noReport
  } else {
    glog.V(2).Infof("Parent Report Level: %s, and me: %s, and parser: %s", parentRL.String(), me.behavior.String(), me.parser.onError.String())
  }
  _, ok := me.err.(inspectable)
  if !ok && rl == report {
    if me.err != nil {
      glog.V(2).Infof("Original error was: %+v", me.err)
    }
    me.dumpStream()
  }
  return rl
}

func (me *inspectable) dumpStream() {
  glog.V(2).Infoln("Inspecting stack during Inspectable error")
  // rewind in stream
  // TOKENSTOKENS
  //   ˆ
  //   Error position
  // Start from:
  // max ( error position - delta, 0)
  // highlight at:
  // min ( me.delta, error position)
  me.parser.pos = int(math.Max(float64(me.pos-me.delta), 0))
  me.parser.inspectHilight(me.delta*2, int(math.Min(float64(me.delta), float64(me.pos)-1)))
}

func outOfBoundsPanic(p *parser, s int) {
  myerr := inspectable{}
  p.defaultInspectable(&myerr)
  msg := fmt.Sprintf("Out of bounds Panic (from %d : %s)", s, p.items[s:])
  myerr.message = msg
  myerr.behavior = ignoreSectionBehavior // p.onError // ignoreSectionBehavior // p.env.onError
  myerr.class = EOFException

  glog.V(2).Infoln(msg)
  panic(myerr)
}

func (p *parser) syntaxEError(err interface{}, pos int, format string, params ...interface{}) {
  delta := 4
  p.pos = int(math.Max(0, float64(pos-delta)))
  glog.V(2).Infof("Syntax error at %q (%d):", p.items[pos].String(), p.pos)
  glog.V(2).Infof(format, params...)
  p.inspectHilight(delta*2, delta)
  panic(err)
}

func (p *parser) inspectHilight(ahead int, hi int) {
  glog.V(2).Infof("Error Context:\n")
  prefix := " "
  line := "\""

  if p.pos > len(p.items) {
    glog.V(2).Infof("Cannot inspect from here: at end of stream (pos=%d)", p.pos)
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
  glog.V(2).Infoln(line + "\"")
  glog.V(2).Infoln(prefix)

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

    case ignoreLineBehavior:
      glog.V(2).Infoln("ignoreLineBehavior: Last inspectable", last_valid_inspectable)
      // Reset parser internal state
      p.pos = 0
      p.consumed = 0
      p.nextLine()
      glog.V(2).Infof("Now at position %d\n", p.pos)
      ret = make([]Node, 0)

      ret = append(ret, Node{Typ: NodeInvalid, Val: fmt.Sprintf("> (ignored until %d )<", p.pos)})
      return ret

    case ignoreSectionBehavior:
      // We were told to ignore the syntax error. We will move on until we meet 2 consecutives \n
      // and start parsing again
      glog.V(2).Infoln("ignoreSectionBehavior: Last inspectable", last_valid_inspectable)
      // Reset parser internal state
      p.pos = 0
      p.consumed = 0
      p.nextBlock()
      glog.V(2).Infof("Now at position %d\n", p.pos)
      ret = make([]Node, 0)

      ret = append(ret, Node{Typ: NodeInvalid, Val: fmt.Sprintf("> (ignored until %d )<", p.pos)})
      return ret
    }
  default:
    p.syntaxEError(err, p.pos, "Error not handled by the parser :-/")
  }
  // Go down in the error hierarchy
  return ret
}
