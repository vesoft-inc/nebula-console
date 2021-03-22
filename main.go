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
	"strconv"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-console/cli"
	"github.com/vesoft-inc/nebula-console/printer"
	nebula "github.com/vesoft-inc/nebula-go"
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
	fmt.Printf("Welcome to Nebula Graph!\n")
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

func playData(data string) (string, error) {
	path, exist := datasets[data]
	if !exist {
		return "", fmt.Errorf("dataset %s, not existed", data)
	}
	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	c := cli.NewnCli(fd, false, "", func() { fd.Close() })
	c.PlayingData(true)
	defer c.PlayingData(false)
	fmt.Printf("Start loading dataset %s...\n", data)
	childSession, err := pool.GetSession(*username, *password)
	if err != nil {
		log.Panicf("Fail to create a new session from connection pool, %s", err.Error())
	}
	defer childSession.Release()
	err = loop(childSession, c)
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
func isConsoleCmd(cmd string) (isLocal bool, localCmd int, args []string) {
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

func executeConsoleCmd(cmd int, args []string) (newSpace string) {
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
		newSpace, err = playData(args[0])
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

func printResultSet(res *nebula.ResultSet, duration time.Duration) {
	if !res.IsSucceed() && !res.IsPartialSucceed() {
		fmt.Printf("[ERROR (%d)]: %s", res.GetErrorCode(), res.GetErrorMsg())
		fmt.Println()
		fmt.Println()
		return
	}
	// Show table
	if res.IsSetData() {
		dataSetPrinter.PrintDataSet(res)
		numRows := res.GetRowSize()
		if numRows > 0 {
			fmt.Printf("Got %d rows (time spent %d/%d us)\n", numRows, res.GetLatency(), duration/1000)
		} else {
			fmt.Printf("Empty set (time spent %d/%d us)\n", res.GetLatency(), duration/1000)
		}
	} else {
		fmt.Printf("Execution succeeded (time spent %d/%d us)\n", res.GetLatency(), duration/1000)
	}

	if res.IsPartialSucceed() {
		fmt.Println()
		fmt.Printf("[WARNING]: Got partial result.")
	}

	if res.IsSetComment() {
		fmt.Println()
		fmt.Printf("[WARNING]: %s", res.GetComment())
	}

	if res.IsSetPlanDesc() {
		fmt.Println()
		fmt.Printf("Execution Plan (optimize time %d us)\n", res.GetPlanDesc().GetOptimizeTimeInUs())
		fmt.Println()
		planDescPrinter.PrintPlanDesc(res)
	}
	fmt.Println()
}

// Loop the request util fatal or timeout
// We treat one line as one query
// Add line break yourself as `SHOW \<CR>HOSTS`
func loop(session *nebula.Session, c cli.Cli) error {
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
		if isLocal, cmd, args := isConsoleCmd(line); isLocal {
			if cmd == Quit {
				return nil
			}
			newSpace := executeConsoleCmd(cmd, args)
			if newSpace != "" {
				c.SetSpace(newSpace)
				session.Execute(fmt.Sprintf("USE %s", newSpace))
				if err != nil {
					return err
				}
			}
			continue
		}
		// Server side command
		start := time.Now()
		res, err := session.Execute(line)
		if err != nil {
			return err
		}
		if !res.IsSucceed() && !res.IsPartialSucceed() {
			c.SetRespError(fmt.Sprintf("[ERROR (%d)]: %s", res.GetErrorCode(), res.GetErrorMsg()))
			if c.IsPlayingData() {
				break
			}
		}
		duration := time.Since(start)
		if c.Output() {
			printResultSet(res, duration)
			fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
			fmt.Println()
		}
		c.SetSpace(res.GetSpaceName())
	}
	return nil
}

// Nebula Console version related
var (
	gitCommit string
	buildDate string
)

var (
	address  *string = flag.String("addr", "127.0.0.1", "The Nebula Graph IP/HOST address")
	port     *int    = flag.Int("P", -1, "The Nebula Graph Port")
	username *string = flag.String("u", "", "The Nebula Graph login user name")
	password *string = flag.String("p", "", "The Nebula Graph login password")
	timeout  *int    = flag.Int("t", 0, "The Nebula Graph client connection timeout in seconds, 0 means never timeout")
	script   *string = flag.String("e", "", "The nGQL directly")
	file     *string = flag.String("f", "", "The nGQL script file name")
	version  *bool   = flag.Bool("v", false, "The Nebula Console version")
)

func init() {
	flag.StringVar(address, "address", "127.0.0.1", "The Nebula Graph IP/HOST address")
	flag.IntVar(port, "port", -1, "The Nebula Graph Port")
	flag.StringVar(username, "user", "", "The Nebula Graph login user name")
	flag.StringVar(password, "password", "", "The Nebula Graph login password")
	flag.IntVar(timeout, "timeout", 0, "The Nebula Graph client connection timeout in seconds, 0 means never timeout")
	flag.StringVar(script, "eval", "", "The nGQL directly")
	flag.StringVar(file, "file", "", "The nGQL script file name")
	flag.BoolVar(version, "version", false, "The Nebula Console version")
}

func validateFlags() {
	if *port == -1 {
		log.Panicf("Error: argument port is missed!")
	}
	if len(*username) == 0 {
		log.Panicf("Error: username is empty!")
	}
	if len(*password) == 0 {
		log.Panicf("Error: password is empty!")
	}
}

var pool *nebula.ConnectionPool

func main() {
	flag.Parse()

	if flag.NFlag() == 1 && *version {
		fmt.Printf("nebula-console version Git: %s, Build Time: %s\n", gitCommit, buildDate)
		return
	}

	// Check if flags are valid
	validateFlags()

	interactive := *script == "" && *file == ""

	historyHome := os.Getenv("HOME")
	if historyHome == "" {
		ex, err := os.Executable()
		if err != nil {
			log.Panicf("Get executable failed: %s", err.Error())
		}
		historyHome = filepath.Dir(ex) // Set to executable folder
	}

	hostAddress := nebula.HostAddress{Host: *address, Port: *port}
	hostList := []nebula.HostAddress{hostAddress}
	poolConfig := nebula.PoolConfig{
		TimeOut:         time.Duration(*timeout) * time.Millisecond,
		IdleTime:        0 * time.Millisecond,
		MaxConnPoolSize: 2,
		MinConnPoolSize: 0,
	}
	var err error
	pool, err = nebula.NewConnectionPool(hostList, poolConfig, nebula.DefaultLogger{})
	if err != nil {
		log.Panicf(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", *address, *port, err.Error()))
	}
	defer pool.Close()

	session, err := pool.GetSession(*username, *password)
	if err != nil {
		log.Panicf("Fail to create a new session from connection pool, %s", err.Error())
	}
	defer session.Release()

	welcome(interactive)
	defer bye(*username, interactive)

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
			log.Panicf("Open file %s failed, %s", *file, err.Error())
		}
		c = cli.NewnCli(fd, true, *username, func() { fd.Close() })
	}

	if c == nil {
		return
	}

	defer c.Close()
	err = loop(session, c)
	if err != nil {
		log.Panicf("Loop error, %s", err.Error())
	}
}
