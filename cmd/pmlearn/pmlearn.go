package pmlearn

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/ioutl"
)

var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "pmlearn"
	Cmd.Short = "parameter learning with complete data"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		src := cm.Flag.String("i", "", "input file (in list of parents format)")
		dst := cm.Flag.String("o", "", "output file")
		dsname := cm.Flag.String("d", "", "dataset file")
		cm.Flag.Parse(args)
		if len(*src) == 0 || len(*dst) == 0 || len(*dsname) == 0 {
			fmt.Printf("Error: missing arguments!\n\n")
			cm.Flag.PrintDefaults()
			return
		}
		ParmLearn(*src, *dst, *dsname)
	}
}

func ParmLearn(inFile, outFile, dsname string) {
	paMap := parseParentMat(inFile)
	ds := data.NewDataset(dsname, false)
	bn := buildStruct(ds.Variables(), paMap)
	learnParms(bn, ds)
	log.Printf("writing %v\n", outFile)
	bn.Write(outFile)
}

func parseParentMat(fname string) map[int][]int {
	paMap := make(map[int][]int)
	r := ioutl.OpenFile(fname)
	defer r.Close()
	scanner := bufio.NewScanner(r)
	paSep := "<-"
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), paSep, 2)
		if len(line) < 2 {
			break
		}
		paStr := strings.Split(line[1], ",")
		paMap[conv.Atoi(line[0])] = conv.Satoi(paStr[:len(paStr)-1])
	}
	return paMap
}

func buildStruct(vs vars.VarList, paMap map[int][]int) *model.BNet {
	bn := model.NewBNet()
	for _, v := range vs {
		family := vars.VarList{v}
		for _, j := range paMap[v.ID()] {
			family.Add(vs.FindByID(j))
		}
		nd := model.NewBNode(v)
		nd.SetPotential(factor.NewZeroes(family...))
		bn.AddNode(nd)
	}
	return bn
}

func learnParms(bn *model.BNet, ds *data.Dataset) {

}
