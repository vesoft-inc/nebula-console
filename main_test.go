/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package main

import "testing"

func TestIsConsoleCmdMissingRequiredArguments(t *testing.T) {
	commands := []string{
		":",
		":sleep",
		":play",
		":repeat",
		":csv",
		":dot",
		":profile",
		":explain",
	}

	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			isLocal, localCmd, args, err := isConsoleCmd(command)
			if !isLocal {
				t.Fatalf("expected %q to be recognized as a local command", command)
			}
			if err == nil {
				t.Fatalf("expected a missing-argument error for %q", command)
			}
			if localCmd != Unknown {
				t.Errorf("expected Unknown command on parse error, got %d", localCmd)
			}
			if len(args) != 0 {
				t.Errorf("expected no parsed arguments on error, got %v", args)
			}
		})
	}
}

func TestIsConsoleCmdRepeatWithArgument(t *testing.T) {
	isLocal, localCmd, args, err := isConsoleCmd(":repeat 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isLocal {
		t.Fatal("expected a local command")
	}
	if localCmd != Repeat {
		t.Fatalf("expected Repeat command, got %d", localCmd)
	}
	if len(args) != 1 || args[0] != "3" {
		t.Fatalf("unexpected arguments: %v", args)
	}
}

func TestIsConsoleCmdUnknownCommand(t *testing.T) {
	isLocal, localCmd, args, err := isConsoleCmd(":unknown")
	if err != nil {
		t.Fatalf("unexpected error for an unknown command: %v", err)
	}
	if !isLocal {
		t.Fatal("expected an unknown colon-prefixed command to remain a local command")
	}
	if localCmd != Unknown {
		t.Fatalf("expected Unknown command, got %d", localCmd)
	}
	if len(args) != 0 {
		t.Fatalf("expected no arguments, got %v", args)
	}
}

func TestIsConsoleCmdParamsWithoutArgument(t *testing.T) {
	isLocal, localCmd, args, err := isConsoleCmd(":params")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isLocal {
		t.Fatal("expected a local command")
	}
	if localCmd != Params {
		t.Fatalf("expected Params command, got %d", localCmd)
	}
	if len(args) != 1 || args[0] != ":params" {
		t.Fatalf("unexpected arguments: %v", args)
	}
}
