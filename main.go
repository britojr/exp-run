package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/britojr/exp-run/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printDefaultsExit()
	}
	if sc, ok := subCommandMap[os.Args[1]]; ok {
		sc.Run()
	} else {
		printDefaultsExit()
	}
}

const (
	ctConvert = "convert"
)

var subCommandMap map[string]subCommand

type subCommand struct {
	Description string
	Run         func()
}

func init() {
	subCommandMap = make(map[string]subCommand)

	subCommandMap[ctConvert] = subCommand{
		"",
		func() {
			os.Args = os.Args[1:]
			fmt.Println(os.Args)
			src := flag.String("i", "", "input file")
			dst := flag.String("o", "", "output file")
			dsname := flag.String("d", "", "dataset file")
			convType := flag.String("t", "", "conversion type ("+strings.Join(cmd.ConvTypes(), "|")+")")
			flag.Parse()
			if len(*src) == 0 || len(*dst) == 0 || len(*convType) == 0 {
				flag.PrintDefaults()
				os.Exit(1)
			}
			cmd.Convert(*src, *dst, *convType, *dsname)
		},
	}
}

func printDefaultsExit() {
	commName := "exp-run"
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\t%s <command> [options]\n\n", commName)
	fmt.Printf("Commands:\n\n")
	for name, sc := range subCommandMap {
		fmt.Printf("\t%v\t\t%v\n", name, sc.Description)
	}
	fmt.Println()
	fmt.Printf("For usage details of each command, run:\n\n")
	fmt.Printf("\t%s <command> --help\n", commName)
	fmt.Println()
	os.Exit(1)
}
