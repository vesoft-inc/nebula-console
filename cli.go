package main

import (
	"io"
	"bufio"
	"fmt"
	"path"
	"log"
	"os"

	readline "github.com/shylock-hg/readline"
)

const ttyColorPrefix = "\033["
const ttyColorSuffix = "m"
const ttyColorRed = "31"
const ttyColorBold = "1"
const ttyColorReset = "0"

var completer = readline.NewPrefixCompleter(
	// show
	readline.PcItem("SHOW",
		readline.PcItem("HOSTS"),
		readline.PcItem("SPACES"),
		readline.PcItem("PARTS"),
		readline.PcItem("TAGS"),
		readline.PcItem("EDGES"),
		readline.PcItem("USERS"),
		readline.PcItem("ROLES"),
		readline.PcItem("USER"),
		readline.PcItem("CONFIGS"),
	),

	// describe
	readline.PcItem("DESCRIBE",
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("SPACE"),
	),
	readline.PcItem("DESC",
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("SPACE"),
	),
	// get configs
	readline.PcItem("GET",
		readline.PcItem("CONFIGS"),
	),
	// create
	readline.PcItem("CREATE",
		readline.PcItem("SPACE"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("USER"),
	),
	// drop
	readline.PcItem("DROP",
		readline.PcItem("SPACE"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("USER"),
	),
	// alter
	readline.PcItem("ALTER",
		readline.PcItem("USER"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
	),

	// insert
	readline.PcItem("INSERT",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// update
	readline.PcItem("UPDATE",
		readline.PcItem("CONFIGS"),
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// upsert
	readline.PcItem("UPSERT",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// delete
	readline.PcItem("DELETE",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),

	// grant
	readline.PcItem("GRANT",
		readline.PcItem("ROLE"),
	),
	// revoke
	readline.PcItem("REVOKE",
		readline.PcItem("ROLE"),
	),
	// change password
	readline.PcItem("CHANGE",
		readline.PcItem("PASSWORD"),
	),
)

func promptString(space string, user string, isErr bool, isTTY bool) string {
	prompt := ""
	// (user@nebula) [(space)] >
	if isTTY {
		prompt += fmt.Sprintf("%s%s%s", ttyColorPrefix, ttyColorBold, ttyColorSuffix)
	}
	if isTTY && isErr {
		prompt += fmt.Sprintf("%s%s%s", ttyColorPrefix, ttyColorRed, ttyColorSuffix)
	}
	prompt += fmt.Sprintf("(%s@%s) [(%s)]> ", user, NebulaLabel, space)
	if isTTY {
		prompt += fmt.Sprintf("%s%s%s", ttyColorPrefix, ttyColorReset, ttyColorSuffix)
	}
	return prompt
}

type Cli interface {
	ReadLine() (/*line*/ string, /*err*/ error, /*exit*/ bool)
	Interactive() bool
	SetisErr(bool)
	SetSpace(string)
}

// interactive
type iCli struct {
	input *readline.Instance
	user string
	space string
	isErr bool
	isTTY bool
}

func NewiCli(home string, user string) *iCli {
	r, err := readline.NewEx(&readline.Config{
			// See https://github.com/chzyer/readline/issues/169
			Prompt:          nil,
			HistoryFile:     path.Join(home, ".nebula_history"),
			AutoComplete:    completer,
			InterruptPrompt: "^C",
			EOFPrompt:       "",
			HistorySearchFold:   true,
			FuncFilterInputRune: nil,
		})
	if err != nil {
		log.Fatalf("Create readline failed, %s.", err.Error())
	}
	isTTY := readline.IsTerminal(int(os.Stdout.Fd()))
	icli := &iCli{r, user, "", false,isTTY}
	icli.input.SetPrompt(func() []rune {
		return []rune(promptString(icli.space, icli.user, icli.isErr, icli.isTTY))
	})
	return icli
}

func (l *iCli) SetSpace(space string) {
	l.space = space
}

func (l *iCli) SetisErr(isErr bool) {
	l.isErr = isErr
}

func (l iCli) ReadLine() (string, error, bool) {
	get, err := l.input.Readline()
	if err == io.EOF || err == readline.ErrInterrupt {
		// Ending not error
		return get, nil, true
	}
	if err != nil {
		return get, err, true
	}
	return get, err, false
}

func (l iCli) Interactive() bool {
	return true
}

// non-interactive
type nCli struct {
	io *bufio.Reader
}

func NewnCli(i io.Reader) nCli {
	return nCli{bufio.NewReader(i)}
}

func (l nCli) ReadLine() (string, error, bool) {
	s, _, e := l.io.ReadLine()
	if e == io.EOF {
		return string(s), nil, true
	}
	if e != nil {
		return string(s), e, true
	}
	return string(s), e, false
}

func (l nCli) Interactive() bool {
	return false
}

func (l nCli) SetSpace(space string) {
	// nothing
}

func (l nCli) SetisErr(isErr bool) {
	// nothing
}