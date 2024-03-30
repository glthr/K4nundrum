package frequencies

import (
	"testing"

	"github.com/glethuillier/K4nundrum/groups"
)

func TestLetterFrequencyAnalysis(t *testing.T) {
	type test struct {
		name                   string
		collection             groups.Collection
		sameDistributionShapes bool
	}

	tests := []test{
		{
			name: "identical distribution shapes",
			collection: groups.Collection{
				Groups: []groups.Group{
					{
						Segments: []string{"AAAAA", "Z"},
					},
					{
						Segments: []string{"BBBBB", "Y"},
					},
					{
						Segments: []string{"CCCCC", "X"},
					},
				},
			},
			sameDistributionShapes: true,
		},
		{
			name: "identical distribution shapes (K4)",
			collection: groups.Collection{
				Groups: []groups.Group{
					{
						Segments: []string{
							"OBKRUOXOGHULBSOLIFBB",
							"TQSJQSSEKZZ",
							"INFBNYPVTTMZFPK",
						},
					},
					{
						Segments: []string{
							"FLRVQQPRNGKSSOT",
							"ATJKLUDIA",
							"GDKZXTJCDIGKUHUAUEKCAR",
						},
					},
				},
			},
			sameDistributionShapes: true,
		},
		{
			name: "different distribution shapes",
			collection: groups.Collection{
				Groups: []groups.Group{
					{
						Segments: []string{"AAAAA", "Z"},
					},
					{
						Segments: []string{"BBBBB", "Y"},
					},
					{
						Segments: []string{
							"DCCCC", // D instead of C
							"X",
						},
					},
				},
			},
			sameDistributionShapes: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			identicalShapes := HaveIdenticalShapes(&tc.collection)
			if identicalShapes != tc.sameDistributionShapes {
				t.Errorf("expected: %t, got: %t",
					tc.sameDistributionShapes,
					identicalShapes,
				)
			}
		})
	}
}

func TestComputeLetterFrequency(t *testing.T) {
	group := &groups.Group{
		Segments: []string{
			// all letters used only once
			"CWM",
			"FJORD",
			"BANK",
			"GLYPHS",
			"VEXT",
			"QUIZ",
			// 5 additional occurrences
			"ZZZZZ",
		},
	}

	computeLetterFrequency(group)

	for c := 'A'; c < 'Z'; c++ {
		if group.LetterFrequency[c] != 1 {
			t.Errorf("%s — expected: 1, got: %d",
				string(c),
				group.LetterFrequency[c],
			)
		}
	}

	if group.LetterFrequency['Z'] != 6 {
		t.Errorf("Z — expected: 6, got: %d",
			group.LetterFrequency['Z'],
		)
	}
}

func BenchmarkFrequencyAnalysis(b *testing.B) {
	collection := groups.Collection{
		Groups: []groups.Group{
			{
				Segments: []string{
					"OBKRUOXOGHULBSOLIFBB",
					"TQSJQSSEKZZ",
					"INFBNYPVTTMZFPK",
				},
			},
			{
				Segments: []string{
					"FLRVQQPRNGKSSOT",
					"ATJKLUDIA",
					"GDKZXTJCDIGKUHUAUEKCAR",
				},
			},
		},
	}

	for n := 0; n < b.N; n++ {
		HaveIdenticalShapes(&collection)
	}
}
