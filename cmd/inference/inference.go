package inference

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/convert"
	"github.com/britojr/utl/cmdsh"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/ioutl"
	"github.com/gonum/floats"
)

var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "infer"
	Cmd.Short = "performs inference on a given model"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		mFile := cm.Flag.String("m", "", "input model in bif format")
		qFile := cm.Flag.String("q", "", "query file")
		evFile := cm.Flag.String("ev", "", "evidence file")
		logFile := cm.Flag.String("log", "", "output file")
		cm.Flag.Parse(args)
		if len(*mFile) == 0 || len(*qFile) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		Infer(*mFile, *qFile, *evFile, *logFile)
	}
}

func Infer(mFile, qFile, evFile, logFile string) {
	basename := strings.TrimSuffix(mFile, filepath.Ext(mFile))
	dainame := basename + ".uai"
	convert.Convert(mFile, dainame, convert.Bif2uai, "", "", 0.0)
	var probQev []float64
	if len(evFile) != 0 {
		daiEv, daiQu := dainame+".evid", dainame+".quer"
		qevFile := dainame + ".temp"
		defer os.Remove(qevFile)
		mergeQev(qFile, evFile, qevFile)
		convert.Convert(qevFile, daiQu, convert.Ev2evid, "", "", 0.0)
		convert.Convert(evFile, daiEv, convert.Ev2evid, "", "", 0.0)
		probQev = computeProb(dainame, daiQu)
		probEv := computeProb(dainame, daiEv)
		floats.Sub(probQev, probEv)
	} else {
		daiQu := dainame + ".quer"
		convert.Convert(qFile, daiQu, convert.Ev2evid, "", "", 0.0)
		probQev = computeProb(dainame, daiQu)
	}
	if len(logFile) == 0 {
		writeProbs(basename+".infkey", probQev)
	} else {
		writeProbs(logFile, probQev)
	}
}

func mergeQev(qFile, evFile, qevFile string) {
	rq := ioutl.OpenFile(qFile)
	defer rq.Close()
	rev := ioutl.OpenFile(evFile)
	defer rev.Close()
	w := ioutl.CreateFile(qevFile)
	defer w.Close()
	scq := bufio.NewScanner(rq)
	scev := bufio.NewScanner(rev)
	for scq.Scan() {
		scev.Scan()
		line := strings.Split(scev.Text(), ",")
		for i, v := range strings.Split(scq.Text(), ",") {
			if v != "*" {
				line[i] = v
			}
		}
		fmt.Fprintln(w, strings.Join(line, ","))
	}
}

func computeProb(mdName, evName string) []float64 {
	seed := time.Now().UnixNano()
	cmdsh.ExecPrint(fmt.Sprintf(
		"uai2010-aie-solver %s %s %v PR",
		mdName, evName, seed,
	), 0)
	return parsePR(mdName + ".PR")
}

func writeProbs(fname string, probs []float64) {
	w := ioutl.CreateFile(fname)
	defer w.Close()
	sum := 0.0
	for _, v := range probs {
		sum += v
		fmt.Fprintf(w, "%.8f\n", v)
	}
	if len(probs) > 0 {
		fmt.Fprintf(w, "avg = %.8f\n", sum/float64(len(probs)))
	}
}

func parsePR(fname string) (fs []float64) {
	r := ioutl.OpenFile(fname)
	defer r.Close()
	scanner := bufio.NewScanner(r)
	scanner.Scan() //read PR header
	for scanner.Scan() {
		if strings.Index(scanner.Text(), "BEGIN") >= 0 {
			continue
		}
		n := conv.Atoi(scanner.Text())
		if len(fs) < n {
			fs = make([]float64, n)
		}
		for i := 0; i < n; i++ {
			scanner.Scan()
			fs[i] = conv.Atof(scanner.Text())
		}
	}
	return
}
