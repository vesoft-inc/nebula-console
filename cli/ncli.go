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

type Cleanup func()

// non-interactive
type nCli struct {
	status
	io      *bufio.Reader
	output  bool
	cleanup Cleanup
}

func NewnCli(i io.Reader, output bool, user string, cleanup Cleanup) Cli {
	return &nCli{
		status: status{
			user:        user,
			space:       "(none)",
			respErr:     "",
			promptLen:   -1,
			promptColor: -1,
			line:        "",
			joined:      false,
		},
		io:      bufio.NewReader(i),
		output:  output,
		cleanup: cleanup,
	}
}

func readln(r *bufio.Reader) (string, error) {
	var (
		isPartial bool  = true
		err       error = nil
		line, ln  []byte
	)
	for isPartial && err == nil {
		line, isPartial, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func (l *nCli) Output() bool {
	return l.output
}

func (l *nCli) ReadLine() (string, bool, error) {
	for {
		input, err := readln(l.io)
		if err == nil {
			if l.output {
				fmt.Print(l.status.nebulaPrompt())
				// not record input to historyFile now
				fmt.Println(input)
			}
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

func (l *nCli) SetRespError(msg string) {
	l.status.respErr = msg
}

func (l *nCli) GetRespError() string {
	return l.status.respErr
}

func (l *nCli) SetSpace(space string) {
	if len(space) > 0 {
		l.status.space = space
	} else {
		l.status.space = "(none)"
	}
}

func (l *nCli) GetSpace() string {
	return l.status.space
}

func (l *nCli) Close() {
	if l.cleanup != nil {
		l.cleanup()
	}
}
