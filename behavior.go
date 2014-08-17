package wikimediaparser

import (
  "fmt"
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
  case ignoreLineBehavior:
    return "Ignore until the end of current line"
  }
  panic(fmt.Sprintf("Unknown behavior %d", be))
}
