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

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/vesoft-inc/nebula-console/cli"
	"github.com/vesoft-inc/nebula-console/printer"
	ngdb "github.com/vesoft-inc/nebula-go/v2"
	graph "github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

const (
	Version = "v2.0.0-alpha"
)

var o = &printer.OutCsv{}

func welcome(interactive bool) {
	if !interactive {
		return
	}
	fmt.Println()
	fmt.Printf("Welcome to Nebula Graph %s!\n", Version)
	fmt.Println()
}

func bye(username string, interactive bool) {
	defer o.UnsetOutCsv()
	if !interactive {
		return
	}
	fmt.Printf("Bye %s!\n", username)
	fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
	fmt.Println()
}

// client side cmd, will not be sent to server
func clientCmd(cmd string) (isLocal, exit bool, err error) {
	plain := strings.TrimSpace(strings.ToLower(cmd))
	if len(plain) < 1 || plain[0] != ':' {
		return
	}

	isLocal = true
	runes := strings.Fields(plain[1:])
	if len(runes) == 1 && (runes[0] == "exit" || runes[0] == "quit") {
		exit = true
	} else if len(runes) == 3 && (runes[0] == "set" && runes[1] == "outcsv") {
		o.SetOutCsv(runes[2])
		exit = false
	} else if len(runes) == 2 && (runes[0] == "unset" && runes[1] == "outcsv") {
		o.UnsetOutCsv()
		exit = false
	} else {
		exit = false
		err = fmt.Errorf("Error: this local command not exists!")
	}

	return
}

func printResp(resp *graph.ExecutionResponse, duration time.Duration) {
	// Error
	if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
		fmt.Printf("[ERROR (%d)]: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		fmt.Println()
		fmt.Println()
		return
	}
	// Show table
	if resp.IsSetData() {
		printer.PrintDataSet(resp.GetData(), o)
		if len(resp.GetData().GetRows()) > 0 {
			fmt.Printf("Got %d rows (time spent %d/%d us)\n",
				len(resp.GetData().GetRows()), resp.GetLatencyInUs(), duration/1000)
		} else {
			fmt.Printf("Empty set (time spent %d/%d us)\n", resp.GetLatencyInUs(), duration/1000)
		}
	} else {
		fmt.Printf("Execution succeeded (time spent %d/%d us)\n", resp.GetLatencyInUs(), duration/1000)
	}

	if resp.IsSetPlanDesc() {
		fmt.Println()
		fmt.Println(text.Bold.Sprint("Execution Plan"))
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
		line, err, exit := c.ReadLine()
		if exit { // Ctrl+D
			return nil
		}
		if err != nil {
			return err
		}
		if len(line) == 0 {
			continue
		}
		isLocal, exit, err := clientCmd(line)
		// Client side command
		if isLocal {
			if exit {
				// Quit
				return nil
			} else if err != nil {
				return err
			} else {
				continue
			}
		}
		// Server side command
		start := time.Now()
		resp, err := client.Execute(line)
		duration := time.Since(start)
		if err != nil {
			return err
		}
		printResp(resp, duration)
		fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
		fmt.Println()
		c.SetSpace(string(resp.SpaceName))
	}

	return nil
}

func main() {
	address := flag.String("address", "127.0.0.1", "The Nebula Graph IP address")
	port := flag.Int("port", 3699, "The Nebula Graph Port")
	username := flag.String("u", "user", "The Nebula Graph login user name")
	password := flag.String("p", "password", "The Nebula Graph login password")
	script := flag.String("e", "", "The nGQL directly")
	file := flag.String("f", "", "The nGQL script file name")
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
	historyFile := path.Join(historyHome, ".nebula_history")
	client, err := ngdb.NewClient(fmt.Sprintf("%s:%d", *address, *port))
	if err != nil {
		log.Panicf("Fail to create client, address: %s, port: %d, %s", *address, *port, err.Error())
	}

	if err = client.Connect(*username, *password); err != nil {
		log.Panicf("Fail to connect server, username: %s, password: %s, %s", *username, *password, err.Error())
	}

	welcome(interactive)

	defer bye(*username, interactive)
	defer client.Disconnect()

	// Loop the request
	if interactive {
		c := cli.NewiCli(historyFile, *username)
		defer c.Close()
		err = loop(client, c)
	} else if *script != "" {
		err = loop(client, cli.NewnCli(strings.NewReader(*script)))
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Panicf("Open file %s failed, %s", *file, err.Error())
		}
		defer fd.Close()
		err = loop(client, cli.NewnCli(fd))
	}

	if err != nil {
		log.Panicf("Loop error, %s", err.Error())
	}
}
