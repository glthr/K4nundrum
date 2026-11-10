package rectangle

import (
	"reflect"
	"strings"
	"testing"
)

const k4 = "OBKR" +
	"UOXOGHULBSOLIFBBWFLRVQQPRNGKSSO" +
	"TWTQSJQSSEKZZWATJKLUDIAWINFBNYP" +
	"VTTMZFPKWGDKZXTJCDIGKUHUAUEKCAR"

// grid builds a text of n letters, filled with 'A',
// that contains 'X' at the given positions
func grid(n int, positions ...int) string {
	text := []byte(strings.Repeat("A", n))
	for _, p := range positions {
		text[p] = 'X'
	}
	return string(text)
}

func hasWidth(rects []Rectangle, width int) bool {
	for _, r := range rects {
		if r.Width == width {
			return true
		}
	}
	return false
}

func TestFindK4(t *testing.T) {
	expected := []Rectangle{
		{
			Width:        19,
			Height:       5,
			Row:          1,
			Col:          1,
			Positions:    [4]int{20, 36, 58, 74},
			CenterMark:   true,
			InsetCorners: true,
		},
	}

	if got := Find(k4, 'W'); !reflect.DeepEqual(got, expected) {
		t.Errorf("W — expected: %+v, got: %+v", expected, got)
	}

	for letter := 'A'; letter <= 'Z'; letter++ {
		if letter == 'W' {
			continue
		}

		if got := Find(k4, letter); len(got) != 0 {
			t.Errorf("%c — expected no rectangle, got: %+v", letter, got)
		}
	}
}

func TestFind(t *testing.T) {
	type test struct {
		name     string
		text     string
		letter   rune
		width    int // width of the rectangle that is looked for
		expected *Rectangle
	}

	tests := []test{
		{
			name:   "4 x 6 grid",
			text:   grid(24, 1, 4, 19, 22),
			letter: 'X',
			width:  6,
			expected: &Rectangle{
				Width:     6,
				Height:    4,
				Row:       0,
				Col:       1,
				Positions: [4]int{1, 4, 19, 22},
			},
		},
		{
			name:   "center mark and inset corners (5 x 7 grid)",
			text:   grid(35, 8, 12, 22, 26, 17),
			letter: 'X',
			width:  7,
			expected: &Rectangle{
				Width:        7,
				Height:       5,
				Row:          1,
				Col:          1,
				Positions:    [4]int{8, 12, 22, 26},
				CenterMark:   true,
				InsetCorners: true,
			},
		},
		{
			name:   "center mark one cell away from the center",
			text:   grid(35, 8, 12, 22, 26, 18),
			letter: 'X',
			width:  7,
			expected: &Rectangle{
				Width:        7,
				Height:       5,
				Row:          1,
				Col:          1,
				Positions:    [4]int{8, 12, 22, 26},
				CenterMark:   true,
				InsetCorners: true,
			},
		},
		{
			name:   "fifth occurrence too far from the center",
			text:   grid(35, 8, 12, 22, 26, 19),
			letter: 'X',
			width:  7,
			expected: &Rectangle{
				Width:        7,
				Height:       5,
				Row:          1,
				Col:          1,
				Positions:    [4]int{8, 12, 22, 26},
				InsetCorners: true,
			},
		},
		{
			name:     "fewer than four occurrences",
			text:     grid(24, 1, 4, 19),
			letter:   'X',
			expected: nil,
		},
		{
			name: "occurrences in the leftover do not count",
			// 4 x 6 grid + 2 leftover letters: the 4th corner
			// stands in the leftover
			text:     grid(26, 1, 4, 19, 24),
			letter:   'X',
			width:    6,
			expected: nil,
		},
		{
			name:     "middle row",
			text:     grid(30, 12, 17, 13, 16),
			letter:   'X',
			width:    6,
			expected: nil,
		},
		{
			name:     "middle column",
			text:     grid(30, 2, 27, 7, 22),
			letter:   'X',
			width:    5,
			expected: nil,
		},
		{
			name:     "not mirror images",
			text:     grid(24, 1, 4, 19, 21),
			letter:   'X',
			width:    6,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rects := Find(tc.text, tc.letter)

			if tc.expected == nil {
				if hasWidth(rects, tc.width) && tc.width != 0 {
					t.Errorf("expected no rectangle at width %d, got: %+v",
						tc.width, rects)
				}
				if tc.width == 0 && len(rects) != 0 {
					t.Errorf("expected no rectangle, got: %+v", rects)
				}
				return
			}

			for _, r := range rects {
				if reflect.DeepEqual(r, *tc.expected) {
					return
				}
			}
			t.Errorf("expected: %+v, got: %+v", *tc.expected, rects)
		})
	}
}

func TestFindInPositionsOrderAndDuplicates(t *testing.T) {
	// the same four mirror cells, given unordered and with duplicates
	rects := FindInPositions([]int{22, 1, 19, 4, 4, 1}, 24)
	if len(rects) == 0 {
		t.Fatal("expected at least one rectangle")
	}

	for i := 1; i < len(rects); i++ {
		a, b := rects[i-1], rects[i]
		if a.Width > b.Width ||
			(a.Width == b.Width && (a.Row > b.Row ||
				(a.Row == b.Row && a.Col >= b.Col))) {
			t.Errorf("results are not ordered: %+v before %+v", a, b)
		}
	}

	if !hasWidth(rects, 6) {
		t.Errorf("expected a rectangle at width 6, got: %+v", rects)
	}
}

func TestCenterMarkPosition(t *testing.T) {
	r := Rectangle{Width: 19, Height: 5, Positions: [4]int{20, 36, 58, 74}}

	if q, ok := r.CenterMarkPosition([]int{20, 36, 48, 58, 74}); !ok || q != 48 {
		t.Errorf("expected 48, got %d (%t)", q, ok)
	}

	// the center of a 95-letter grid is position 47: 46 and 48 are both
	// one away, 47 is the nearest
	if q, ok := r.CenterMarkPosition([]int{46, 47, 48}); !ok || q != 47 {
		t.Errorf("expected 47, got %d (%t)", q, ok)
	}

	if _, ok := r.CenterMarkPosition([]int{20, 36, 58, 74, 50}); ok {
		t.Error("expected no center mark")
	}
}
