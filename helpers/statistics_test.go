package helpers

import (
	"testing"

	"github.com/glethuillier/K4nundrum/groups"
)

func TestAppropriatelySizedSegments(t *testing.T) {
	type test struct {
		name                  string
		groups                []groups.Group
		areAppropriatelySized bool
	}

	tests := []test{
		{
			name: "appropriately sized segments",
			groups: []groups.Group{
				{
					Segments: []string{
						"ABCD",
						"EFGHIJKLMN",
					},
				},
				{
					Segments: []string{
						"OPQ",
						"RSTUVWXYZ",
					},
				},
			},
			areAppropriatelySized: true,
		},
		{
			name: "at least one tiny segment (1 letter)",
			groups: []groups.Group{
				{
					Segments: []string{
						"A", // tiny segment (1 letter)
						"BCD",
						"EFG",
					},
				},
				{
					Segments: []string{"HIJ", "KLM"},
				},
			},
			areAppropriatelySized: false,
		},
		{
			name: "at least one tiny segment (2 letters)",
			groups: []groups.Group{
				{
					Segments: []string{
						"AB", // tiny segment (2 letters)
						"CDE",
					},
				},
				{
					Segments: []string{"FGH", "IJK"},
				},
			},
			areAppropriatelySized: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			areAppropriatelySizedGroups := segmentsAreAppropriatelySized(tc.groups)
			if areAppropriatelySizedGroups != tc.areAppropriatelySized {
				t.Errorf("expected: %t, got: %t",
					tc.areAppropriatelySized,
					areAppropriatelySizedGroups,
				)
			}
		})
	}
}

func TestGroupsAlternating(t *testing.T) {
	type test struct {
		name           string
		ciphertext     string
		groups         []groups.Group
		areAlternating bool
	}

	tests := []test{
		{
			name:       "alternating groups (2 groups)",
			ciphertext: "AAYYBBZZ",
			groups: []groups.Group{
				{
					Segments: []string{
						"AA",
						"BB",
					},
				},
				{
					Segments: []string{
						"YY",
						"ZZ",
					},
				},
			},
			areAlternating: true,
		},
		{
			name:       "alternating groups (3 groups)",
			ciphertext: "AAOOYYBBPPZZ",
			groups: []groups.Group{
				{
					Segments: []string{
						"AA",
						"BB",
					},
				},
				{
					Segments: []string{
						"OO",
						"PP",
					},
				},
				{
					Segments: []string{
						"YY",
						"ZZ",
					},
				},
			},
			areAlternating: true,
		},
		{
			name:       "nonalternating groups (2 groups)",
			ciphertext: "AABBYYZZ",
			groups: []groups.Group{
				{
					Segments: []string{
						"AA",
						"BB",
					},
				},
				{
					Segments: []string{
						"YY",
						"ZZ",
					},
				},
			},
			areAlternating: false,
		},
		{
			name:       "nonalternating groups (3 groups)",
			ciphertext: "AAYYOOBBPPZZ",
			groups: []groups.Group{
				{
					Segments: []string{
						"AA",
						"BB",
					},
				},
				{
					Segments: []string{
						"OO",
						"PP",
					},
				},
				{
					Segments: []string{
						"YY",
						"ZZ",
					},
				},
			},
			areAlternating: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groupsAreAlternating := groupsAlternate(tc.ciphertext, tc.groups)
			if groupsAreAlternating != tc.areAlternating {
				t.Errorf("expected: %t, got: %t",
					tc.areAlternating,
					groupsAreAlternating,
				)
			}
		})
	}
}

func TestRecord(t *testing.T) {
	type test struct {
		name                       string
		cipher                     string
		groups                     []groups.Group
		segmentsAppropriatelySized uint
		groupsAlternate            uint
		k4Like                     uint
	}

	tests := []test{
		{
			name:   "unremarkable groups",
			cipher: "ABCDE",
			groups: []groups.Group{
				{
					Segments: []string{"A", "BC"},
				},
				{
					Segments: []string{"D", "E"},
				},
			},
			segmentsAppropriatelySized: 0,
			groupsAlternate:            0,
			k4Like:                     0,
		},
		{
			name:   "non-alternating appropriately sized groups",
			cipher: "ABCDEFGHIJ",
			groups: []groups.Group{
				{
					Segments: []string{"ABC", "DEF"},
				},
				{
					Segments: []string{"GHI", "JKL"},
				},
			},
			segmentsAppropriatelySized: 1,
			groupsAlternate:            0,
			k4Like:                     0,
		},
		{
			name:   "alternating tiny groups",
			cipher: "ABCDEFGH",
			groups: []groups.Group{
				{
					Segments: []string{"A", "CDE"},
				},
				{
					Segments: []string{"B", "FGH"},
				},
			},
			segmentsAppropriatelySized: 0,
			groupsAlternate:            1,
			k4Like:                     0,
		},
		{
			name:   "K4-like groups",
			cipher: "ABCDEFGHIJKL",
			groups: []groups.Group{
				{
					Segments: []string{"ABC", "GHI"},
				},
				{
					Segments: []string{"DEF", "JKL"},
				},
			},
			segmentsAppropriatelySized: 1,
			groupsAlternate:            1,
			k4Like:                     1,
		},
		{
			name: "K4",
			cipher: "OBKR" +
				"UOXOGHULBSOLIFBBWFLRVQQPRNGKSSO" +
				"TWTQSJQSSEKZZWATJKLUDIAWINFBNYP" +
				"VTTMZFPKWGDKZXTJCDIGKUHUAUEKCAR",
			groups: []groups.Group{
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
			segmentsAppropriatelySized: 1,
			groupsAlternate:            1,
			k4Like:                     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := GetStatisticsRecorder()
			recorder.Record(tc.cipher, tc.groups)

			if recorder.appropriatelySizedGroupsCount != tc.segmentsAppropriatelySized {
				t.Errorf("appropriately sized groups — expected: %d, got %d",
					tc.segmentsAppropriatelySized,
					recorder.appropriatelySizedGroupsCount,
				)
			}

			if recorder.alternatingGroupsCount != tc.groupsAlternate {
				t.Errorf("alternating groups — expected: %d, got %d",
					tc.groupsAlternate,
					recorder.alternatingGroupsCount,
				)
			}

			if recorder.k4LikeGroupsCount != tc.k4Like {
				t.Errorf("K4-like — expected: %d, got %d",
					tc.k4Like,
					recorder.k4LikeGroupsCount,
				)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	expectedSimulationsCount := 0
	recorder := GetStatisticsRecorder()

	for i := 0; i < 10_000; i++ {
		recorder.Update(uint(i))

		if recorder.simulationsCount != uint(i) {
			t.Errorf("simulations count — expected: %d, got %d",
				expectedSimulationsCount,
				recorder.simulationsCount,
			)
		}
	}
}
