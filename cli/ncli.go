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
)

// non-interactive
type nCli struct {
	status
	io       *bufio.Reader
	callback func()
}

func NewnCli(i io.Reader, user string) Cli {
	return &nCli{
		status: status{
			user:        user,
			space:       "(none)",
			promptLen:   -1,
			promptColor: -1,
			line:        "",
			joined:      false,
		},
		io: bufio.NewReader(i),
	}
}

func (l *nCli) ReadLine() (string, bool, error) {
	for {
		s, _, err := l.io.ReadLine()
		input := string(s)
		if err == nil {
			fmt.Printf(l.status.nebulaPrompt())
			// not record input to historyFile now
			fmt.Println(input)
			l.status.checkJoined(input)
			if l.status.joined {
				continue
			}
			return l.status.line, false, nil
		} else if err == io.EOF {
			return "", true, nil
		} else {
			return "", false, err
		}
	}
}

func (l *nCli) Interactive() bool {
	return false
}

func (l *nCli) SetSpace(space string) {
	// nothing
}

func (l *nCli) Close() {
	// nothing
}
