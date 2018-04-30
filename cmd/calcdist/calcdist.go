package calcdist

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/utl/ioutl"
	"github.com/britojr/utl/stats"
	"github.com/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

func distFuncs() (dfs []string) {
	for name := range distances {
		dfs = append(dfs, name)
	}
	return
}

// distance functions
var distances = map[string]func(a, b []float64) float64{
	"mse":           stats.MSE,
	"max-abs":       stats.MaxAbsErr,
	"hellinger":     stats.HellDist,
	"cross-entropy": stat.CrossEntropy,
	"kl":            stat.KullbackLeibler,
	// "hellinger-gonum": stat.Hellinger,
	"l1norm": func(a, b []float64) float64 { return floats.Distance(a, b, 1) },
	"l2norm": func(a, b []float64) float64 { return floats.Distance(a, b, 2) },
}

// Cmd command struct
var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "difcalc"
	Cmd.Short = "compute distance between two inference results"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		inFile1 := cm.Flag.String("i1", "", "inference result 1")
		inFile2 := cm.Flag.String("i2", "", "inference result 1")
		outFile := cm.Flag.String("log", "", "output file")
		distOpt := cm.Flag.String("dif", "", "distance function ("+strings.Join(distFuncs(), "|")+")")
		cm.Flag.Parse(args)
		if len(*inFile1) == 0 || len(*inFile2) == 0 || len(*distOpt) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		CalcDist(*inFile1, *inFile2, *outFile, *distOpt)
	}
}

// CalcDist calculates distance between values of two files
func CalcDist(inFile1, inFile2, outFile, distOpt string) {
	if dist, ok := distances[distOpt]; ok {
		f1Vals := readInfFile(inFile1)
		log.Printf("%v: %v values read\n", inFile1, len(f1Vals))
		f2Vals := readInfFile(inFile2)
		log.Printf("%v: %v values read\n", inFile2, len(f2Vals))
		if len(f1Vals) != len(f2Vals) {
			log.Printf("error: different size\n")
			return
		}
		dVal := dist(f1Vals, f2Vals)
		if len(outFile) != 0 {
			f := ioutl.CreateFile(outFile)
			fmt.Fprintf(f, "%v\n", dVal)
			f.Close()
		} else {
			fmt.Printf("%v\n", dVal)
		}
		return
	}
	log.Printf("error: invalid option: (%v)\n\n", distOpt)
	Cmd.Flag.PrintDefaults()
	return
}

func readInfFile(fname string) (vs []float64) {
	f := ioutl.OpenFile(fname)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		v, err := strconv.ParseFloat(scanner.Text(), 64)
		if err == nil {
			vs = append(vs, math.Exp(v))
		}
	}
	return
}
