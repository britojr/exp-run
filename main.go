package main

import (
	"fmt"
	"os"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/convert"
)

var commandMap = map[string]*cmd.Command{
	convert.CmdName: convert.CmdConvert,
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
