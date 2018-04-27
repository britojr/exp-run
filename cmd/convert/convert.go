package convert

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/britojr/bnutils/bif"
	"github.com/britojr/exp-run/cmd"
	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
	"github.com/britojr/utl/errchk"
	"github.com/britojr/utl/ioutl"
)

// conversion types
const (
	Bi2bif  = "bi2bif"
	Bi2xml  = "bi2xml"
	Xml2bif = "xml2bif"
	Bif2fg  = "bif2fg"
	Bif2uai = "bif2uai"

	Ev2evid  = "ev2evid"
	Csv2arff = "csv2arff"
	Mo2mar   = "mo2mar"
)

func ConvTypes() []string {
	return []string{Bi2bif, Bi2xml, Xml2bif, Bif2fg, Bif2uai, Ev2evid, Csv2arff, Mo2mar}
}

var Cmd = &cmd.Command{}

func init() {
	Cmd.Name = "convert"
	Cmd.Short = "converts between different types of models"
	Cmd.Flag = flag.NewFlagSet(Cmd.Name, flag.ExitOnError)
	Cmd.Run = func(cm *cmd.Command, args []string) {
		src := cm.Flag.String("i", "", "input file")
		dst := cm.Flag.String("o", "", "output file")
		hdrname := cm.Flag.String("h", "", "header/schema file")
		bname := cm.Flag.String("b", "", "bnet bif file")
		smooth := cm.Flag.Float64("smooth", 0.0, "smooth deterministic probs")
		convType := cm.Flag.String("t", "", "conversion type ("+strings.Join(ConvTypes(), "|")+")")
		cm.Flag.Parse(args)
		if len(*src) == 0 || len(*dst) == 0 || len(*convType) == 0 {
			log.Printf("error: missing arguments!\n")
			cm.Flag.PrintDefaults()
			return
		}
		Convert(*src, *dst, *convType, *hdrname, *bname, *smooth)
	}
}

func Convert(src, dst, convType, hdrname, bname string, smooth float64) {
	log.Printf("converts: (%v) %v -> %v\n", convType, src, dst)
	vs := []*vars.Var{}
	if len(hdrname) != 0 {
		vs = parseHeader(hdrname)
	}
	switch convType {
	case Bi2bif:
		potentials, _ := parseLTMbif(src, vs)
		ct := buildCTree(potentials)
		writeBif(ct, dst)
	case Bi2xml:
		potentials, _ := parseLTMbif(src, vs)
		ct := buildCTree(potentials)
		writeXML(ct, dst)
	case Xml2bif:
		writeXMLToBif(src, dst)
	case Bif2fg:
		writeBifToFG(src, dst)
	case Bif2uai:
		writeBifToUAI(src, dst, smooth)
	case Ev2evid:
		writeEvToEvid(src, dst)
	case Csv2arff:
		if len(vs) == 0 {
			log.Printf("error: header/schema file needed\n")
			Cmd.Flag.PrintDefaults()
			return
		}
		writeCsvToArff(src, dst, vs)
	case Mo2mar:
		writeMoToMar(src, dst)
	default:
		log.Printf("error: invalid conversion option: (%v)\n\n", convType)
		Cmd.Flag.PrintDefaults()
		return
	}
}

func parseHeader(hdrname string) (vs vars.VarList) {
	r := ioutl.OpenFile(hdrname)
	defer r.Close()
	var lines [][]string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		lines = append(lines, strings.Split(scanner.Text(), ","))
	}
	if len(lines) == 1 {
		for i, c := range lines[0] {
			vs.Add(vars.New(i, conv.Atoi(c), "", false))
		}
	} else {
		if len(lines) > 1 {
			for i, name := range lines[0] {
				vs.Add(vars.New(i, conv.Atoi(lines[1][i]), name, false))
			}
		}
	}
	log.Printf("header: %v\n", vs)
	return
}

func writeBif(ct *model.CTree, fname string) {
	f := ioutl.CreateFile(fname)
	defer f.Close()
	fmt.Fprintf(f, "network unknown {}\n")
	vs := ct.Variables()
	for _, v := range vs {
		fmt.Fprintf(f, "variable %v {\n", v.Name())
		fmt.Fprintf(f, "  type discrete [ %v ] { %v };\n", v.NState(), strings.Join(v.States(), ", "))
		fmt.Fprintf(f, "}\n")
	}
	nds := ct.Nodes()
	for _, nd := range nds {
		if nd.Parent() != nil {
			xvs := nd.Variables().Diff(nd.Parent().Variables())
			pavs := nd.Variables().Intersec(nd.Parent().Variables())
			fmt.Fprintf(f, "probability ( %v | %v ) {\n", strings.Join(varNames(xvs), ", "), strings.Join(varNames(pavs), ", "))

			ixf := vars.NewIndexFor(pavs, pavs)
			for !ixf.Ended() {
				attrbMap := ixf.Attribution()
				attrbStr := make([]string, 0, len(attrbMap))
				for _, v := range pavs {
					attrbStr = append(attrbStr, v.States()[attrbMap[v.ID()]])
				}
				p := nd.Potential().Copy()
				p.Reduce(attrbMap).SumOut(pavs...)
				tableInd := strings.Join(attrbStr, ", ")
				tableVal := strings.Join(conv.Sftoa(p.Values()), ", ")
				fmt.Fprintf(f, "  (%v) %v;\n", tableInd, strings.Replace(tableVal, "E+00", "", -1))
				ixf.Next()
			}
		} else {
			fmt.Fprintf(f, "probability ( %v ) {\n", strings.Join(varNames(nd.Variables()), ", "))
			tableVal := strings.Join(conv.Sftoa(nd.Potential().Values()), ", ")
			fmt.Fprintf(f, "  table %v;\n", strings.Replace(tableVal, "E+00", "", -1))
		}
		fmt.Fprintf(f, "}\n")
	}
}

func varNames(vs vars.VarList) (s []string) {
	for _, v := range vs {
		s = append(s, v.Name())
	}
	return
}

func maxID(vs vars.VarList) int {
	if len(vs) > 0 {
		return vs[len(vs)-1].ID()
	}
	return -1
}

func parseLTMbif(fname string, vs vars.VarList) ([]*factor.Factor, vars.VarList) {
	var (
		pots    []*factor.Factor
		nstate  int
		w, name string
		latent  bool
	)
	id := maxID(vs) + 1
	fi := ioutl.OpenFile(fname)
	defer fi.Close()

	_, err := fmt.Fscanf(fi, "%s", &w)
	for err != io.EOF {
		if w == "variable" {
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			v := vs.FindByName(name)
			if v == nil {
				latent = false
				if strings.Index(name, "variable") >= 0 {
					latent = true
				}
				for strings.Index(w, "discrete") != 0 {
					fmt.Fscanf(fi, "%s", &w)
				}
				nstate = conv.Atoi(strings.Trim(w[len("discrete"):], "[]"))
				vs.Add(vars.New(id, nstate, name, latent))
				id++
			}
		}
		if w == "probability" {
			varOrd := make([]*vars.Var, 0, 2)
			clq := vars.VarList{}
			fmt.Fscanf(fi, "%s", &w)
			fmt.Fscanf(fi, "%s", &name)
			name = strings.Trim(name, "\"")
			varOrd = append(varOrd, vs.FindByName(name))
			clq.Add(varOrd[0])
			fmt.Fscanf(fi, "%s", &w)
			if w == "|" {
				fmt.Fscanf(fi, "%s", &name)
				name = strings.Trim(name, "\"")
				varOrd = append(varOrd, vs.FindByName(name))
				clq.Add(varOrd[1])
			}

			for strings.Index(w, "table") != 0 {
				fmt.Fscanf(fi, "%s", &w)
			}
			values := []float64{}
			fmt.Fscanf(fi, "%s", &w)
			for w != "}" {
				w = strings.Trim(w, ";")
				values = append(values, conv.Atof(w))
				fmt.Fscanf(fi, "%s", &w)
			}

			if len(clq) == 1 {
				pots = append(pots, factor.New(clq...).SetValues(values))
			} else {
				// need to invert variable order
				arranged := make([]float64, len(values))
				ixf := vars.NewOrderedIndex(clq, varOrd)
				for _, v := range values {
					arranged[ixf.I()] = v
					ixf.NextRight()
				}
				pots = append(pots, factor.New(clq...).SetValues(arranged))
			}
		}
		_, err = fmt.Fscanf(fi, "%s", &w)
	}
	return pots, vs
}

func buildCTree(fs []*factor.Factor) *model.CTree {
	var r *factor.Factor
	var fi []*factor.Factor

	r, fs = getRoot(fs)
	nd := model.NewCTNode()
	nd.SetPotential(r)
	queue := []*model.CTNode{nd}

	ct := model.NewCTree()
	for len(queue) > 0 {
		nd := queue[0]
		queue = queue[1:]
		ct.AddNode(nd)
		fi, fs = getMaxIntersec(nd.Potential(), fs)
		for _, f := range fi {
			ch := model.NewCTNode()
			ch.SetPotential(f)
			nd.AddChild(ch)
			queue = append(queue, ch)
		}
	}
	return ct
}

func getRoot(fs []*factor.Factor) (*factor.Factor, []*factor.Factor) {
	f, i := fs[0], 0
	for j, g := range fs {
		if len(g.Variables()) < len(f.Variables()) {
			f, i = g, j
		}
	}
	return f, append(fs[:i], fs[i+1:]...)
}

func getMaxIntersec(f *factor.Factor, fs []*factor.Factor) (fi []*factor.Factor, fr []*factor.Factor) {
	max := 0
	for _, g := range fs {
		l := len(f.Variables().Intersec(g.Variables()))
		if l > max {
			max = l
		}
	}
	if max == 0 {
		return nil, fs
	}
	for _, g := range fs {
		l := len(f.Variables().Intersec(g.Variables()))
		if l == max {
			fi = append(fi, g)
		} else {
			fr = append(fr, g)
		}
	}
	return
}

func writeXMLToBif(inFile, outFile string) {
	xmlbn := model.ReadBNetXML(inFile).XMLStruct()

	f := ioutl.CreateFile(outFile)
	defer f.Close()
	if len(xmlbn.Name) == 0 {
		xmlbn.Name = "unknown"
	}
	fmt.Fprintf(f, "network %v {}\n", xmlbn.Name)
	vs := vars.VarList{}
	for i, v := range xmlbn.Variables {
		u := vars.New(i, len(v.States), v.Name, false)
		fmt.Fprintf(f, "variable %v {\n", u.Name())
		fmt.Fprintf(f, "  type discrete [ %v ] { %v };\n", u.NState(), strings.Join(u.States(), ", "))
		fmt.Fprintf(f, "}\n")
		vs.Add(u)
	}
	for _, p := range xmlbn.Probs {
		if len(p.Given) > 0 {
			fmt.Fprintf(f, "probability ( %v | %v ) {\n", p.For[0], strings.Join(p.Given, ", "))
			xv := vs.FindByName(p.For[0])
			pavs := []*vars.Var{}
			for _, name := range p.Given {
				pavs = append(pavs, vs.FindByName(name))
			}
			ixf := vars.NewOrderedIndex(pavs, pavs)
			k := 0
			tableVals := strings.Fields(strings.Trim(p.Table, " "))
			for !ixf.Ended() {
				attrbMap := ixf.Attribution()
				attrbStr := make([]string, 0, len(attrbMap))
				for _, v := range pavs {
					attrbStr = append(attrbStr, v.States()[attrbMap[v.ID()]])
				}
				tableInd := strings.Join(attrbStr, ", ")
				tableVal := strings.Join(tableVals[k:k+xv.NState()], ", ")
				tableVal = strings.Replace(tableVal, "E+00", "", -1)
				fmt.Fprintf(f, "  (%v) %v;\n", tableInd, tableVal)
				ixf.Next()
				k += xv.NState()
			}
		} else {
			fmt.Fprintf(f, "probability ( %v ) {\n", p.For[0])
			tableVal := strings.Replace(strings.Trim(p.Table, " "), " ", ", ", -1)
			tableVal = strings.Replace(tableVal, "E+00", "", -1)
			fmt.Fprintf(f, "  table %v;\n", tableVal)
		}
		fmt.Fprintf(f, "}\n")
	}
}

func writeXML(ct *model.CTree, fname string) {
	f := ioutl.CreateFile(fname)
	defer f.Close()

	bn := model.XMLBIF{BNetXML: ct.XMLStruct()}

	data, err := xml.MarshalIndent(bn, "", "\t")
	errchk.Check(err, "")
	f.Write(data)
}

func writeBifToFG(src, dst string) {
	b, err := bif.ParseStruct(src)
	errchk.Check(err, "")
	w := ioutl.CreateFile(dst)
	defer w.Close()
	fmt.Fprintf(w, "%v\n", len(b.Variables()))
	fmt.Fprintln(w)
	for _, v := range b.Variables() {
		fc := b.Factor(v.Name())
		fmt.Fprintf(w, "%v\n", len(fc.Variables()))
		for _, u := range fc.Variables() {
			fmt.Fprintf(w, "%v ", u.ID())
		}
		fmt.Fprintln(w)
		for _, u := range fc.Variables() {
			fmt.Fprintf(w, "%v ", u.NState())
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%v\n", len(fc.Values()))
		for i, vl := range fc.Values() {
			fmt.Fprintf(w, "%v\t%v\n", i, vl)
		}
		fmt.Fprintln(w)
	}
}

func writeBifToUAI(src, dst string, smooth float64) {
	b, err := bif.ParseStruct(src)
	errchk.Check(err, "")
	w := ioutl.CreateFile(dst)
	defer w.Close()
	fmt.Fprintln(w, "MARKOV")
	fmt.Fprintf(w, "%v\n", len(b.Variables()))
	for _, v := range b.Variables() {
		fmt.Fprintf(w, "%v ", v.NState())
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%v\n", len(b.Variables())) // num factor is the same as num vars for BNs
	for _, v := range b.Variables() {
		fc := b.Factor(v.Name())
		fmt.Fprintf(w, "%v\t", len(fc.Variables()))
		for _, u := range fc.Variables() {
			fmt.Fprintf(w, "%v ", u.ID())
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	for _, v := range b.Variables() {
		fc := b.Factor(v.Name())
		fmt.Fprintf(w, "%v\n", len(fc.Values()))
		ixf := vars.NewOrderedIndex(fc.Variables(), fc.Variables())
		for !ixf.Ended() {
			fmt.Fprintf(w, "%v ", smoothValues(fc.Values(), smooth)[ixf.I()])
			ixf.NextRight()
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w)
	}
}

func smoothValues(values []float64, smooth float64) []float64 {
	ws := append([]float64(nil), values...)
	for i, v := range ws {
		if v == 1.0 {
			ws[i] = v - smooth
		}
		if v == 0.0 {
			ws[i] = v + smooth
		}
	}
	return ws
}

func writeEvToEvid(src, dst string) {
	r := ioutl.OpenFile(src)
	defer r.Close()
	parsed := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		psID, psVL := []string{}, []string{}
		for i, v := range strings.Split(scanner.Text(), ",") {
			if v != "*" {
				psID = append(psID, strconv.Itoa(i))
				psVL = append(psVL, v)
			}
		}
		pstr := strconv.Itoa(len(psID)) + " "
		for i := range psID {
			pstr += psID[i] + " " + psVL[i] + " "
		}
		parsed = append(parsed, pstr)
	}
	w := ioutl.CreateFile(dst)
	defer w.Close()
	fmt.Fprintf(w, "%v\n", len(parsed))
	for _, line := range parsed {
		fmt.Fprintf(w, "%v\n", line)
	}
}

func writeCsvToArff(src, dst string, vs vars.VarList) {
	hdr := "@relation data\n"
	for _, v := range vs {
		states := make([]string, v.NState())
		for j := range states {
			states[j] = strconv.Itoa(j)
		}
		// needs to add a char because BI parser cannot receive a pure integer as a name
		hdr += fmt.Sprintf("@attribute %s {%s}\n", "x"+v.Name(), strings.Join(states, ","))
	}
	hdr += "@data"

	r := ioutl.OpenFile(src)
	w := ioutl.CreateFile(dst)
	fmt.Fprintln(w, hdr)
	_, err := io.Copy(w, r)
	errchk.Check(err, "")
}

func writeMoToMar(src, dst string) {
	r := ioutl.OpenFile(src)
	defer r.Close()
	var parsed [][]float64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		parsed = append(parsed, conv.Satof(strings.Fields(scanner.Text())))
	}
	w := ioutl.CreateFile(dst)
	defer w.Close()
	fmt.Fprintf(w, "MAR\n%v ", len(parsed))
	for _, line := range parsed {
		fmt.Fprintf(w, "%v ", len(line))
		for _, v := range line {
			fmt.Fprintf(w, "%.7f ", v)
		}
	}
	fmt.Fprintln(w)
}
