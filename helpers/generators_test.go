package helpers

import (
	"reflect"
	"sort"
	"testing"
)

func TestSplit(t *testing.T) {
	type test struct {
		name           string
		input          string
		separator      rune
		expectedOutput []string
	}

	tests := []test{
		{
			name:      "two segments, one separator — X",
			input:     "ABCDEFXGHIJKL",
			separator: 'X',
			expectedOutput: []string{
				"ABCDEF",
				"GHIJKL",
			},
		},
		{
			name:      "two segments, doublet separator — XX",
			input:     "ABCDEFXXGHIJKL",
			separator: 'X',
			expectedOutput: []string{
				"ABCDEF",
				"GHIJKL",
			},
		},
		{
			name:      "multiple segments, one separator — X",
			input:     "ABCDEFXGHIJKLXMNOPQR",
			separator: 'X',
			expectedOutput: []string{
				"ABCDEF",
				"GHIJKL",
				"MNOPQR",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := Split(tc.input, tc.separator)
			if !reflect.DeepEqual(
				output,
				tc.expectedOutput,
			) {
				t.Errorf("expected: %v, got %v",
					tc.expectedOutput,
					output,
				)
			}
		})
	}
}

func sort2DSlice(slice [][]string) {
	for _, s := range slice {
		sort.Strings(s)
	}

	sort.Slice(slice, func(i, j int) bool {
		return slice[i][0] < slice[j][0]
	})
}

func TestPermutations(t *testing.T) {
	type test struct {
		name                 string
		input                []string
		expectedPermutations [][]string
	}

	tests := []test{
		{
			name:  "positive test 1",
			input: []string{"A", "B"},
			expectedPermutations: [][]string{
				{
					"A", "B",
				},
				{
					"B", "A",
				},
			},
		},
		{
			name:  "positive test 2",
			input: []string{"ABC", "DEF", "GHI"},
			expectedPermutations: [][]string{
				{
					"ABC", "DEF", "GHI",
				},
				{
					"ABC", "GHI", "DEF",
				},
				{
					"DEF", "ABC", "GHI",
				},
				{
					"DEF", "GHI", "ABC",
				},
				{
					"GHI", "ABC", "DEF",
				},
				{
					"GHI", "DEF", "ABC",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var allPermutations [][]string

			for permutation := range GeneratePermutations(
				tc.input,
			) {
				allPermutations = append(allPermutations, permutation)
			}

			sort2DSlice(tc.expectedPermutations)
			sort2DSlice(allPermutations)

			if !reflect.DeepEqual(
				tc.expectedPermutations,
				allPermutations,
			) {
				t.Errorf("expected: %v, got %v",
					tc.expectedPermutations,
					allPermutations,
				)
			}
		})
	}
}
