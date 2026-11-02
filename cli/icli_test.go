/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package cli

import (
	"io"
	"reflect"
	"testing"
)

type historyTestTerminal struct {
	inputs   []string
	appended []string
}

func (t *historyTestTerminal) ReadHistory(io.Reader) error {
	return nil
}

func (t *historyTestTerminal) AppendHistory(item string) {
	t.appended = append(t.appended, item)
}

func (t *historyTestTerminal) WriteHistory(io.Writer) error {
	return nil
}

func (t *historyTestTerminal) Close() {}

func (t *historyTestTerminal) Prompt(string) (string, error) {
	if len(t.inputs) == 0 {
		return "", io.EOF
	}
	input := t.inputs[0]
	t.inputs = t.inputs[1:]
	return input, nil
}

func TestIsCmdWithPassword(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{
			name: "password",
			cmd:  `CREATE USER test WITH PASSWORD "secret"`,
			want: true,
		},
		{
			name: "uppercase password",
			cmd:  `CREATE USER test WITH PASSWORD "secret"`,
			want: true,
		},
		{
			name: "mixed case password",
			cmd:  `ALTER USER test WITH PaSsWoRd "secret"`,
			want: true,
		},
		{
			name: "pwd",
			cmd:  `ALTER USER test WITH PWD "secret"`,
			want: true,
		},
		{
			name: "ordinary command",
			cmd:  "SHOW HOSTS",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCmdWithPassword(tt.cmd); got != tt.want {
				t.Errorf("isCmdWithPassword(%q) = %t, want %t", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestICliReadLineHistoryFiltering(t *testing.T) {
	tests := []struct {
		name       string
		inputs     []string
		wantLine   string
		wantHistory []string
	}{
		{
			name:     "ordinary command is appended",
			inputs:   []string{"SHOW HOSTS"},
			wantLine: "SHOW HOSTS",
			wantHistory: []string{
				"SHOW HOSTS",
			},
		},
		{
			name:     "password command is not appended",
			inputs:   []string{`CREATE USER test WITH PASSWORD "secret"`},
			wantLine: `CREATE USER test WITH PASSWORD "secret"`,
		},
		{
			name: "multiline password command is not appended",
			inputs: []string{
				"CREATE USER test WITH \\",
				`PASSWORD "secret"`,
			},
			wantLine: `CREATE USER test WITH PASSWORD "secret"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminal := &historyTestTerminal{inputs: tt.inputs}
			c := &iCli{
				status:   status{user: "root", space: "(none)", promptLen: -1},
				terminal: terminal,
			}

			gotLine, exit, err := c.ReadLine()
			if err != nil {
				t.Fatalf("ReadLine() returned an error: %v", err)
			}
			if exit {
				t.Fatal("ReadLine() unexpectedly requested exit")
			}
			if gotLine != tt.wantLine {
				t.Errorf("ReadLine() line = %q, want %q", gotLine, tt.wantLine)
			}
			if !reflect.DeepEqual(terminal.appended, tt.wantHistory) {
				t.Errorf("appended history = %#v, want %#v", terminal.appended, tt.wantHistory)
			}
		})
	}
}
