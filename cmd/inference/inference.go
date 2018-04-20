package inference

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/exp-run/cmd/convert"
	"github.com/britojr/utl/ioutl"
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
		cm.Flag.Parse(args)
		if len(*mFile) == 0 || len(*qFile) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		Infer(*mFile, *qFile, *evFile)
	}
}

func Infer(mFile, qFile, evFile string) {
	dainame := strings.TrimSuffix(mFile, filepath.Ext(mFile)) + ".uai"
	convert.Convert(mFile, dainame, convert.Bif2uai, "", "", 0.0)
	daiEv, daiQu := dainame+".evid", dainame+".quer"
	qevFile := dainame + ".temp"
	defer os.Remove(qevFile)
	mergeQev(qFile, evFile, qevFile)
	convert.Convert(qevFile, daiQu, convert.Ev2evid, "", "", 0.0)
	convert.Convert(evFile, daiEv, convert.Ev2evid, "", "", 0.0)

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
