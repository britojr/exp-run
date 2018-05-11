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
	"abs":           stats.MAE,
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
	var result float64
	dist, ok := distances[distOpt]
	if !ok {
		log.Printf("error: invalid option: (%v)\n\n", distOpt)
		Cmd.Flag.PrintDefaults()
		return
	}
	f1 := parseValues(inFile1)
	f2 := parseValues(inFile2)
	if len(f2) < len(f1) || len(f1) < 1 {
		log.Printf("error: size not enough to compare i1=%v i2=%v\n", len(f1), len(f2))
		return
	}
	if len(f1) == 1 {
		result = dist(f1[0], f2[0])
	} else {
		rs := make([]float64, len(f1))
		for i := range rs {
			rs[i] = dist(f1[i], f2[i])
		}
		result = stat.Mean(rs, nil)
	}
	if len(outFile) != 0 {
		f := ioutl.CreateFile(outFile)
		fmt.Fprintf(f, "%v\n", result)
		f.Close()
	} else {
		fmt.Printf("%v\n", result)
	}
}

func parseValues(fname string) (fs [][]float64) {
	f := ioutl.OpenFile(fname)
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	line := scanner.Text()
	f.Close()
	if line == "MAR" {
		fs = readMarFile(fname)
		log.Printf("%v: read %v variables\n", fname, len(fs))
	} else {
		fvals := readInfFile(fname)
		log.Printf("%v: read %v values\n", fname, len(fvals))
		fs = append(fs, fvals)
	}
	return
}

func readMarFile(fname string) (ma [][]float64) {
	r := ioutl.OpenFile(fname)
	defer r.Close()
	mar := ""
	fmt.Fscanln(r, &mar)
	var n int
	fmt.Fscanf(r, "%d", &n)
	ma = make([][]float64, n)
	for i := range ma {
		fmt.Fscanf(r, "%d", &n)
		ma[i] = make([]float64, n)
		for j := range ma[i] {
			fmt.Fscanf(r, "%f", &ma[i][j])
		}
	}
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
