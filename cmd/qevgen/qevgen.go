package qevgen

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"path"
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
		src := cm.Flag.String("i", "", "input file in bif format")
		dst := cm.Flag.String("o", "", "output file in q/ev format")
		num := cm.Flag.Int("n", 1, "number of queries/evidences to generate")
		cm.Flag.Parse(args)
		if len(*src) == 0 || len(*dst) == 0 {
			fmt.Printf("Error: missing arguments!\n\n")
			cm.Flag.PrintDefaults()
			return
		}
		switch path.Ext(*dst) {
		case ".q":
			QevGenerate(*src, *dst, *num, true)
		case ".ev":
			QevGenerate(*src, *dst, *num, false)
		default:
			fmt.Printf("\n error: output file must be '.q' or '.ev'\n\n")
			cm.Flag.PrintDefaults()
			return
		}
	}
}

func QevGenerate(inpFile, outFile string, num int, query bool) {
	b := bif.ParseStruct(inpFile)
	f := ioutl.CreateFile(outFile)
	defer f.Close()
	if query {
		sampleQuery(b, f, num)
	} else {
		sampleEvid(b, f, num)
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
