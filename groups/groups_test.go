package groups

import (
	"testing"
)

func TestGroups(t *testing.T) {
	type test struct {
		name        string
		permutation []string
		groupsCount uint
	}

	tests := []test{
		{
			name:        "no collection",
			permutation: []string{"AAAAAAAAAAAAAAAAAAAAAAA", "BB"},
			groupsCount: 0,
		},
		{
			name:        "2 collections of groups",
			permutation: []string{"AA", "BB", "CC", "DD"},

			// expected groups:
			// AA | BB | CC | DD
			// AA BB | CC DD
			groupsCount: 2,
		},
		{
			name:        "3 collections of groups",
			permutation: []string{"AA", "BB", "CC", "DD", "EE", "FF"},

			// expected groups:
			// AA | BB | CC | DD | EE | FF
			// AA BB | CC DD | EE FF
			// AA BB CC | DD EE FF
			groupsCount: 3,
		},
		{
			name: "1 collection of groups (K4)",
			permutation: []string{
				"INFBNYPVTTMZFPK",
				"OBKRUOXOGHULBSOLIFBB",
				"TQSJQSSEKZZ",
				"ATJKLUDIA",
				"FLRVQQPRNGKSSOT",
				"GDKZXTJCDIGKUHUAUEKCAR",
			},
			groupsCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups := GetGroupsGenerator().
				GetSuitableCollections(tc.permutation)
			if uint(len(groups)) != tc.groupsCount {
				t.Errorf("expected: %d, got: %d",
					tc.groupsCount,
					len(groups),
				)
			}
		})
	}
}
