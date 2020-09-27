/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-console/cli"
	"github.com/vesoft-inc/nebula-console/printer"
	ngdb "github.com/vesoft-inc/nebula-go/v2"
	graph "github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

const (
	Version = "v2.0.0-alpha"
)

var dataSetPrinter = printer.NewDataSetPrinter()

func welcome(interactive bool) {
	defer dataSetPrinter.UnsetOutCsv()
	if !interactive {
		return
	}
	fmt.Println()
	fmt.Printf("Welcome to Nebula Graph %s!\n", Version)
	fmt.Println()
}

func bye(username string, interactive bool) {
	fmt.Println()
	fmt.Printf("Bye %s!\n", username)
	fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
	fmt.Println()
}

// client side cmd, will not be sent to server
func clientCmd(cmd string) (isLocal, exit bool) {
	// currently, command "exit" and  "quit" can also exit the console
	if cmd == "exit" || cmd == "quit" {
		isLocal = true
		exit = true
		return
	}

	plain := strings.TrimSpace(strings.ToLower(cmd))
	if len(plain) < 1 || plain[0] != ':' {
		return
	}

	isLocal = true
	runes := strings.Fields(plain[1:])
	if len(runes) == 1 && (runes[0] == "exit" || runes[0] == "quit") {
		exit = true
	} else if len(runes) == 3 && (runes[0] == "set" && runes[1] == "csv") {
		dataSetPrinter.SetOutCsv(runes[2])
		exit = false
	} else if len(runes) == 2 && (runes[0] == "unset" && runes[1] == "csv") {
		dataSetPrinter.UnsetOutCsv()
		exit = false
	} else {
		exit = false
		fmt.Println("Error: this local command not exists!")
		fmt.Println()
		fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
		fmt.Println()
	}

	return
}

func printResp(resp *graph.ExecutionResponse, duration time.Duration) {
	if ngdb.IsError(resp) {
		fmt.Printf("[ERROR (%d)]: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		fmt.Println()
		fmt.Println()
		return
	}
	// Show table
	if resp.IsSetData() {
		dataSetPrinter.PrintDataSet(resp.GetData())
		numRows := len(resp.GetData().GetRows())
		if numRows > 0 {
			fmt.Printf("Got %d rows (time spent %d/%d us)\n", numRows, resp.GetLatencyInUs(), duration/1000)
		} else {
			fmt.Printf("Empty set (time spent %d/%d us)\n", resp.GetLatencyInUs(), duration/1000)
		}
	} else {
		fmt.Printf("Execution succeeded (time spent %d/%d us)\n", resp.GetLatencyInUs(), duration/1000)
	}

	if resp.IsSetPlanDesc() {
		fmt.Println()
		fmt.Println("Execution Plan")
		fmt.Println()
		p := printer.NewPlanDescPrinter(resp.GetPlanDesc())
		fmt.Println(p.Print())
	}
	fmt.Println()
}

// Loop the request util fatal or timeout
// We treat one line as one query
// Add line break yourself as `SHOW \<CR>HOSTS`
func loop(client *ngdb.GraphClient, c cli.Cli) error {
	for {
		line, exit, err := c.ReadLine()
		if err != nil {
			return err
		}
		if exit { // Ctrl+D
			fmt.Println()
			return nil
		}
		if len(line) == 0 {
			continue
		}
		// Client side command
		if isLocal, quit := clientCmd(line); isLocal {
			if quit { // :exit, :quit, exit, quit
				return nil
			} else {
				continue
			}
		}
		// Server side command
		start := time.Now()
		resp, err := client.Execute(line)
		if err != nil {
			return err
		}

		duration := time.Since(start)
		printResp(resp, duration)
		fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
		fmt.Println()
		c.SetSpace(string(resp.SpaceName))
	}
}

var address *string = flag.String("addr", "127.0.0.1", "The Nebula Graph IP address")
var port *int = flag.Int("port", 3699, "The Nebula Graph Port")
var username *string = flag.String("u", "", "The Nebula Graph login user name")
var password *string = flag.String("p", "", "The Nebula Graph login password")
var timeout *int = flag.Int("t", 120, "The Nebula Graph client connection timeout in seconds")
var script *string = flag.String("e", "", "The nGQL directly")
var file *string = flag.String("f", "", "The nGQL script file name")

func init() {
	flag.StringVar(address, "address", "127.0.0.1", "The Nebula Graph IP address")
	flag.StringVar(username, "user", "", "The Nebula Graph login user name")
	flag.StringVar(password, "password", "", "The Nebula Graph login password")
	flag.IntVar(timeout, "timeout", 120, "The Nebula Graph client connection timeout in seconds")
	flag.StringVar(script, "eval", "", "The nGQL directly")
	flag.StringVar(file, "file", "", "The nGQL script file name")
}

func main() {
	flag.Parse()

	interactive := *script == "" && *file == ""

	historyHome := os.Getenv("HOME")
	if historyHome == "" {
		ex, err := os.Executable()
		if err != nil {
			log.Panicf("Get executable failed: %s", err.Error())
		}
		historyHome = filepath.Dir(ex) // Set to executable folder
	}
	// when the value of timeout is set to 0, connection will not timeout
	clientTimeout := ngdb.WithTimeout(time.Duration(*timeout) * time.Second)
	client, err := ngdb.NewClient(fmt.Sprintf("%s:%d", *address, *port), clientTimeout)
	if err != nil {
		log.Panicf("Fail to create client, address: %s, port: %d, %s", *address, *port, err.Error())
	}

	if len(*username) == 0 || len(*password) == 0 {
		log.Panicf("Error: username or password is empty!")
	}

	if err = client.Connect(*username, *password); err != nil {
		log.Panicf("Fail to connect server, ", err.Error())
	}

	welcome(interactive)

	defer bye(*username, interactive)
	defer client.Disconnect()

	var c cli.Cli = nil
	// Loop the request
	if interactive {
		historyFile := path.Join(historyHome, ".nebula_history")
		c = cli.NewiCli(historyFile, *username)
	} else if *script != "" {
		c = cli.NewnCli(strings.NewReader(*script), *username, nil)
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Panicf("Open file %s failed, %s", *file, err.Error())
		}
		c = cli.NewnCli(fd, *username, func() { fd.Close() })
	}

	if c == nil {
		return
	}

	defer c.Close()
	err = loop(client, c)
	if err != nil {
		log.Panicf("Loop error, %s", err.Error())
	}
}
