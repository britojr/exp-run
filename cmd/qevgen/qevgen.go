package qevgen

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/ioutl"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "qevgen"
	Cmd.Short = "query and evidence generator"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		bifFile := cm.Flag.String("i", "", "input model in bif format")
		out := cm.Flag.String("o", "", "output file in q/ev format")
		sample := cm.Flag.String("s", "", "sample in csv format")
		num := cm.Flag.Int("n", 1, "number of queries/evidences to generate")
		cm.Flag.Parse(args)
		if len(*bifFile) == 0 || len(*out) == 0 {
			fmt.Printf("Error: missing arguments!\n\n")
			cm.Flag.PrintDefaults()
			return
		}
		QevGenerate(*bifFile, *out, *sample, *num)
	}
}

func QevGenerate(inpFile, outFile, sampFile string, num int) {
	b := bif.ParseStruct(inpFile)
	fq := ioutl.CreateFile(outFile + ".q")
	log.Printf("create %v\n", fq.Name())
	fev := ioutl.CreateFile(outFile + ".ev")
	log.Printf("create %v\n", fev.Name())
	defer fq.Close()
	defer fev.Close()
	if len(sampFile) == 0 {
		sampleQuery(b, fq, num)
		sampleEvid(b, fev, num)
	} else {
		scanner := bufio.NewScanner(ioutl.OpenFile(sampFile))
		for scanner.Scan() {
			read := strings.Split(scanner.Text(), ",")
			write := make([]string, len(read))
			v := sampleVar(b.Roots())
			write[v.ID()] = read[v.ID()]
			writeLine(fq, write)
			write[v.ID()] = ""
			for _, w := range b.Leafs() {
				write[w.ID()] = read[w.ID()]
			}
			writeLine(fev, write)
		}
	}
}

func sampleVar(vs vars.VarList) *vars.Var {
	return vs[randSource.Intn(len(vs))]
}

func sampleState(v *vars.Var) int {
	return randSource.Intn(v.NState())
}

func writeLine(w io.Writer, line []string) {
	for i := range line {
		if line[i] == "" {
			line[i] = "*"
		}
	}
	fmt.Fprintf(w, "%s\n", strings.Join(line, ","))
}

func sampleQuery(b *bif.Struct, w io.Writer, num int) {
	for i := 0; i < num; i++ {
		v := sampleVar(b.Roots())
		state := sampleState(v)
		line := make([]string, len(b.Variables()))
		line[v.ID()] = strconv.Itoa(state)
		writeLine(w, line)
	}
}

func sampleEvid(b *bif.Struct, w io.Writer, num int) {
	for i := 0; i < num; i++ {
		line := make([]string, len(b.Variables()))
		for _, v := range b.Leafs() {
			line[v.ID()] = strconv.Itoa(sampleState(v))
		}
		writeLine(w, line)
	}
}
