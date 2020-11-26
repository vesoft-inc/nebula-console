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
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-console/cli"
	"github.com/vesoft-inc/nebula-console/printer"
	ngdb "github.com/vesoft-inc/nebula-go/v2"
	graph "github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

// Nebula Console version
const (
	Version = "v2.0.0-alpha"
)

// Console side commands
const (
	Unknown  = -1
	Quit     = 0
	SetCsv   = 1
	UnsetCsv = 2
	PlayData = 3
	Sleep    = 4
	SetDot   = 5
	UnsetDot = 6
)

var dataSetPrinter = printer.NewDataSetPrinter()

var planDescPrinter = printer.NewPlanDescPrinter()

var datasets = map[string]string{
	"nba": "./data/nba.ngql",
}

func welcome(interactive bool) {
	defer dataSetPrinter.UnsetOutCsv()
	defer planDescPrinter.UnsetOutDot()
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

func printConsoleResp(msg string) {
	fmt.Println(msg)
	fmt.Println()
	fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
	fmt.Println()
}

func playData(client *ngdb.GraphClient, data string) (string, error) {
	path, exist := datasets[data]
	if !exist {
		return "", fmt.Errorf("dataset %s, not existed", data)
	}
	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	c := cli.NewnCli(fd, false, "", func() { fd.Close() })
	fmt.Printf("Start loading dataset %s...\n", data)
	err = loop(client, c, true)
	if err != nil {
		return "", err
	}
	respErr := c.GetRespError()
	if respErr != "" {
		return "", fmt.Errorf(respErr)
	}
	return c.GetSpace(), nil
}

// Console side cmd will not be sent to server
func isConsoleCmd(client *ngdb.GraphClient, cmd string) (isLocal bool, localCmd int, args []string) {
	// Currently, command "exit" and  "quit" can also exit the console
	if cmd == "exit" || cmd == "quit" {
		isLocal = true
		localCmd = Quit
		return
	}

	plain := strings.TrimSpace(strings.ToLower(cmd))
	if len(plain) < 1 || plain[0] != ':' {
		return
	}

	isLocal = true
	words := strings.Fields(plain[1:])
	switch len(words) {
	case 1:
		if words[0] == "exit" || words[0] == "quit" {
			localCmd = Quit
		} else {
			localCmd = Unknown
		}
	case 2:
		if words[0] == "unset" && words[1] == "csv" {
			localCmd = UnsetCsv
		} else if words[0] == "unset" && words[1] == "dot" {
			localCmd = UnsetDot
		} else if words[0] == "sleep" {
			localCmd = Sleep
			args = []string{words[1]}
		} else if words[0] == "play" {
			localCmd = PlayData
			args = []string{words[1]}
		} else {
			localCmd = Unknown
		}
	case 3:
		if words[0] == "set" && words[1] == "csv" {
			localCmd = SetCsv
			args = []string{words[2]}
		} else if words[0] == "set" && words[1] == "dot" {
			localCmd = SetDot
			args = []string{words[2]}
		} else {
			localCmd = Unknown
		}
	default:
		localCmd = Unknown
	}

	return
}

func executeConsoleCmd(client *ngdb.GraphClient, cmd int, args []string) (newSpace string) {
	switch cmd {
	case SetCsv:
		dataSetPrinter.SetOutCsv(args[0])
	case UnsetCsv:
		dataSetPrinter.UnsetOutCsv()
	case SetDot:
		planDescPrinter.SetOutDot(args[0])
	case UnsetDot:
		planDescPrinter.UnsetOutDot()
	case PlayData:
		var err error
		newSpace, err = playData(client, args[0])
		if err != nil {
			printConsoleResp("Error: load dataset failed, " + err.Error())
		} else {
			printConsoleResp("Load dataset succeeded!")
		}
	case Sleep:
		i, err := strconv.Atoi(args[0])
		if err != nil {
			printConsoleResp("Error: invalid integer, " + err.Error())
		}
		time.Sleep(time.Duration(i) * time.Second)
	default:
		printConsoleResp("Error: this local command not exists!")
	}
	return newSpace
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
		planDescPrinter.PrintPlanDesc(resp.GetPlanDesc())
	}
	fmt.Println()
}

func parseIP(address string) (string, error) {
	addrs, err := net.LookupHost(address)
	if err != nil {
		return "", err
	}
	// Return the first matched Ipv4 address
	for _, addr := range addrs {
		if net.ParseIP(addr).To4() != nil {
			return addr, nil
		}
	}
	return "", fmt.Errorf("No matching IPv4 address was found")
}

// Loop the request util fatal or timeout
// We treat one line as one query
// Add line break yourself as `SHOW \<CR>HOSTS`
func loop(client *ngdb.GraphClient, c cli.Cli, isPlayingData bool) error {
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
		// Console side command
		if isLocal, cmd, args := isConsoleCmd(client, line); isLocal {
			if cmd == Quit {
				return nil
			}
			newSpace := executeConsoleCmd(client, cmd, args)
			if newSpace != "" {
				c.SetSpace(newSpace)
			}
			continue
		}
		// Server side command
		start := time.Now()
		resp, err := client.Execute(line)
		if err != nil {
			return err
		}
		if ngdb.IsError(resp) {
			c.SetRespError(fmt.Sprintf("[ERROR (%d)]: %s", resp.GetErrorCode(), resp.GetErrorMsg()))
			if isPlayingData {
				break
			}
		}
		duration := time.Since(start)
		if c.Output() {
			printResp(resp, duration)
			fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
			fmt.Println()
		}
		c.SetSpace(string(resp.SpaceName))
	}
	return nil
}

var address *string = flag.String("addr", "127.0.0.1", "The Nebula Graph IP/HOST address")
var port *int = flag.Int("P", -1, "The Nebula Graph Port")
var username *string = flag.String("u", "", "The Nebula Graph login user name")
var password *string = flag.String("p", "", "The Nebula Graph login password")
var timeout *int = flag.Int("t", 120, "The Nebula Graph client connection timeout in seconds")
var script *string = flag.String("e", "", "The nGQL directly")
var file *string = flag.String("f", "", "The nGQL script file name")

func init() {
	flag.StringVar(address, "address", "127.0.0.1", "The Nebula Graph IP/HOST address")
	flag.IntVar(port, "port", -1, "The Nebula Graph Port")
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
			log.Fatalf("Get executable failed: %s", err.Error())
		}
		historyHome = filepath.Dir(ex) // Set to executable folder
	}
	if *port == -1 {
		log.Fatalf("Error: argument port is missed!")
	}
	ip, err := parseIP(*address)
	if err != nil {
		log.Fatalf("Error: address is invalid, %s", err.Error())
	}
	// when the value of timeout is set to 0, connection will not timeout
	clientTimeout := ngdb.WithTimeout(time.Duration(*timeout) * time.Second)
	client, err := ngdb.NewClient(fmt.Sprintf("%s:%d", ip, *port), clientTimeout)
	if err != nil {
		log.Fatalf("Fail to create client, address: %s, port: %d, %s", ip, *port, err.Error())
	}

	if len(*username) == 0 || len(*password) == 0 {
		log.Fatalf("Error: username or password is empty!")
	}

	if err = client.Connect(*username, *password); err != nil {
		log.Fatalf("Fail to connect server, %s", err.Error())
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
		c = cli.NewnCli(strings.NewReader(*script), true, *username, nil)
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Fatalf("Open file %s failed, %s", *file, err.Error())
		}
		c = cli.NewnCli(fd, true, *username, func() { fd.Close() })
	}

	if c == nil {
		return
	}

	defer c.Close()
	err = loop(client, c, false)
	if err != nil {
		log.Fatalf("Loop error, %s", err.Error())
	}
}
