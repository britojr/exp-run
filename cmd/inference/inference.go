package inference

import (
	"flag"
	"log"

	"github.com/britojr/exp-run/cmd"
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
	// convert.Convert(mFile, dst, comvert. convType, dsname, smooth)
}
