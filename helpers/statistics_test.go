package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/glethuillier/K4nundrum/groups"
	"github.com/glethuillier/K4nundrum/rectangle"
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
		k4LikeRectangle            uint
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
			k4LikeRectangle:            1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := GetStatisticsRecorder()
			recorder.Record(tc.cipher, 'W', tc.groups)

			if recorder.k4LikeRectangleCount != tc.k4LikeRectangle {
				t.Errorf("K4-like + rectangle — expected: %d, got %d",
					tc.k4LikeRectangle,
					recorder.k4LikeRectangleCount,
				)
			}

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

// textWithX builds a text of n letters with an 'X' at the given positions;
// the other letters are all different (as far as possible)
func textWithX(n int, positions ...int) string {
	text := make([]byte, n)
	next := 0
	for i := range text {
		if next%26 == 'X'-'A' {
			next++
		}
		text[i] = byte('A' + next%26)
		next++
	}

	for _, p := range positions {
		text[p] = 'X'
	}

	return string(text)
}

func TestRecordRectangles(t *testing.T) {
	const k4 = "OBKR" +
		"UOXOGHULBSOLIFBBWFLRVQQPRNGKSSO" +
		"TWTQSJQSSEKZZWATJKLUDIAWINFBNYP" +
		"VTTMZFPKWGDKZXTJCDIGKUHUAUEKCAR"

	type test struct {
		name       string
		ciphertext string
		rectangle  uint
		tall       uint
		center     uint
		inset      uint
	}

	tests := []test{
		{
			name:       "no rectangle",
			ciphertext: strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 3),
		},
		{
			name:       "K4 (once, although the W forms one rectangle)",
			ciphertext: k4,
			rectangle:  1,
			tall:       1,
			center:     1,
			inset:      1,
		},
		{
			// 4 x 6 grid: neither 5 rows, nor center mark
			name:       "small rectangle",
			ciphertext: textWithX(24, 1, 4, 19, 22),
			rectangle:  1,
		},
		{
			// 5 x 7 grid, with a center mark but corners
			// on the border (no inset corners)
			name:       "tall rectangle with center mark",
			ciphertext: textWithX(35, 0, 6, 28, 34, 17),
			rectangle:  1,
			tall:       1,
			center:     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := GetStatisticsRecorder()
			recorder.RecordRectangles(tc.ciphertext)

			rectangle, tall, center, inset, _ := recorder.GetRectangleCounts()
			if rectangle != tc.rectangle || tall != tc.tall ||
				center != tc.center || inset != tc.inset {
				t.Errorf("expected: %d %d %d %d, got: %d %d %d %d",
					tc.rectangle, tc.tall, tc.center, tc.inset,
					rectangle, tall, center, inset,
				)
			}
		})
	}
}

func TestRecordRectanglesConcurrent(t *testing.T) {
	recorder := GetStatisticsRecorder()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				recorder.Update(uint(j + 1))
				recorder.RecordRectangles(textWithX(24, 1, 4, 19, 22))
			}
		}()
	}
	wg.Wait()

	if rectangle, _, _, _, _ := recorder.GetRectangleCounts(); rectangle != 800 {
		t.Errorf("expected 800 rectangles, got %d", rectangle)
	}
}

func TestFormatRareStatistics(t *testing.T) {
	expected := "Symmetric rectangle      \t0.0415%\t     26734/64446024\n"
	if got := formatRareStatistics("Symmetric rectangle", 26734, 64446024); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}
}

func TestSave(t *testing.T) {
	// stats.txt is written in the working directory
	dir := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	}()

	recorder := GetStatisticsRecorder()
	recorder.Update(10)
	recorder.RecordRectangles(textWithX(24, 1, 4, 19, 22))
	recorder.Save()

	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) != 9 {
		t.Fatalf("expected 9 lines, got %d:\n%s", len(lines), content)
	}

	for i, prefix := range []string{
		"Same distribution shapes",
		"Groups length > 2",
		"Alternating groups",
		"K4-like groups",
		"Symmetric rectangle",
		"Rectangle, 5+ rows",
		"Rectangle + center",
		"Rectangle, inset",
		"K4-like + rectangle",
	} {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d — expected prefix %q, got %q", i, prefix, lines[i])
		}
	}

	if !strings.Contains(lines[4], "10.0000%") ||
		!strings.Contains(lines[4], "1/10") {
		t.Errorf("unexpected rectangle line: %q", lines[4])
	}
}

func TestFormatRectangle(t *testing.T) {
	positions := []int{20, 36, 48, 58, 74}
	rects := rectangle.FindInPositions(positions, 97)
	if len(rects) != 1 {
		t.Fatalf("expected one rectangle, got %d", len(rects))
	}

	expected := "rows of 19 (5 rows): W at (2,2) (2,18) (4,2) (4,18); center mark at (3,11)"
	if got := FormatRectangle(rects[0], 'W', positions); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	// no center mark
	rects = rectangle.FindInPositions([]int{20, 36, 58, 74}, 97)
	expected = "rows of 19 (5 rows): W at (2,2) (2,18) (4,2) (4,18)"
	if got := FormatRectangle(rects[0], 'W', []int{20, 36, 58, 74}); got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}
}
