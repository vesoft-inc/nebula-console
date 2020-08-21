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

var o = &printer.Outcsv{}

func welcome(interactive bool) {
	if !interactive {
		return
	}
	fmt.Println()
	fmt.Printf("Welcome to Nebula Graph %s!\n", Version)
	fmt.Println()
}

func bye(username string, interactive bool) {
	defer o.UnsetOutFile()
	if !interactive {
		return
	}
	fmt.Printf("Bye %s!\n", username)
	fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
	fmt.Println()
}

// client side cmd, will not be sent to server
func clientCmd(cmd string) (bool, bool) {
	plain := strings.Fields(strings.ToLower(cmd))
	if len(plain) == 1 && (plain[0] == "exit" || plain[0] == "quit") {
		return true, true
	} else if len(plain) == 3 && (plain[0] == "set" && plain[1] == "outfile") {
		o.SetOutFile(plain[2])
		return false, true
	} else if len(plain) == 2 && (plain[0] == "unset" && plain[1] == "outfile") {
		o.UnsetOutFile()
		return false, true
	}
	return false, false
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
		// Client Side command
		if exit, isLocal := clientCmd(line); isLocal {
			if exit {
				// Quit
				return nil
			} else {
				continue
			}
		}
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
