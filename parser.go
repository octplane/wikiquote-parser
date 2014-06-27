package main

import (
  "bufio"
  "fmt"
  "launchpad.net/xmlpath"
  "log"
  "os"
  "regexp"
  "strings"
)

type Command struct {
  Cmd        string
  Parameters []string
}

func markupExtractor(body string) []Command {
  markup, _ := regexp.Compile("(?s){{[^}]+}}")

  strCommands := markup.FindAllString(body, -1)

  commands := make([]Command, len(strCommands))

  for i, cmd := range strCommands {
    args := strings.Split(cmd, "|")
    commands[i] = Command{Cmd: args[0][2:], Parameters: args[1:]}
  }
  return commands
}

func main() {
  pageXPath := xmlpath.MustCompile("/mediawiki/page")
  textXPath := xmlpath.MustCompile(("revision/text"))
  //  fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
  fi, err := os.Open("sample.xml")

  if err != nil {

    panic(err)
  }
  // close fi on exit and check for its returned error
  defer func() {
    if err := fi.Close(); err != nil {
      panic(err)
    }
  }()
  // make a read buffer
  r := bufio.NewReader(fi)

  root, err := xmlpath.Parse(r)
  if err != nil {
    log.Fatal(err)
  }
  iter := pageXPath.Iter(root)
  for iter.Next() {
    page := iter.Node()
    content, _ := textXPath.String(page)
    commands := markupExtractor(content)
    for _, cmd := range commands {
      fmt.Println(cmd)
    }
  }
}
