/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/vesoft-inc/nebula-console/completer"

	"github.com/vesoft-inc/readline"
)

const NebulaLabel = "Nebula-Console"

const ttyColorPrefix = "\033["
const ttyColorSuffix = "m"
const ttyColorRed = "31"
const ttyColorBold = "1"
const ttyColorReset = "0"

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
	ReadLine() ( /*line*/ string /*err*/, error /*exit*/, bool)
	Interactive() bool
	SetisErr(bool)
	SetSpace(string)
}

// interactive
type iCli struct {
	input *readline.Instance
	user  string
	space string
	isErr bool
	isTTY bool
}

func NewiCli(home string, user string) *iCli {
	r, err := readline.NewEx(&readline.Config{
		// See https://github.com/chzyer/readline/issues/169
		Prompt:              nil,
		HistoryFile:         path.Join(home, ".nebula_history"),
		AutoComplete:        completer.NewCompleter(),
		InterruptPrompt:     "^C",
		EOFPrompt:           "",
		HistorySearchFold:   true,
		FuncFilterInputRune: nil,
	})
	if err != nil {
		log.Fatalf("Create readline failed, %s.", err.Error())
	}
	isTTY := readline.IsTerminal(int(os.Stdout.Fd()))
	icli := &iCli{r, user, "", false, isTTY}
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
