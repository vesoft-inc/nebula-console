/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package cli

import (
	"io"
	"log"
	"os"

	"github.com/dutor/liner"
	"github.com/vesoft-inc/nebula-console/completer"
)

// interactive
type iCli struct {
	status
	terminal *liner.State
}

func NewiCli(historyFile, user string) Cli {
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
	// SetMultiLineMode sets whether line is auto-wrapped. The default is false (single line)
	c.SetMultiLineMode(true)

	f, err := os.OpenFile(historyFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Open history file %s failed, %s", historyFile, err.Error())
	}
	defer f.Close()
	c.ReadHistory(f)

	return &iCli{
		status: status{
			historyFile: historyFile,
			user:        user,
			space:       "(none)",
			promptLen:   -1,
			promptColor: -1,
			line:        "",
			joined:      false,
		},
		terminal: c,
	}
}

func (l *iCli) ReadLine() (string, bool, error) {
	for {
		input, err := l.terminal.Prompt(l.status.nebulaPrompt())
		if err == nil {
			if len(input) > 0 {
				l.terminal.AppendHistory(input)
			}
			l.status.checkJoined(input)
			if l.status.joined {
				continue
			}
			return l.status.line, false, nil
		} else if err == liner.ErrPromptAborted {
			l.status.joined = false
			return "", false, nil
		} else if err == io.EOF {
			return "", true, nil
		} else {
			return "", false, err
		}
	}
}

func (l *iCli) Interactive() bool {
	return true
}

func (l *iCli) SetSpace(space string) {
	if len(space) > 0 {
		l.status.space = space
	} else {
		l.status.space = "(none)"
	}
}

func (l *iCli) Close() {
	defer l.terminal.Close()
	f, err := os.Create(l.status.historyFile)
	if err != nil {
		log.Panicf("Write history file %s failed, %s", l.status.historyFile, err.Error())
	}
	defer f.Close()
	l.terminal.WriteHistory(f)
}
