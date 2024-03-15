/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package cli

import (
	"io"
	"log"
	"os"
)

// interactive
type iCli struct {
	status
	terminal Terminal
}

func NewiCli(historyFile, user string, enableGoPrompt bool) Cli {
	var t Terminal
	if enableGoPrompt {
		t = NewGoPromptTerminal()
	} else {
		t = NewLinerTerminal()
	}

	f, err := os.OpenFile(historyFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Open or create history file %s failed, %s", historyFile, err.Error())
	}
	defer f.Close()
	t.ReadHistory(f)

	return &iCli{
		status: status{
			historyFile:          historyFile,
			user:                 user,
			space:                "(none)",
			promptLen:            -1,
			promptColor:          -1,
			playingData:          false,
			line:                 "",
			joinedByTripleQuotes: false,
			joinedByBackSlash:    false,
		},
		terminal: t,
	}
}

func (l *iCli) Output() bool {
	return true
}

func (l *iCli) ReadLine() (string, bool, error) {
	for {
		input, err := l.terminal.Prompt(l.status.nebulaPrompt())
		if err == nil {
			l.status.checkJoined(input)
			if l.status.joinedByTripleQuotes || l.status.joinedByBackSlash {
				continue
			}
			if len(l.status.line) > 0 {
				l.terminal.AppendHistory(l.status.line)
			}
			return l.status.line, false, nil
		} else if err == ErrPromptAborted {
			l.status.joinedByTripleQuotes = false
			l.status.joinedByBackSlash = false
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

func (l *iCli) SetRespError(msg string) {
	l.status.respErr = msg
}

func (l *iCli) GetRespError() string {
	return l.status.respErr
}

func (l *iCli) SetSpace(space string) {
	if len(space) > 0 {
		l.status.space = space
	} else {
		l.status.space = "(none)"
	}
}

func (l *iCli) GetSpace() string {
	return l.status.space
}

func (l *iCli) PlayingData(b bool) {
	l.playingData = b
}

func (l iCli) IsPlayingData() bool {
	return l.playingData
}

func (l *iCli) Close() {
	defer l.terminal.Close()
	f, err := os.OpenFile(l.status.historyFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Open or create history file %s failed, %s", l.status.historyFile, err.Error())
	}
	defer f.Close()
	l.terminal.WriteHistory(f)
}
