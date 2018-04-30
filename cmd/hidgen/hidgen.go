package hidgen

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
	"github.com/kniren/gota/dataframe"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "hidgen"
	Cmd.Short = "generate a cut on internal variables"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		bifFile := cm.Flag.String("m", "", "model in bif format")
		cutFile := cm.Flag.String("c", "", "cut file (list of variables to hide)")
		inFile := cm.Flag.String("i", "", "input file to cut")
		outFile := cm.Flag.String("o", "", "output resulting file")
		num := cm.Flag.Int("n", 0, "number of variables to hide")
		cm.Flag.Parse(args)
		if len(*bifFile) != 0 && len(*cutFile) != 0 && *num != 0 {
			generateCut(*bifFile, *cutFile, *num)
			return
		}
		if len(*cutFile) != 0 && len(*inFile) != 0 && len(*outFile) != 0 {
			applyCut(*cutFile, *inFile, *outFile)
			return
		}
		log.Printf("error: missing arguments!\n")
		cm.Flag.PrintDefaults()
	}
}

func generateCut(bifFile, cutFile string, num int) {
	b, err := bif.ParseStruct(bifFile)
	errchk.Check(err, "")
	f := ioutl.CreateFile(cutFile)
	defer f.Close()
	xs := sampleInternals(b, num)
	sort.Ints(xs)
	log.Printf("create %v with %v variables\n", f.Name(), len(xs))
	fmt.Fprintln(f, strings.Join(conv.Sitoa(xs), ","))
}

func applyCut(cutFile, inFile, outFile string) {
	cf := ioutl.OpenFile(cutFile)
	defer cf.Close()
	scanner := bufio.NewScanner(cf)
	scanner.Scan()
	xs := conv.Satoi(strings.Split(scanner.Text(), ","))
	makeFileCut(inFile, outFile, xs)
}

func sampleInternals(b *bif.Struct, n int) (xs []int) {
	is := b.Internals()
	perm := randSource.Perm(len(is))
	if n > len(is) {
		n = len(is)
	}
	for _, v := range perm[:n] {
		xs = append(xs, is[v].ID())
	}
	return
}

func makeFileCut(fi, fo string, cols []int) {
	log.Printf("creating %v\n", fo)
	df := dataframe.ReadCSV(ioutl.OpenFile(fi), dataframe.HasHeader(false)).Drop(cols)
	err := df.WriteCSV(ioutl.CreateFile(fo), dataframe.WriteHeader(false))
	errchk.Check(err, "")
}
