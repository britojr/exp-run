package fstats

import (
	"flag"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/gonum/floats"
)

var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "fstats"
	Cmd.Short = "provide file information"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		src := cm.Flag.String("i", "", "input file")
		cm.Flag.Parse(args)
		if len(*src) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		FileStats(*src)
	}
}

func FileStats(fname string) {
	fmt.Printf("File name: %v\n", fname)
	switch path.Ext(fname) {
	case ".bif":
		bifStats(fname)
	default:
		fmt.Printf("format not supported\n")
		return
	}
}

func bifStats(fname string) {
	b, err := bif.ParseStruct(fname)
	errchk.Check(err, "")
	params := 0
	determ := false
	unnorm := false
	schema := make([]int, len(b.Variables()))
	vnames := make([]string, len(b.Variables()))
	for i, v := range b.Variables() {
		f := b.Factor(v.Name())
		schema[i] = v.NState()
		vnames[i] = v.Name()
		params += len(f.Values())
		for _, p := range f.Values() {
			if p == 1.0 || p == 0.0 {
				determ = true
				break
			}
		}
		g, err := f.Copy().Normalize(v)
		if err != nil {
			fmt.Printf("error: %v on factor %v\n", err, v.Name())
		}
		if !floats.EqualApprox(g.Values(), f.Values(), 1e-6) {
			unnorm = true
		}
	}

	vs, rs, ls, is := len(b.Variables()), len(b.Roots()), len(b.Leafs()), len(b.Internals())
	fmt.Printf("Names: %v\n", strings.Join(vnames, ","))
	fmt.Printf("Schema: %v\n", strings.Join(conv.Sitoa(schema), ","))
	fmt.Printf("Variables: %v\n", vs)
	fmt.Printf("Roots:\t%v\t(%.2f%%)\n", rs, 100.0*float64(rs)/float64(vs))
	fmt.Printf("Leafs:\t%v\t(%.2f%%)\n", ls, 100.0*float64(ls)/float64(vs))
	fmt.Printf("Internals:\t%v\t(%.2f%%)\n", is, 100.0*float64(is)/float64(vs))
	fmt.Printf("Parameters: %v\n", params)
	fmt.Printf("Unnormalized: %v\n", unnorm)
	fmt.Printf("Deterministic values: %v\n", determ)
}
