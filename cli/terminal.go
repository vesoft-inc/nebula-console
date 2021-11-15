/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/c-bata/go-prompt"
	"github.com/jievince/liner"
	"github.com/vesoft-inc/nebula-console/completer"
)

type Terminal interface {
	ReadHistory(r io.Reader) error
	AppendHistory(item string)
	WriteHistory(w io.Writer) error
	Close()
	Prompt(p string) (string, error)
}

type LinerTerminal struct {
	state *liner.State
}

// ErrPromptAborted is the error for prompt aborted. i.e. ctrl+d.
// not implement for go prompt
var ErrPromptAborted = errors.New("prompt aborted")

func NewLinerTerminal() Terminal {
	c := liner.NewLiner()
	t := &LinerTerminal{state: c}
	c.SetCtrlCAborts(true)
	// Two tab styles are currently available:
	// 1.TabCircular cycles through each completion item and displays it directly on
	// the prompt.
	// 2.TabPrints prints the list of completion items to the screen after a second
	// tab key is pressed. This behaves similar to GNU readline and BASH (which
	// uses readline).
	// TabCircular is the default style.
	c.SetTabCompletionStyle(liner.TabPrints)
	c.SetMultiLineMode(true)
	c.SetWordCompleter(completer.NewCompleter)
	return t
}

func (l *LinerTerminal) ReadHistory(r io.Reader) error {
	_, err := l.state.ReadHistory(r)
	return err
}

func (l *LinerTerminal) AppendHistory(item string) {
	l.state.AppendHistory(item)
}

func (l *LinerTerminal) WriteHistory(w io.Writer) error {
	_, err := l.state.WriteHistory(w)
	return err
}

func (l *LinerTerminal) Close() {
	l.state.Close()
}

func (l *LinerTerminal) Prompt(p string) (string, error) {
	s, err := l.state.Prompt(p)
	if err == liner.ErrPromptAborted {
		return "", ErrPromptAborted
	}
	return s, err
}

type GoPromptTerminal struct {
	prompt *prompt.Prompt
}

func goPromptSuggest(d prompt.Document) []prompt.Suggest {
	suggests := make([]prompt.Suggest, 0)
	line, pos := d.CurrentLine(), d.CursorPositionCol()
	_, completions, _ := completer.NewCompleter(line, pos)
	for _, s := range completions {
		suggests = append(suggests, prompt.Suggest{Text: s})
	}
	return suggests
}

func NewGoPromptTerminal() Terminal {
	executor := func(s string) {
	}
	p := prompt.New(executor, goPromptSuggest)
	t := &GoPromptTerminal{prompt: p}
	return t
}

// TODO should enhance go-prompt
func (g *GoPromptTerminal) ReadHistory(r io.Reader) error {
	h := make([]string, 0)
	in := bufio.NewReader(r)
	for {
		line, part, err := in.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if part {
			return fmt.Errorf("history file is too long")
		}
		if !utf8.Valid(line) {
			return fmt.Errorf("invalid string")
		}
		h = append(h, string(line))
	}
	return prompt.OptionHistory(h)(g.prompt)
}

func (g *GoPromptTerminal) AppendHistory(item string) {
	return
}

// TODO go-prompt has no interface to write history.
func (g *GoPromptTerminal) WriteHistory(w io.Writer) error {
	return nil
}

func (g *GoPromptTerminal) Close() {
	return
}

func (g *GoPromptTerminal) Prompt(p string) (string, error) {
	prompt.OptionPrefix(p)(g.prompt)
	s := g.prompt.Input()
	return s, nil
}
