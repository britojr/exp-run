package sample

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/convert"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/cmdsh"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
var Cmd = &cmd.Command{}

// const extensions
const (
	cTrain = ".train"
	cTest  = ".test"
	cValid = ".valid"
)

func init() {
	Cmd.Name = "sample"
	Cmd.Short = "sample data and header"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		bifFile := cm.Flag.String("m", "", "input model in bif format")
		outFile := cm.Flag.String("o", "", "basename of file to write tr/te/va format")
		nTrain := cm.Flag.Int("tr", 0, "number of samples for training set")
		nTest := cm.Flag.Int("te", 0, "number of samples for testing set")
		nValid := cm.Flag.Int("va", 0, "number of samples for validation set")
		cm.Flag.Parse(args)
		if len(*bifFile) == 0 || len(*outFile) == 0 || *nTrain+*nTest+*nValid == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		Generate(*bifFile, *outFile, *nTrain, *nTest, *nValid)
	}
}

func Generate(bifFile, outFile string, nTrain, nTest, nValid int) {
	uaiFile := strings.TrimSuffix(bifFile, filepath.Ext(bifFile)) + ".uai"
	convert.Convert(bifFile, uaiFile, convert.Bif2uai, "", "", 0.0)

	runSample(uaiFile, outFile+cTrain, nTrain)
	runSample(uaiFile, outFile+cTest, nTest)
	runSample(uaiFile, outFile+cValid, nValid)

	b, err := bif.ParseStruct(bifFile)
	errchk.Check(err, "")
	writeHeaders(b.Variables(), outFile)

}

func runSample(mdName, outName string, nSamp int) {
	if nSamp <= 0 {
		return
	}
	cmdsh.ExecPrint(fmt.Sprintf(
		"example_gibbs %s %d %v",
		mdName, nSamp, outName,
	), 0)
}

func writeHeaders(vs vars.VarList, outFile string) {
	cards, names, maxs := make([]string, len(vs)), make([]string, len(vs)), make([]string, len(vs))
	for i, v := range vs {
		cards[i] = strconv.Itoa(v.NState())
		names[i] = v.Name()
		maxs[i] = strconv.Itoa(v.NState() - 1)
	}
	f := ioutl.CreateFile(outFile + ".schema")
	fmt.Fprintf(f, "%s\n", strings.Join(cards, ","))
	f.Close()
	fh := ioutl.CreateFile(outFile + ".hdr")
	fmt.Fprintf(fh, "%s\n", strings.Join(names, ","))
	fmt.Fprintf(fh, "%s\n", strings.Join(maxs, ","))
	fh.Close()
}
