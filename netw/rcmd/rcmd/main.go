package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Centny/gwf/log"

	"github.com/Centny/gwf/util"

	"github.com/Centny/gwf/netw/rcmd"
)

const (
	DefaultAddr = ":2984"
)

func main() {
	if len(os.Args) < 2 {
		runControl("127.0.0.1"+DefaultAddr, "Ctrl-local", "local")
		return
	}
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-m":
			runMaster(DefaultAddr, `{"Slave-abc":1,"Ctrl-abc":1}`)
		case "-c":
			runControl("127.0.0.1"+DefaultAddr, "Ctrl-local", "local")
		default:
			printUsage(1)
		}
	}
	switch os.Args[1] {
	case "-m":
		runMaster(os.Args[2:]...)
	case "-c":
		runControl(os.Args[2:]...)
	case "-s":
		runSlave(os.Args[2:]...)
	case "-h":
		printUsage(0)
	default:
		printUsage(1)
	}
}

func runControl(args ...string) {
	if len(args) < 2 {
		printUsage(1)
		return
	}
	logpath := "/tmp/rc_control.log"
	logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	log.SetWriter(bufio.NewWriter(logfile))
	rcaddr, token := args[0], args[1]
	alias, _ := os.Hostname()
	if len(args) > 2 {
		alias = args[2]
	}
	if len(alias) < 1 {
		alias = "control"
	}
	fmt.Printf("run control by rcaddr(%v),token(%v),alias(%v)\n", rcaddr, token, alias)
	err = rcmd.StartControl(alias, rcaddr, token)
	if err != nil {
		panic(err)
	}
	// rcmd.SharedControl.Wait()
	// stdin := bufio.NewReader(os.Stdin)
	for {
		baseline, err := String("> ")
		if err == io.EOF {
			break
		}
		line := strings.TrimSpace(baseline)
		if strings.HasPrefix(line, "ls") {
			res, err := rcmd.SharedControl.List()
			if err == nil {
				fmt.Println(util.S2Json(res))
			} else {
				fmt.Printf("ls cmd fail with %v", err)
			}
			AddHistory(baseline)
			continue
		}
		if strings.HasPrefix(line, "start") {
			line = strings.TrimPrefix(line, "start ")
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, ">", 2)
			cmds := parts[0]
			logfile := ""
			if len(parts) > 1 {
				logfile = parts[1]
			}
			res, err := rcmd.SharedControl.StartCmd(cmds, logfile)
			if err == nil {
				fmt.Println(util.S2Json(res))
			} else {
				fmt.Printf("start cmd fail with %v", err)
			}
			AddHistory(baseline)
			continue
		}
		if strings.HasPrefix(line, "stop") {
			line = strings.TrimPrefix(line, "stop ")
			line = strings.TrimSpace(line)
			line = regexp.MustCompile("[ ]+").ReplaceAllString(line, " ")
			parts := strings.SplitN(line, " ", 2)
			var res util.Map
			switch len(parts) {
			case 1:
				res, err = rcmd.SharedControl.StopCmd("", parts[0])
			default:
				res, err = rcmd.SharedControl.StopCmd(parts[0], parts[1])
			}
			if err == nil {
				fmt.Println(util.S2Json(res))
			} else {
				fmt.Printf("stop cmd fail with %v", err)
			}
			AddHistory(baseline)
			continue
		}
		if strings.HasPrefix(line, "help") {
			printCtrlUsage()
			continue
		}
		if strings.HasPrefix(line, "exit") {
			break
		}
		fmt.Println("unknow:", line)
	}
	rcmd.StopControl()
}

func runMaster(args ...string) {
	if len(args) < 2 {
		printUsage(1)
		return
	}
	rcaddr, tokens := args[0], args[1]
	fmt.Printf("run master by rcaddr(%v),tokens(%v)\n", rcaddr, tokens)
	var ts = map[string]int{}
	util.Json2S(tokens, &ts)
	ts["Ctrl-local"] = 1
	err := rcmd.StartMaster(rcaddr, ts)
	if err != nil {
		panic(err)
	}
	rcmd.SharedMaster.Wait()
}

func runSlave(args ...string) {
	if len(args) < 2 {
		printUsage(1)
		return
	}
	rcaddr, token := args[0], args[1]
	alias, _ := os.Hostname()
	if len(args) > 2 {
		alias = args[2]
	}
	if len(alias) < 1 {
		panic("the slave alias is empty")
	}
	fmt.Printf("run slave by rcaddr(%v),token(%v),alias(%v)\n", rcaddr, token, alias)
	err := rcmd.StartSlave(alias, rcaddr, token)
	if err != nil {
		panic(err)
	}
	// rcmd.SharedSlave.Wait()
	wait := make(chan int)
	<-wait
}

func printUsage(exit int) {
	_, name := filepath.Split(os.Args[0])
	fmt.Printf(`Usage:
	%v -m [<listen addr> <token config>]	run as master
	%v -s <master addr> <token> [<alias>]	run as slave
	%v -c <master addr> <token> [<alias>]	run as control
	%v	run as local control%v`,
		name, name, name, name, "\n")
	os.Exit(exit)
}

func printCtrlUsage() {
	fmt.Println(`
	ls		=>list the running task
	start <command and args> > log file => start command
	stop <cid> <tid> => stop command on special client
	stop <tid> => stop command on all client
	help	=> show this`)
}