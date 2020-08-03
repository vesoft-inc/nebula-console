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

func welcome(interactive bool) {
	if !interactive {
		return
	}
	fmt.Println()
	fmt.Printf("Welcome to Nebula Graph %s!\n", Version)
	fmt.Println()
}

func bye(username string, interactive bool) {
	if !interactive {
		return
	}
	fmt.Printf("Bye %s!", username)
	fmt.Println()
}

// return , does exit
func clientCmd(query string) bool {
	plain := strings.ToLower(strings.TrimSpace(query))
	if plain == "exit" || plain == "quit" {
		return true
	}
	return false
}

func printResp(resp *graph.ExecutionResponse, duration time.Duration) {
	// Error
	if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
		fmt.Printf("[ERROR (%d)]: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		fmt.Println()
		return
	}
	// Show table
	if resp.IsSetData() {
		printer.PrintDataSet(resp.GetData())
		// Show time
		fmt.Printf("time spent %d/%d us\n", resp.GetLatencyInUs(), duration/1000)
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
		lineString := string(line)
		if exit {
			return err
		}
		if len(line) == 0 {
			fmt.Println()
			continue
		}

		// Client side command
		if clientCmd(lineString) {
			// Quit
			return nil
		}

		start := time.Now()
		resp, err := client.Execute(lineString)
		duration := time.Since(start)
		if err != nil {
			log.Fatalf("Execute error, %s", err.Error())
		}
		printResp(resp, duration)
		fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
		c.SetSpace(string(resp.SpaceName))
		c.SetisErr(resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED)
		fmt.Println()
	}
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
			log.Fatalf("Get executable failed: %s", err.Error())
		}
		historyHome = filepath.Dir(ex) // Set to executable folder
	}

	client, err := ngdb.NewClient(fmt.Sprintf("%s:%d", *address, *port))
	if err != nil {
		log.Fatalf("Fail to create client, address: %s, port: %d, %s", *address, *port, err.Error())
	}

	if err = client.Connect(*username, *password); err != nil {
		log.Fatalf("Fail to connect server, username: %s, password: %s, %s", *username, *password, err.Error())
	}

	welcome(interactive)

	defer bye(*username, interactive)
	defer client.Disconnect()

	// Loop the request
	var exit error = nil
	if interactive {
		exit = loop(client, cli.NewiCli(historyHome, *username))
	} else if *script != "" {
		exit = loop(client, cli.NewnCli(strings.NewReader(*script)))
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Fatalf("Open file %s failed, %s", *file, err.Error())
		}
		defer fd.Close()
		exit = loop(client, cli.NewnCli(fd))
	}

	if exit != nil {
		os.Exit(1)
	}
}
