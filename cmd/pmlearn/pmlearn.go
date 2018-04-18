package pmlearn

import (
	"bufio"
	"flag"
	"log"
	"strings"

	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
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
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		ParmLearn(*src, *dst, *dsname)
	}
}

func ParmLearn(inFile, outFile, dsname string) {
	paMap, vNames := parseParentMat(inFile)
	ds := data.NewDataset(dsname, false)
	vs := ds.Variables()
	for i, name := range vNames {
		vs.FindByID(i).SetName(name)
	}
	bn := buildStruct(vs, paMap)
	learnParms(bn, ds.IntMaps())
	log.Printf("writing %v\n", outFile)
	bn.Write(outFile)
}

func parseParentMat(fname string) (map[string][]string, []string) {
	paMap := make(map[string][]string)
	var vNames []string
	r := ioutl.OpenFile(fname)
	defer r.Close()
	scanner := bufio.NewScanner(r)
	paSep := "<-"
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), paSep, 2)
		if len(line) < 2 {
			break
		}
		vNames = append(vNames, line[0])
		paStr := strings.Split(line[1], ",")
		paMap[line[0]] = paStr[:len(paStr)-1]
	}
	return paMap, vNames
}

func buildStruct(vs vars.VarList, paMap map[string][]string) *model.BNet {
	bn := model.NewBNet()
	for _, v := range vs {
		family := vars.VarList{v}
		for _, j := range paMap[v.Name()] {
			family.Add(vs.FindByName(j))
		}
		nd := model.NewBNode(v)
		nd.SetPotential(factor.NewZeroes(family...))
		bn.AddNode(nd)
	}
	return bn
}

func learnParms(bn *model.BNet, ds []map[int]int) {
	for _, v := range bn.Variables() {
		nd := bn.Node(v)
		family := nd.Potential().Variables()
		pjoint, err := factor.New(family...).SetValues(
			countValues(ds, family),
		).Normalize(v)
		if err != nil {
			log.Printf("warning: var %v, %v\n", v, err.Error())
		}
		nd.SetPotential(pjoint)
	}
}

func countValues(ds []map[int]int, vs []*vars.Var) []float64 {
	strides := make(map[int]int)
	step := 1
	for _, v := range vs {
		strides[v.ID()] = step
		step *= v.NState()
	}
	count := make([]float64, step)
	for _, line := range ds {
		index := 0
		for id, step := range strides {
			index += line[id] * step
		}
		count[index]++
	}
	return count
}
