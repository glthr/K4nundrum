package rectangle

import (
	"strings"
	"testing"
)

// bruteForceScarcity is a straightforward implementation of ComputeScarcity
func bruteForceScarcity(n int) Scarcity {
	s := Scarcity{TextLength: n, Separators: 5}

	for p0 := 0; p0 < n; p0++ {
		for p1 := p0 + 1; p1 < n; p1++ {
			for p2 := p1 + 1; p2 < n; p2++ {
				for p3 := p2 + 1; p3 < n; p3++ {
					for p4 := p3 + 1; p4 < n; p4++ {
						s.Placements++

						segments := []int{
							p0, p1 - p0 - 1, p2 - p1 - 1,
							p3 - p2 - 1, p4 - p3 - 1, n - 1 - p4,
						}

						odd := segments[0] + segments[2] + segments[4]
						even := segments[1] + segments[3] + segments[5]
						alternating := odd == even

						k4Like := alternating
						for _, length := range segments {
							if length < 3 {
								k4Like = false
							}
						}

						rects := FindInPositions([]int{p0, p1, p3, p4}, n)
						center, inset := false, false
						for _, r := range rects {
							size := r.Width * r.Height
							if d := 2*p2 - (size - 1); d >= -2 && d <= 2 {
								center = true
							}
							if r.InsetCorners {
								inset = true
							}
						}

						if alternating {
							s.AlternatingGroups++
						}
						if k4Like {
							s.K4LikeStructure++
						}
						if len(rects) > 0 {
							s.Rectangle++
							if alternating {
								s.Both++
							}
							if k4Like {
								s.BothK4Like++
							}
						}
						if center {
							s.RectangleCenter++
						}
						if inset {
							s.RectangleInset++
						}
					}
				}
			}
		}
	}

	return s
}

func TestComputeScarcityBruteForce(t *testing.T) {
	for _, n := range []int{5, 12, 31, 40} {
		expected := bruteForceScarcity(n)
		if got := ComputeScarcity(n); got != expected {
			t.Errorf("n = %d — expected: %+v, got: %+v", n, expected, got)
		}
	}
}

func TestComputeScarcityK4(t *testing.T) {
	if testing.Short() {
		t.Skip("enumerates 64 million placements")
	}

	got := ComputeScarcity(97)

	if got.TextLength != 97 || got.Separators != 5 {
		t.Errorf("unexpected text length or separators: %+v", got)
	}

	type test struct {
		name     string
		got      uint64
		expected uint64
	}

	tests := []test{
		{"placements", got.Placements, 64_446_024},
		{"alternating groups", got.AlternatingGroups, 1_272_384},
		{"rectangle", got.Rectangle, 26_734},
		{"rectangle + center", got.RectangleCenter, 1_754},
		{"rectangle + inset", got.RectangleInset, 775},
		{"both", got.Both, 650},
	}

	for _, tc := range tests {
		if tc.got != tc.expected {
			t.Errorf("%s — expected: %d, got: %d", tc.name, tc.expected, tc.got)
		}
	}
}

func TestPlacementHelpers(t *testing.T) {
	// the separators of K4
	k4Separators := [5]int{20, 36, 48, 58, 74}
	if !Alternates(k4Separators, 97) || !K4Like(k4Separators, 97) {
		t.Error("K4 should alternate and be K4-like")
	}

	// alternating, but with an empty segment
	if p := [5]int{0, 10, 20, 30, 40}; K4Like(p, 51) {
		t.Error("segments of fewer than 3 letters should not be K4-like")
	}

	if Alternates([5]int{1, 2, 3, 4, 5}, 97) {
		t.Error("unexpected alternating groups")
	}
}

func TestThousands(t *testing.T) {
	tests := map[uint64]string{
		0:          "0",
		999:        "999",
		1000:       "1,000",
		64446024:   "64,446,024",
		1272384:    "1,272,384",
		1000000000: "1,000,000,000",
	}

	for value, expected := range tests {
		if got := thousands(value); got != expected {
			t.Errorf("%d — expected: %s, got: %s", value, expected, got)
		}
	}
}

func TestScarcityString(t *testing.T) {
	s := Scarcity{
		TextLength:        97,
		Separators:        5,
		Placements:        64_446_024,
		AlternatingGroups: 1_272_384,
		K4LikeStructure:   549_081,
		Rectangle:         26_734,
		RectangleCenter:   1_754,
		RectangleInset:    775,
		Both:              650,
		BothK4Like:        347,
	}

	got := s.String()

	for _, expected := range []string{
		"5 separators placed at random among 97 positions",
		"64,446,024",
		"1,272,384      1.9743%          51",
		"26,734      0.0415%       2,411",
		"1,754      0.0027%      36,742",
		"775      0.0012%      83,156",
		"650      0.0010%      99,148",
		"650 / 1,272,384      0.0511%       1,958",
		"347 / 549,081      0.0632%       1,582",
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("missing %q in:\n%s", expected, got)
		}
	}

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 12 {
		t.Errorf("expected 12 lines, got %d", len(lines))
	}
}
