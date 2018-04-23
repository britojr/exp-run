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
)

// distance functions
const (
	Mse = "mse" //mean squared error
)

func DistFuncs() []string {
	return []string{Mse}
}

var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "difcalc"
	Cmd.Short = "compute distance between two inference results"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		inFile1 := cm.Flag.String("i1", "", "inference result 1")
		inFile2 := cm.Flag.String("i2", "", "inference result 1")
		outFile := cm.Flag.String("log", "", "output file")
		distOpt := cm.Flag.String("dif", Mse, "distance function ("+strings.Join(DistFuncs(), "|")+")")
		cm.Flag.Parse(args)
		if len(*inFile1) == 0 || len(*inFile2) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		CalcDist(*inFile1, *inFile2, *outFile, *distOpt)
	}
}

func CalcDist(inFile1, inFile2, outFile, distOpt string) {
	switch distOpt {
	case Mse:
		calcMSE(inFile1, inFile2, outFile)
	default:
		log.Printf("error: invalid option: (%v)\n\n", distOpt)
		Cmd.Flag.PrintDefaults()
		return
	}
}

func calcMSE(inFile1, inFile2, outFile string) {
	f1Vals := readInfFile(inFile1)
	log.Printf("%v: %v values read\n", inFile1, len(f1Vals))
	f2Vals := readInfFile(inFile2)
	log.Printf("%v: %v values read\n", inFile2, len(f2Vals))
	if len(f1Vals) != len(f2Vals) {
		return
	}
	mse := stats.MSE(f1Vals, f2Vals)
	if len(outFile) != 0 {
		f := ioutl.CreateFile(outFile)
		fmt.Fprintf(f, "%v\n", mse)
		f.Close()
	} else {
		fmt.Printf("%v\n", mse)
	}
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
