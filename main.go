package main

import (
	"fmt"
	"os"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/calcdist"
	"github.com/britojr/exp-run/cmd/convert"
	"github.com/britojr/exp-run/cmd/fstats"
	"github.com/britojr/exp-run/cmd/inference"
	"github.com/britojr/exp-run/cmd/pmlearn"
	"github.com/britojr/exp-run/cmd/qevgen"
	"github.com/britojr/exp-run/cmd/sample"
)

var commands = []*cmd.Command{
	convert.Cmd,
	fstats.Cmd,
	qevgen.Cmd,
	pmlearn.Cmd,
	inference.Cmd,
	calcdist.Cmd,
	sample.Cmd,
}

var commandMap map[string]*cmd.Command

func init() {
	commandMap = make(map[string]*cmd.Command)
	for _, cm := range commands {
		commandMap[cm.Name] = cm
	}
}

func printDefaultsExit() {
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\t%s <command> [options]\n\n", os.Args[0])
	fmt.Printf("Commands:\n\n")
	for _, cm := range commands {
		fmt.Printf("\t%v\t\t%v\n", cm.Name, cm.Short)
	}
	fmt.Println()
	fmt.Printf("For usage details of each command, run:\n\n")
	fmt.Printf("\t%s <command> --help\n", os.Args[0])
	fmt.Println()
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		printDefaultsExit()
	}
	if cm, ok := commandMap[os.Args[1]]; ok {
		cm.Run(cm, os.Args[2:])
	} else {
		printDefaultsExit()
	}
}
