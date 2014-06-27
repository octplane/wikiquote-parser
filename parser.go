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

type Book struct {
  Title  string
  Editor string
  Year   int
  Page   int
  Isbn   string
}

type Quote struct {
  Text   string
  Author string
  Book   Book
  Topic  string
}

func markupExtractor(body string) []Command {
  markup, _ := regexp.Compile("(?s){{[^}]+}}")

  strCommands := markup.FindAllString(body, -1)

  commands := make([]Command, len(strCommands))

  for i, cmd := range strCommands {
    cmd = cmd[2 : len(cmd)-2]
    args := strings.Split(cmd, "|")
    cmd := strings.ToLower(args[0])

    commands[i] = Command{Cmd: cmd, Parameters: args[1:]}
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
    var buffer *Command = nil
    for _, cmd := range commands {
      if len(cmd.Cmd) > 2 && cmd.Cmd[0:4] == "réf" {
        if buffer != nil {
          fmt.Println("Quote")
        } else {
          fmt.Println("Found ref without quote")
          fmt.Println(cmd)
        }
        buffer = nil
      } else if strings.Index(cmd.Cmd, "citation") == 0 {
        buffer = &cmd
      } else {
        fmt.Println("Unknown command")
        fmt.Println(cmd.Cmd)
        buffer = nil
      }
    }
  }
}
