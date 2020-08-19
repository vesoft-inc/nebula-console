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
    "path"

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
func loop(client *ngdb.GraphClient, c cli.Cli) {
    for {
        line, err, exit:= c.ReadLine()
        if exit {
            return
        }
        if err == nil {
            if len(line) == 0 {
                continue
            }
            // Client Side command
            if clientCmd(line) {
                // Quit
                break
            }
            start := time.Now()
            resp, err := client.Execute(line)
            duration := time.Since(start)
            if err != nil {
                log.Fatalf("Execute error, %s", err.Error())
            }
            printResp(resp, duration)
            fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
            c.SetSpace("(none)")
            if len(string(resp.SpaceName)) > 0 {
                c.SetSpace(string(resp.SpaceName))
            }
        } else {
            log.Print("err:", err)
            break
        }
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
    historyFile := path.Join(historyHome, ".nebula_history")
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
	if interactive {
        c := cli.NewiCli(historyFile, *username)
        defer c.Terminal.Close()
        loop(client, c)
        if f, err := os.Create(historyFile); err != nil {
            log.Print("error writing history file: ", err)
	    } else {
		    c.Terminal.WriteHistory(f)
	        defer f.Close()
        }
    } else if *script != "" {
		loop(client, cli.NewnCli(strings.NewReader(*script)))
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Fatalf("Open file %s failed, %s", *file, err.Error())
		}
		defer fd.Close()
	    loop(client, cli.NewnCli(fd))
	}
}
