package main

import (
	"fmt"
	"os"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/convert"
	"github.com/britojr/exp-run/cmd/fstats"
	"github.com/britojr/exp-run/cmd/pmlearn"
	"github.com/britojr/exp-run/cmd/qevgen"
)

var commandMap = map[string]*cmd.Command{
	convert.Cmd.Name: convert.Cmd,
	qevgen.Cmd.Name:  qevgen.Cmd,
	fstats.Cmd.Name:  fstats.Cmd,
	pmlearn.Cmd.Name: pmlearn.Cmd,
}

func printDefaultsExit() {
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\t%s <command> [options]\n\n", os.Args[0])
	fmt.Printf("Commands:\n\n")
	for name, cm := range commandMap {
		fmt.Printf("\t%v\t\t%v\n", name, cm.Short)
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
