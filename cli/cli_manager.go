/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package cli

import (
	"fmt"
	"strings"
)

type Cli interface {
	Output() bool
	ReadLine() (line string, exit bool, err error)
	Interactive() bool
	SetRespError(msg string)
	GetRespError() string
	SetSpace(string)
	GetSpace() string
	PlayingData(bool)
	IsPlayingData() bool
	Close()
}

type status struct {
	// prompt
	historyFile string
	user        string
	space       string
	respErr     string
	playingData bool
	promptLen   int
	promptColor int

	// multi-line seperated by '\' or enclosed in triple quotes
	line string
	// tripleQuotes has a higher priority than backSlash
	joinedByTripleQuotes bool
	joinedByBackSlash    bool
}

func (stat *status) checkJoined(input string) {
	var pureInput = strings.TrimSpace(input)
	var tripleQuotesFound = len(pureInput) == 3 && (pureInput == "\"\"\"" || pureInput == "'''")
	var backSlashFound = len(input) >= 1 && input[len(input)-1] == 92 // '\'
	if stat.joinedByTripleQuotes {
		if tripleQuotesFound {
			stat.joinedByTripleQuotes = false
		} else {
			if backSlashFound { // backslash can be used in a block enclosed by threequotes
				stat.line += string(input[:len(input)-1])
			} else {
				stat.line += input
				stat.line += " "
			}
		}
	} else if stat.joinedByBackSlash {
		if backSlashFound {
			stat.line += string(input[:len(input)-1])
		} else {
			stat.line += input
			stat.joinedByBackSlash = false
		}
	} else {
		if tripleQuotesFound {
			stat.line = ""
			stat.joinedByTripleQuotes = true
		} else if backSlashFound {
			stat.line = string(input[:len(input)-1])
			stat.joinedByBackSlash = true
		} else {
			stat.line = input
		}
	}
}

func (stat *status) nebulaPrompt() string {
	//ttyColor := prompter.color + 31
	//prompter.color = (prompter.color + 1) % 6
	prompt := ""
	//prompt += fmt.Sprintf("\033[%v;1m", ttyColor)
	if stat.joinedByTripleQuotes || stat.joinedByBackSlash {
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
