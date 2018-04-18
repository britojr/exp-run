package pmlearn

import (
	"reflect"
	"sort"
	"testing"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
)

func TestCountValues(t *testing.T) {
	cases := []struct {
		vs     []*vars.Var
		ds     []map[int]int
		result []float64
	}{{
		[]*vars.Var{
			vars.New(1, 2, "", false),
			vars.New(2, 2, "", false),
		},
		[]map[int]int{
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 1, 2: 1},
			{0: 0, 1: 1, 2: 1},
			{0: 0, 1: 0, 2: 1},
		},
		[]float64{3, 0, 1, 2},
	}, {
		[]*vars.Var{
			vars.New(0, 3, "", false),
			vars.New(2, 2, "", false),
		},
		[]map[int]int{
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 0, 2: 0},
			{0: 0, 1: 1, 2: 1},
			{0: 0, 1: 1, 2: 1},
			{0: 0, 1: 0, 2: 1},
			{0: 1, 1: 0, 2: 1},
			{0: 1, 1: 0, 2: 1},
			{0: 2, 1: 0, 2: 1},
		},
		[]float64{3, 0, 0, 3, 2, 1},
	}}
	for _, tt := range cases {
		got := countValues(tt.ds, tt.vs)
		if !reflect.DeepEqual(tt.result, got) {
			t.Errorf("wrong count, want:\n%v\n!=\n%v\n", tt.result, got)
		}
	}
}

func TestLearnParms(t *testing.T) {
	vs := []*vars.Var{
		vars.New(0, 2, "", false),
		vars.New(1, 2, "", false),
		vars.New(2, 2, "", false),
	}
	bn := model.NewBNet()
	for _, v := range vs {
		nd := model.NewBNode(v)
		bn.AddNode(nd)
	}
	bn.Node(vs[0]).SetPotential(factor.New(vs[0]))
	bn.Node(vs[2]).SetPotential(factor.New(vs[2]))
	bn.Node(vs[1]).SetPotential(factor.New(vs[0], vs[1], vs[2]))
	ds := []map[int]int{
		{0: 0, 1: 1, 2: 0},
		{0: 1, 1: 0, 2: 0},
		{0: 0, 1: 0, 2: 1},
		{0: 0, 1: 0, 2: 1},
		{0: 0, 1: 0, 2: 1},
		{0: 0, 1: 0, 2: 1},
		{0: 0, 1: 1, 2: 1},
		{0: 0, 1: 1, 2: 1},
		{0: 1, 1: 0, 2: 1},
		{0: 1, 1: 1, 2: 1},
		// 000: 0
		// 100: 1
		// 010: 1
		// 110: 0
		// 001: 2/3
		// 101: .5
		// 011: 1/3
		// 111: .5
	}
	result := [][]float64{
		{.7, .3},
		{0, 1, 1, 0, 2.0 / 3.0, .5, 1.0 / 3.0, .5},
		{.2, .8},
	}
	learnParms(bn, ds)
	for _, v := range vs {
		got := bn.Node(v).Potential().Values()
		if !reflect.DeepEqual(result[v.ID()], got) {
			t.Errorf("wrong parameters (%v), want:\n%v\n!=\n%v\n", v, result[v.ID()], got)
		}
	}
}

func TestBuildStruct(t *testing.T) {
	vl := []*vars.Var{
		vars.New(0, 2, "0", false),
		vars.New(1, 2, "1", false),
		vars.New(2, 2, "2", false),
		vars.New(3, 2, "3", false),
		vars.New(4, 2, "4", false),
	}
	cases := []struct {
		vs    []*vars.Var
		paMap map[string][]string
	}{{
		vl, map[string][]string{
			"0": []string{"4", "2"},
			"1": []string{"4"},
			"2": []string{},
			"3": []string{"0", "1"},
			"4": []string{},
		},
	}}
	for _, tt := range cases {
		bn := buildStruct(tt.vs, tt.paMap)
		if !bn.Variables().Equal(tt.vs) {
			t.Errorf("wrong set of variables %v", bn.Variables())
		}
		for _, v := range tt.vs {
			nd := bn.Node(v)
			if nd == nil {
				t.Errorf("nil node (%v)", v)
			}
			got := nd.Parents()
			res := tt.paMap[v.Name()]
			sort.Strings(res)
			if !reflect.DeepEqual(conv.Sitoa(got.DumpAsInts()), res) {
				t.Errorf("wrong parents of %v, want\n%v\ngot\n%v\n",
					v.Name(), got.DumpAsInts(), res,
				)
			}
		}
	}
}
