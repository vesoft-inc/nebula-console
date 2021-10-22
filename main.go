/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-console/box"
	"github.com/vesoft-inc/nebula-console/cli"
	"github.com/vesoft-inc/nebula-console/printer"
	nebula "github.com/vesoft-inc/nebula-go/v2"
)

// Console side commands
const (
	Unknown   = -1
	Quit      = 0
	PlayData  = 1
	Sleep     = 2
	ExportCsv = 3
	ExportDot = 4
	Repeat    = 5
)

var dataSetPrinter = printer.NewDataSetPrinter()

var planDescPrinter = printer.NewPlanDescPrinter()

/* Every statement will be repeatedly executed `g_repeats` times,
in order to get the total and avearge execution time of the statement") */
var g_repeats = 1

func welcome(interactive bool) {
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
	boxfilePath := "/" + data + ".ngql"
	posixfilePath := "./data/" + data + ".ngql"
	var c cli.Cli
	// First find it in directory ./data/. If not found, then find it in the embeded box
	if fd, err := os.Open(posixfilePath); err == nil {
		c = cli.NewnCli(fd, false, "", func() { fd.Close() })
	} else if box.Has(boxfilePath) {
		fileStr := string(box.Get(boxfilePath))
		c = cli.NewnCli(strings.NewReader(fileStr), false, "", nil)
	} else {
		return "", fmt.Errorf("file %s.ngql not existed in embed box and file directory ./data/ ", data)
	}

	c.PlayingData(true)
	defer c.PlayingData(false)
	fmt.Printf("Start loading dataset %s...\n", data)
	err := loop(c)
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
	isLocal = false
	localCmd = Unknown
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
	if plain[len(plain)-1] == ';' {
		plain = plain[:len(plain)-1]
	}
	words := strings.Fields(plain[1:])
	localCmdName := words[0]
	switch localCmdName {
	case "exit", "quit":
		localCmd = Quit
	case "sleep":
		{
			localCmd = Sleep
			args = []string{words[1]}
		}
	case "play":
		{
			localCmd = PlayData
			args = []string{words[1]}
		}
	case "repeat":
		{
			localCmd = Repeat
			args = []string{words[1]}
		}
	case "csv":
		{
			localCmd = ExportCsv
			args = []string{words[1]}
		}
	case "dot":
		{
			localCmd = ExportDot
			args = []string{words[1]}
		}
	}
	return
}

func executeConsoleCmd(c cli.Cli, cmd int, args []string) {
	switch cmd {
	case ExportCsv:
		dataSetPrinter.ExportCsv(args[0])
	case ExportDot:
		planDescPrinter.ExportDot(args[0])
	case PlayData:
		newSpace, err := playData(args[0])
		if err != nil {
			printConsoleResp("Error: load dataset failed, " + err.Error())
		} else {
			printConsoleResp("Load dataset succeeded!")
			c.SetSpace(newSpace)
		}
	case Sleep:
		i, err := strconv.Atoi(args[0])
		if err != nil {
			printConsoleResp("Error: invalid integer, " + err.Error())
		}
		time.Sleep(time.Duration(i) * time.Second)
	case Repeat:
		i, err := strconv.Atoi(args[0])
		if err != nil {
			printConsoleResp("Error: invalid integer, " + err.Error())
		} else if i < 1 {
			printConsoleResp("Error: invald integer, repeats should be greater than 1")
		}
		g_repeats = i
	default:
		printConsoleResp("Error: this local command not exists!")
	}
}

func printResultSet(res *nebula.ResultSet, startTime time.Time) (duration time.Duration) {
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
		duration = time.Since(startTime)
		if numRows > 0 {
			fmt.Printf("Got %d rows (time spent %d/%d us)\n", numRows, res.GetLatency(), duration/1000)
		} else {
			fmt.Printf("Empty set (time spent %d/%d us)\n", res.GetLatency(), duration/1000)
		}
	} else {
		duration = time.Since(startTime)
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

	return
}

// Loop the request util fatal or timeout
// We treat one line as one query
// Add line break yourself as `SHOW \<CR>HOSTS`
func loop(c cli.Cli) error {
	for {
		line, exit, err := c.ReadLine()
		if err != nil {
			return err
		}
		if exit { // Ctrl+D
			fmt.Println()
			return nil
		}
		if len(line) == 0 { // 1). The line input is empty, or 2). user presses ctrlC so the input is truncated
			continue
		}
		// Console side command
		if isLocal, cmd, args := isConsoleCmd(line); isLocal {
			if cmd == Quit {
				return nil
			}
			executeConsoleCmd(c, cmd, args)
			continue
		}
		// Server side command
		var t1 int32 = 0
		var t2 int32 = 0
		for i := 0; i < g_repeats; i++ {
			start := time.Now()
			res, err := session.Execute(line)
			if err != nil {
				return err
			}
			if !res.IsSucceed() && !res.IsPartialSucceed() {
				c.SetRespError(fmt.Sprintf("an error occurred when executing: %s, [ERROR (%d)]: %s", line, res.GetErrorCode(), res.GetErrorMsg()))
				if c.IsPlayingData() {
					return nil
				}
			}
			t1 += res.GetLatency()
			if c.Output() {
				duration := printResultSet(res, start)
				t2 += int32(duration / 1000)
				fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
				fmt.Println()
			}
			c.SetSpace(res.GetSpaceName())
		}
		if g_repeats > 1 {
			fmt.Printf("Executed %v times, (total time spent %d/%d us), (average time spent %d/%d us)\n", g_repeats, t1, t2, t1/int32(g_repeats), t2/int32(g_repeats))
			fmt.Println()
		}
		g_repeats = 1
	}
}

// Nebula Console version related
var (
	gitCommit string
	buildDate string
)

var (
	address           *string = flag.String("addr", "127.0.0.1", "The Nebula Graph IP/HOST address")
	port              *int    = flag.Int("P", -1, "The Nebula Graph Port")
	username          *string = flag.String("u", "", "The Nebula Graph login user name")
	password          *string = flag.String("p", "", "The Nebula Graph login password")
	timeout           *int    = flag.Int("t", 0, "The Nebula Graph client connection timeout in seconds, 0 means never timeout")
	script            *string = flag.String("e", "", "The nGQL directly")
	file              *string = flag.String("f", "", "The nGQL script file name")
	version           *bool   = flag.Bool("v", false, "The Nebula Console version")
	enableSsl         *bool   = flag.Bool("enable_ssl", false, "Enable SSL when connecting to Nebula Graph")
	sslRootCAPath     *string = flag.String("ssl_root_ca_path", "", "SSL root certification authority's file path")
	sslCertPath       *string = flag.String("ssl_cert_path", "", "SSL certificate's file path")
	sslPrivateKeyPath *string = flag.String("ssl_private_key_path", "", "SSL private key's file path")
)

func init() {
	flag.StringVar(address, "address", "127.0.0.1", "The Nebula Graph IP/HOST address")
	flag.IntVar(port, "port", -1, "The Nebula Graph Port")
	flag.StringVar(username, "user", "", "The Nebula Graph login user name")
	flag.StringVar(password, "password", "", "The Nebula Graph login password")
	flag.IntVar(timeout, "timeout", 0, "The Nebula Graph client connection timeout in seconds, 0 means never timeout")
	flag.StringVar(script, "eval", "", "The nGQL directly")
	flag.StringVar(file, "file", "", "The nGQL script file name")
}

func validateFlags() {
	if *port == -1 {
		log.Panicf("Error: argument port is missed!")
	}
	if len(*username) == 0 {
		log.Panicf("Error: argument username is empty!")
	}
	if len(*password) == 0 {
		log.Panicf("Error: argument password is empty!")
	}

	if *enableSsl {
		if *sslRootCAPath == "" {
			log.Panicf("Error: argument ssl_root_ca_path should be specified when enable_ssl is true")
		}
		if *sslCertPath == "" {
			log.Panicf("Error: argument ssl_cert_path should be specified when enable_ssl is true")
		}
		if *sslPrivateKeyPath == "" {
			log.Panicf("Error: argument ssl_private_key_path should be specified when enable_ssl is true")
		}
	}
}

var pool *nebula.ConnectionPool

var session *nebula.Session

func openAndReadFile(path string) ([]byte, error) {
	// open file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s: %s", path, err)
	}
	// read file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to ReadAll of file %s: %s", path, err)
	}
	return b, nil
}

func genSslConfig(rootCAPath, certPath, privateKeyPath string) (*tls.Config, error) {
	rootCA, err := openAndReadFile(rootCAPath)
	if err != nil {
		return nil, err
	}
	// log.Println(rootCA)
	cert, err := openAndReadFile(certPath)
	if err != nil {
		return nil, err
	}
	// log.Println(cert)
	privateKey, err := openAndReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	// log.Println(privateKey)

	// generate the client certificate
	clientCert, err := tls.X509KeyPair(cert, privateKey)
	if err != nil {
		return nil, err
	}

	// parse root CA pem and add into CA pool
	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(rootCA)
	if !ok {
		return nil, fmt.Errorf("fail to append supplied cert into tls.Config, please make sure it is a valid certificate")
	}

	// set tls config
	// InsecureSkipVerify is set to true for test purpose ONLY. DO NOT use it in production.
	return &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            rootCAPool,
		InsecureSkipVerify: false,
	}, nil
}

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
	if *enableSsl {
		sslConfig, err2 := genSslConfig(*sslRootCAPath, *sslCertPath, *sslPrivateKeyPath)
		if err2 != nil {
			log.Panicf(fmt.Sprintf("Fail to generate the ssl config, ssl_root_ca_path: %s, ssl_cert_path: %s, ssl_private_key_path: %s, %s", *sslRootCAPath, *sslCertPath, *sslPrivateKeyPath, err2.Error()))
		}
		fmt.Println(sslConfig)
		pool, err = nebula.NewSslConnectionPool(hostList, poolConfig, sslConfig, nebula.DefaultLogger{})
	} else {
		pool, err = nebula.NewConnectionPool(hostList, poolConfig, nebula.DefaultLogger{})
	}
	if err != nil {
		log.Panicf(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", *address, *port, err.Error()))
	}
	defer pool.Close()

	session, err = pool.GetSession(*username, *password)
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

	err = loop(c)

	if err != nil {
		log.Panicf("Loop error, %s", err.Error())
	}
}
