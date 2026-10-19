/* Copyright (c) 2026 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package cli

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type historyTestTerminal struct {
	inputs  []string
	history []string
}

func (t *historyTestTerminal) ReadHistory(io.Reader) error { return nil }

func (t *historyTestTerminal) AppendHistory(item string) {
	t.history = append(t.history, item)
}

func (t *historyTestTerminal) WriteHistory(w io.Writer) error {
	for _, item := range t.history {
		if _, err := io.WriteString(w, item+"\n"); err != nil {
			return err
		}
	}
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

func TestReadLinePasswordHistory(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		line    string
		history []string
	}{
		{"create user", []string{`CREATE USER alice WITH PASSWORD "secret";`}, `CREATE USER alice WITH PASSWORD "secret";`, nil},
		{"alter user", []string{`alter user alice with password "secret";`}, `alter user alice with password "secret";`, nil},
		{"change password", []string{`ChAnGe PaSsWoRd alice FROM "old" TO "new";`}, `ChAnGe PaSsWoRd alice FROM "old" TO "new";`, nil},
		{"pwd", []string{`RETURN "PwD";`}, `RETURN "PwD";`, nil},
		{"backslash continuation", []string{`ALTER USER alice \`, `WITH PASSWORD "secret";`}, `ALTER USER alice WITH PASSWORD "secret";`, nil},
		{"triple quote continuation", []string{`"""`, `CHANGE PASSWORD alice`, `FROM "old" TO "new";`, `"""`}, `CHANGE PASSWORD alice FROM "old" TO "new"; `, nil},
		{"ordinary command", []string{"SHOW SPACES;"}, "SHOW SPACES;", []string{"SHOW SPACES;"}},
		{"ordinary continuation", []string{`SHOW \`, "SPACES;"}, "SHOW SPACES;", []string{"SHOW SPACES;"}},
		{"empty input", []string{""}, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminal := &historyTestTerminal{inputs: tt.inputs}
			historyFile := filepath.Join(t.TempDir(), "nebula_history")
			c := &iCli{status: status{historyFile: historyFile}, terminal: terminal}
			line, exit, err := c.ReadLine()
			if err != nil || exit || line != tt.line {
				t.Fatalf("ReadLine() = (%q, %v, %v), want (%q, false, nil)", line, exit, err, tt.line)
			}
			if !reflect.DeepEqual(terminal.history, tt.history) {
				t.Fatalf("history = %q, want %q", terminal.history, tt.history)
			}
			c.Close()
			data, err := os.ReadFile(historyFile)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if len(tt.history) > 0 {
				want = strings.Join(tt.history, "\n") + "\n"
			}
			if string(data) != want {
				t.Fatalf("persisted history = %q, want %q", data, want)
			}
		})
	}
}
