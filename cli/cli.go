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

	"github.com/peterh/liner"
	"github.com/vesoft-inc/nebula-console/completer"
)

var (
	ErrEOF     = io.EOF
	ErrAborted = liner.ErrPromptAborted
)

type Cli interface {
	ReadLine() ( /*line*/ string /*err*/, error /*exit*/, bool)
	Interactive() bool
	SetSpace(string)
	Close()
}

// interactive
type iCli struct {
	terminal *liner.State
	// prompt
	historyFile string
	user        string
	space       string
	promptLen   int
	promptColor int

	// multi-line seperated by '\'
	line   string
	joined bool
}

func NewiCli(historyFile, user string) *iCli {
	c := liner.NewLiner()
	c.SetCtrlCAborts(false)
	// Two tab styles are currently available:
	// 1.TabCircular cycles through each completion item and displays it directly on
	// the prompt.
	// 2.TabPrints prints the list of completion items to the screen after a second
	// tab key is pressed. This behaves similar to GNU readline and BASH (which
	// uses readline).
	// TabCircular is the default style.
	c.SetTabCompletionStyle(liner.TabPrints)
	c.SetWordCompleter(completer.NewCompleter)

	if f, err := os.Open(historyFile); err == nil {
		c.ReadHistory(f)
		f.Close()
	}
	icli := &iCli{
		terminal:    c,
		historyFile: historyFile,
		user:        user,
		space:       "(none)",
		promptLen:   -1,
		promptColor: -1,
		line:        "",
		joined:      false,
	}
	return icli
}

func (l *iCli) checkJoined(input string) {
	runes := []rune(input)
	var backSlashFound = false
	if len(runes) > 1 && runes[len(runes)-1] == 92 { // '\'
		backSlashFound = true
	}
	if l.joined {
		if backSlashFound {
			l.line += string(runes[:len(runes)-1])
		} else {
			l.line += string(runes)
			l.joined = false
		}
	} else {
		if backSlashFound {
			l.line = string(runes[:len(runes)-1])
			l.joined = true
		} else {
			l.line = string(runes)
		}
	}
}

func (l *iCli) nebulaPrompt() string {
	//ttyColor := prompter.color + 31
	//prompter.color = (prompter.color + 1) % 6
	prompt := ""
	//prompt += fmt.Sprintf("\033[%v;1m", ttyColor)
	if l.joined {
		for i := 0; i < l.promptLen-3; i++ {
			prompt += " "
		}
		prompt += "-> "
	} else {
		promptString := fmt.Sprintf("(%s@nebula) [%s]> ", l.user, l.space)
		l.promptLen = len(promptString)
		prompt += promptString
	}
	//prompt += "\033[0m"
	return prompt
}

func (l *iCli) ReadLine() (string, error, bool) {
	for {
		if input, err := l.terminal.Prompt(l.nebulaPrompt()); err == nil {
			if len(input) > 0 {
				l.terminal.AppendHistory(input)
			}
			l.checkJoined(input)
			if l.joined {
				continue
			}
			return l.line, nil, false
		} else if err == ErrAborted {
			return l.line, nil, true
		} else if err == ErrEOF {
			return l.line, nil, true
		} else {
			return l.line, err, false
		}
	}
}

func (l iCli) Interactive() bool {
	return true
}

func (l *iCli) SetSpace(space string) {
	l.space = space
}

func (l *iCli) Close() {
	if f, err := os.Create(l.historyFile); err != nil {
		log.Print("error writing history file ", l.historyFile, err)
	} else {
		l.terminal.WriteHistory(f)
		f.Close()
	}
	l.terminal.Close()
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

func (l nCli) Close() {
	// nothing
}
