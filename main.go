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
    "bufio"

    "github.com/vesoft-inc/nebula-console/completer"
    "github.com/peterh/liner"
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

func checkJoined(input string) {
    runes := []rune(input)
    var backSlashFound = false
    var backSlashIndex int
    for i := len(runes)-1; i >= 0; i-- {
        if runes[i] == 92 { // '\'
            backSlashFound = true
            backSlashIndex = i
        } else if runes[i] == 32 { // ' '
        } else{
            break
        }
    }
    if inputer.joined {
        if backSlashFound {
            inputer.line += string(runes[:backSlashIndex])
        } else {
            inputer.line += string(runes)
            inputer.joined = false
        }
     } else {
         if backSlashFound {
             inputer.line = string(runes[:backSlashIndex])
             inputer.joined = true
         } else {
             inputer.line = string(runes)
         }
     }
}

var inputer struct {
    line string
    joined bool
}

var prompter struct {
    color int
    user string
    space string
    pLen int
}

func nebulaPrompt() string {
    //ttyColor := prompter.color + 31
    //prompter.color = (prompter.color + 1) % 6
    prompt := ""
    //prompt += fmt.Sprintf("\033[%v;1m", ttyColor)
    if inputer.joined {
        for i := 0; i < prompter.pLen-3;i++ {
            prompt += " "
        }
        prompt += "-> "
    } else {
        promptString := fmt.Sprintf("(%s@nebula) [%s]> ", prompter.user, prompter.space)
        prompter.pLen = len(promptString)
        prompt += promptString
    }
    //prompt += "\033[0m"
    return prompt
}

// Loop the request util fatal or timeout
// We treat one line as one query
// Add line break yourself as `SHOW \<CR>HOSTS`
func loop(client *ngdb.GraphClient, username, historyFile string) {
    c := liner.NewLiner()
    defer c.Close()
    c.SetCtrlCAborts(true)
    // Two tab styles are currently available:
    //
    // 1.TabCircular cycles through each completion item and displays it directly on
    // the prompt
    //
    // 2.TabPrints prints the list of completion items to the screen after a second
    // tab key is pressed. This behaves similar to GNU readline and BASH (which
    // uses readline)

    //c.SetTabCompletionStyle(liner.TabPrints)
    c.SetWordCompleter(completer.NewCompleter)
    if f, err := os.Open(historyFile); err == nil {
		c.ReadHistory(f)
		f.Close()
    }

    inputer.line = ""
    inputer.joined = false
    prompter.color = 0
    prompter.user = username
    prompter.space = "(none)"
    for {
        if input, err:= c.Prompt(nebulaPrompt()); err == nil {
            if len(input) > 0 {
                c.AppendHistory(input)
            }
            checkJoined(input)
            if inputer.joined {
                continue
            }
            if len(inputer.line) == 0 {
                continue
            }
            // Client Side command
            if clientCmd(inputer.line) {
                // Quit
                break
            }

		    start := time.Now()
            resp, err := client.Execute(inputer.line)
		    duration := time.Since(start)
		    if err != nil {
			    log.Fatalf("Execute error, %s", err.Error())
		    }
		    printResp(resp, duration)
		    fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
            prompter.space = "(none)"
            if len(string(resp.SpaceName)) > 0 {
                prompter.space = string(resp.SpaceName)
            }
        } else if err == cli.ErrAborted {
            log.Print("Ctrl+C aborted: ", err)
            break
        } else if err == cli.ErrEOF {
            log.Print("EOF: ", err)
            break
        } else {
            log.Print("err:", err)
            break
        }
    }
    if f, err := os.Create(historyFile);  err != nil {
        log.Print("Error writing history file: ", err)
	} else {
		c.WriteHistory(f)
		defer f.Close()
	}

}

func process(client *ngdb.GraphClient, c *bufio.Reader) {
	for {
		line, _, err := c.ReadLine()
		lineString := string(line)
		if err != nil {
			break
		}
		if len(line) == 0 {
			fmt.Println()
			continue
		}

		// Client side command
		if clientCmd(lineString) {
			// Quit
			break
		}

		start := time.Now()
		resp, err := client.Execute(lineString)
		duration := time.Since(start)
		if err != nil {
			log.Fatalf("Execute error, %s", err.Error())
		}
		printResp(resp, duration)
		fmt.Println(time.Now().In(time.Local).Format(time.RFC1123))
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
        loop(client, *username, historyFile)
	} else if *script != "" {
		process(client, bufio.NewReader(strings.NewReader(*script)))
	} else if *file != "" {
		fd, err := os.Open(*file)
		if err != nil {
			log.Fatalf("Open file %s failed, %s", *file, err.Error())
		}
		defer fd.Close()
		process(client, bufio.NewReader(fd))
	}
}
