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
	"strings"

	"github.com/peterh/liner"
	"github.com/vesoft-inc/nebula-console/completer"
)

type Cli interface {
	ReadLine() ( /*line*/ string /*err*/, error /*exit*/, bool)
	Interactive() bool
	SetSpace(string)
	Close()
}

type status struct {
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

func (stat *status) checkJoined(input string) {
	runes := []rune(input)
	var backSlashFound = false
	if len(runes) > 1 && runes[len(runes)-1] == 92 { // '\'
		backSlashFound = true
	}
	if stat.joined {
		if backSlashFound {
			stat.line += string(runes[:len(runes)-1])
		} else {
			stat.line += string(runes)
			stat.joined = false
		}
	} else {
		if backSlashFound {
			stat.line = string(runes[:len(runes)-1])
			stat.joined = true
		} else {
			stat.line = string(runes)
		}
	}
}

func (stat *status) nebulaPrompt() string {
	//ttyColor := prompter.color + 31
	//prompter.color = (prompter.color + 1) % 6
	prompt := ""
	//prompt += fmt.Sprintf("\033[%v;1m", ttyColor)
	if stat.joined {
		prompt += strings.Repeat(" ", stat.promptLen-3)
		prompt += "-> "
	} else {
		promptString := fmt.Sprintf("(%s@nebula) [%s]> ", stat.user, stat.space)
		stat.promptLen = len(promptString)
		prompt += promptString
	}
	//prompt += "\033[0m"
	return prompt
}

// interactive
type iCli struct {
	terminal *liner.State
	stat     status
}

func NewiCli(historyFile, user string) *iCli {
	c := liner.NewLiner()
	c.SetCtrlCAborts(true)
	// Two tab styles are currently available:
	// 1.TabCircular cycles through each completion item and displays it directly on
	// the prompt.
	// 2.TabPrints prints the list of completion items to the screen after a second
	// tab key is pressed. This behaves similar to GNU readline and BASH (which
	// uses readline).
	// TabCircular is the default style.
	c.SetTabCompletionStyle(liner.TabPrints)
	c.SetWordCompleter(completer.NewCompleter)

	if f, err := os.OpenFile(historyFile, os.O_RDONLY|os.O_CREATE, 0666); err != nil {
		log.Panicf("Open history file %s failed, %s", historyFile, err.Error())
	} else {
		defer f.Close()
		c.ReadHistory(f)
	}
	icli := &iCli{
		terminal: c,
		stat: status{
			historyFile: historyFile,
			user:        user,
			space:       "(none)",
			promptLen:   -1,
			promptColor: -1,
			line:        "",
			joined:      false,
		},
	}
	return icli
}

func (l *iCli) ReadLine() (string, error, bool) {
	for {
		input, err := l.terminal.Prompt(l.stat.nebulaPrompt())
		if err == nil {
			if len(input) > 0 {
				l.terminal.AppendHistory(input)
			}
			l.stat.checkJoined(input)
			if l.stat.joined {
				continue
			}
			return l.stat.line, nil, false
		} else if err == liner.ErrPromptAborted {
			l.stat.joined = false
			return "", nil, false
		} else if err == io.EOF {
			return "", nil, true
		} else {
			return "", err, false
		}
	}
}

func (l iCli) Interactive() bool {
	return true
}

func (l *iCli) SetSpace(space string) {
	if len(space) > 0 {
		l.stat.space = space
	} else {
		l.stat.space = "(none)"
	}
}

func (l *iCli) Close() {
	defer l.terminal.Close()
	if f, err := os.Create(l.stat.historyFile); err != nil {
		log.Panicf("Write history file %s failed, %s", l.stat.historyFile, err.Error())
	} else {
		defer f.Close()
		l.terminal.WriteHistory(f)
	}
}

// non-interactive
type nCli struct {
	io   *bufio.Reader
	stat status
}

func NewnCli(i io.Reader, user string) *nCli {
	ncli := &nCli{
		io: bufio.NewReader(i),
		stat: status{
			user:        user,
			space:       "(none)",
			promptLen:   -1,
			promptColor: -1,
			line:        "",
			joined:      false,
		},
	}
	return ncli
}

func (l *nCli) ReadLine() (string, error, bool) {
	for {
		s, _, err := l.io.ReadLine()
		input := string(s)
		if err == nil {
			fmt.Printf(l.stat.nebulaPrompt())
			// not record input to historyFile now
			fmt.Println(input)
			l.stat.checkJoined(input)
			if l.stat.joined {
				continue
			}
			return l.stat.line, nil, false
		} else if err == io.EOF {
			return "", nil, true
		} else {
			return "", err, false
		}
	}
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
