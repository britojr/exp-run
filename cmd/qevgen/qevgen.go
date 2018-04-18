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
		out := cm.Flag.String("o", "", "basename of file to write q/ev format")
		sample := cm.Flag.String("s", "", "sample in csv format")
		num := cm.Flag.Int("n", 1, "number of queries/evidences to generate")
		maxLfs := cm.Flag.Int("maxlfs", -1, "max number of leafs to use as evidence")
		cm.Flag.Parse(args)
		if len(*bifFile) == 0 || len(*out) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		QevGenerate(*bifFile, *out, *sample, *num, *maxLfs)
	}
}

func QevGenerate(inpFile, outFile, sampFile string, num, maxLfs int) {
	b := bif.ParseStruct(inpFile)
	fq := ioutl.CreateFile(outFile + ".q")
	log.Printf("create %v\n", fq.Name())
	fev := ioutl.CreateFile(outFile + ".ev")
	log.Printf("create %v\n", fev.Name())
	defer fq.Close()
	defer fev.Close()
	if len(sampFile) == 0 {
		for i := 0; i < num; i++ {
			sampleLine(b, fq, fev, nil, maxLfs)
		}
	} else {
		scanner := bufio.NewScanner(ioutl.OpenFile(sampFile))
		for scanner.Scan() {
			read := strings.Split(scanner.Text(), ",")
			sampleLine(b, fq, fev, read, maxLfs)
		}
	}
}

func sampleLine(b *bif.Struct, fq, fev io.Writer, read []string, maxLfs int) {
	write := make([]string, len(b.Variables()))
	v := sampleVar(b.Roots())
	write[v.ID()] = sampleState(v, read)
	writeLine(fq, write)
	write[v.ID()] = ""
	lfs := b.Leafs()
	if maxLfs < 0 || maxLfs > len(lfs) {
		maxLfs = len(lfs)
	} else {
		randSource.Shuffle(len(lfs), func(i int, j int) {
			lfs[i], lfs[j] = lfs[j], lfs[i]
		})
	}
	for _, w := range lfs[:maxLfs] {
		write[w.ID()] = sampleState(w, read)
	}
	writeLine(fev, write)
}

func sampleVar(vs vars.VarList) *vars.Var {
	return vs[randSource.Intn(len(vs))]
}

func sampleState(v *vars.Var, line []string) string {
	if len(line) > 0 {
		return line[v.ID()]
	}
	return strconv.Itoa(randSource.Intn(v.NState()))
}

func writeLine(w io.Writer, line []string) {
	for i := range line {
		if line[i] == "" {
			line[i] = "*"
		}
	}
	fmt.Fprintf(w, "%s\n", strings.Join(line, ","))
}
