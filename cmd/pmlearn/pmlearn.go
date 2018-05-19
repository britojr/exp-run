package pmlearn

import (
	"bufio"
	"flag"
	"log"
	"strconv"
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
		dst := cm.Flag.String("o", "", "output file (xml format)")
		dsname := cm.Flag.String("d", "", "dataset file")
		alpha := cm.Flag.Float64("alpha", 0, "smoothing constant to avoid zero probabilities")
		cm.Flag.Parse(args)
		if len(*src) == 0 || len(*dst) == 0 || len(*dsname) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		ParmLearn(*src, *dst, *dsname, *alpha)
	}
}

func ParmLearn(inFile, outFile, dsname string, alpha float64) {
	paMap, vNames := parseParentMat(inFile)
	ds := data.NewDataset(dsname, "", false)
	vs := ds.Variables()
	for i, name := range vNames {
		vs.FindByID(i).SetName(name)
	}
	bn := buildStruct(vs, paMap)
	learnParms(bn, ds.IntMaps(), alpha)
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
		text := strings.Replace(scanner.Text(), ":", paSep, 1)
		text = strings.Replace(text, ",", " ", -1)
		line := strings.SplitN(text, paSep, 2)
		if len(line) < 2 {
			continue
		}
		vNames = append(vNames, line[0])
		paMap[line[0]] = strings.Fields(line[1])
	}
	return paMap, vNames
}

func buildStruct(vs vars.VarList, paMap map[string][]string) *model.BNet {
	bn := model.NewBNet()
	for _, v := range vs {
		family := vars.VarList{v}
		paList, ok := paMap[v.Name()]
		if !ok {
			paList, ok = paMap[strconv.Itoa(v.ID())]
			if !ok {
				log.Printf("warning: cannot find parent list of '%v'\n", v.Name())
				continue
			}
		}
		for _, j := range paList {
			if pa := vs.FindByName(j); pa != nil {
				family.Add(pa)
				continue
			}
			if id, err := strconv.Atoi(j); err == nil && id >= 0 {
				if pa := vs.FindByID(id); pa != nil {
					family.Add(pa)
					continue
				}
				log.Printf("warning: cannot find var id '%v' parent of '%v'\n", j, v.Name())
				continue
			}
			if _, err := strconv.ParseFloat(j, 64); err != nil {
				log.Printf("warning: cannot parse '%v' parent of '%v'\n", j, v.Name())
			}
		}
		nd := model.NewBNode(v)
		nd.SetPotential(factor.NewZeroes(family...))
		bn.AddNode(nd)
	}
	return bn
}

func learnParms(bn *model.BNet, ds []map[int]int, alpha float64) {
	for _, v := range bn.Variables() {
		nd := bn.Node(v)
		family := nd.Potential().Variables()
		values := countValues(ds, family)
		if alpha > 0 {
			for i := range values {
				values[i] += alpha
			}
		}
		pjoint, err := factor.New(family...).SetValues(values).Normalize(v)
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
