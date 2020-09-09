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
	ReadLine() (line string, exit bool, err error)
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
	var backSlashFound = len(runes) > 1 && runes[len(runes)-1] == 92 // '\'
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
