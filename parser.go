package main

import (
  "bufio"
  "fmt"
  "launchpad.net/xmlpath"
  "log"
  "os"
  "regexp"
  "sort"
  "strings"
)

type Command struct {
  Cmd            string
  Arguments      []string
  NamedArguments map[string]string
}

func (cmd *Command) ArgOrEmpty(key string) string {
  return cmd.ArgOrElse(key, "")
}

func (cmd *Command) ArgOrElse(key string, def string) string {
  ret, has := cmd.NamedArguments[key]
  if !has {
    return def
  }
  return ret
}

type Book struct {
  Title  string
  Author string
  Editor string
  Year   string
  Page   string
  Isbn   string
}

type Quote struct {
  Text   string
  Author string
  Book   Book
}

type BookQuote struct {
  Book  Book
  Quote Quote
}

func (bq *BookQuote) ToString() string {
  var out string
  out = fmt.Sprintf("Quote: \"%s\", by %s, from %s, page, %d published in %d", bq.Quote.Text, bq.Quote.Author, bq.Book.Title, bq.Book.Page, bq.Book.Year)
  return out
}

// Thank you andrew https://groups.google.com/d/msg/golang-nuts/FT7cjmcL7gw/Gj4_aEsE_IsJ
// A data structure to hold a key/value pair.
type Pair struct {
  Key   string
  Value int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it.
func sortMapByValue(m map[string]int) PairList {
  p := make(PairList, len(m))
  i := 0
  for k, v := range m {
    p[i] = Pair{k, v}
    i += 1
  }
  sort.Sort(sort.Reverse(p))
  return p
}

func markupExtractor(body string) []Command {
  markup := regexp.MustCompile("(?s){{[^}]+}}")
  param := regexp.MustCompile("([^=]+)=(.*)")

  strCommands := markup.FindAllString(body, -1)

  commands := make([]Command, len(strCommands))
  pos := 0

  for _, cmd := range strCommands {
    cmd = cmd[2 : len(cmd)-2]

    arguments := make([]string, 10000)
    argumentsIndex := 0
    kvArguments := make(map[string]string, 1000)

    for _, arg := range strings.Split(cmd, "|") {
      kv := param.FindStringSubmatch(arg)
      if len(kv) == 3 {
        key := strings.TrimSpace(strings.ToLower(kv[1]))
        val, exists := kvArguments[key]
        if exists && val != kv[0] {
          // FIXME: handle the issue
          // panic(fmt.Sprintf("Parameter %s already exists with value \"%s\", here wants : \"%s\"", key, val, kv[0]))
        } else {
          kvArguments[key] = kv[2]
        }
      } else {
        arguments[argumentsIndex] = arg
        argumentsIndex += 1
      }
    }

    cmd := strings.TrimSpace(strings.ToLower(arguments[0]))

    // Parse special "defaultsort:", "if:", "msg:" commands
    if strings.Index(cmd, ":") != -1 {
      cmd = cmd[0:strings.Index(cmd, ":")]
      // FIXME inject arguments in Command anyway
    }

    // Ignore the empty command
    if cmd != "" {
      commands[pos] = Command{Cmd: cmd, Arguments: arguments[1:], NamedArguments: kvArguments}
      pos += 1
    }
  }
  return commands
}

func BuildQuote(qCommand Command, reference Command) {
  quote, hasQuote := qCommand.NamedArguments["citation"]
  if !hasQuote {
    quote = qCommand.Arguments[0]
    hasQuote = true
  }

  author := reference.ArgOrEmpty("auteur")
  title := reference.ArgOrEmpty("titre")
  page := reference.ArgOrEmpty("page")
  editor := reference.ArgOrEmpty("éditeur")
  year := reference.ArgOrEmpty("année")
  isbn := reference.ArgOrEmpty("isbn")

  if quote == "" || author == "" || title == "" || page == "" || editor == "" || year == "" || isbn == "" {
    fmt.Println("Some fields are missing")
    fmt.Println("", reference)
  }

}

func ExtractQuotes(commands []Command) {
  var buffer *Command = nil
  for ix, cmd := range commands {
    if len(cmd.Cmd) > 2 && cmd.Cmd[0:4] == "réf" {
      if buffer != nil {
        BuildQuote(*buffer, cmd)
      } else {
        fmt.Println("Found ref without quote")
        fmt.Println(cmd)
      }
      buffer = nil
    } else if strings.Index(cmd.Cmd, "citation") == 0 {
      buffer = &commands[ix]
    } else {
      fmt.Println("Unknown command:", cmd.Cmd)
      buffer = nil
    }
  }
}

func ExtractStats(commands []Command) {
  commandPopularity := make(map[string]int, 1000)
  argsPopularity := make(map[string]map[string]int, 100)

  var count int
  var hasCommand bool
  var aCount int
  var hasArg bool

  for _, cmd := range commands {
    count, hasCommand = commandPopularity[cmd.Cmd]
    if hasCommand {
      commandPopularity[cmd.Cmd] = count + 1
    } else {
      commandPopularity[cmd.Cmd] = 1
      argsPopularity[cmd.Cmd] = make(map[string]int, 100)
    }

    for k := range cmd.NamedArguments {
      aCount, hasArg = argsPopularity[cmd.Cmd][k]
      if hasArg {
        argsPopularity[cmd.Cmd][k] = aCount + 1
      } else {
        argsPopularity[cmd.Cmd][k] = 1
      }
    }
  }

  var maxArgsOcc int
  for _, pl := range sortMapByValue(commandPopularity) {
    fmt.Printf("### %s (%d occ.)\n", pl.Key, pl.Value)
    total := pl.Value
    maxArgsOcc = 0
    for _, ar := range sortMapByValue(argsPopularity[pl.Key]) {
      if ar.Value > maxArgsOcc {
        maxArgsOcc = ar.Value
      }
      if ar.Value > maxArgsOcc/10 && 100*ar.Value/total > 1 {
        fmt.Printf("- %s (%d%%)\n", ar.Key, 100*ar.Value/total)
      }
    }
  }
}

func main() {
  pageXPath := xmlpath.MustCompile("/mediawiki/page")
  textXPath := xmlpath.MustCompile(("revision/text"))
  fi, err := os.Open("frwikiquote-20140622-pages-articles-multistream.xml")
  // fi, err := os.Open("sample.xml")

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
  commands := make([]Command, 0)
  iter := pageXPath.Iter(root)
  for iter.Next() {
    page := iter.Node()
    content, _ := textXPath.String(page)
    commands = append(commands, markupExtractor(content)...)
  }
  ExtractStats(commands)
}
