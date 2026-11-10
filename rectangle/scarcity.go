package rectangle

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
)

// Scarcity holds exact counts over all the placements of
// five separators in a text
type Scarcity struct {
	TextLength, Separators int

	// C(n, 5)
	Placements uint64

	// the two alternating groups of segments have the same total length
	AlternatingGroups uint64

	// AlternatingGroups, AND no two adjacent separators,
	// AND all six segments have at least 3 letters
	K4LikeStructure uint64

	// the four outer separators (1st, 2nd, 4th, 5th) form
	// a symmetric rectangle at some width
	Rectangle uint64

	// ... and the middle separator is the center mark of that rectangle
	// (for some width where the rectangle holds)
	RectangleCenter uint64

	// ... the rectangle (for some width) has inset corners
	RectangleInset uint64

	// AlternatingGroups && Rectangle
	Both uint64

	// K4LikeStructure && Rectangle
	BothK4Like uint64
}

const separators = 5

// Alternates identifies whether five separators at the given
// (ascending) positions of a text of length n split it into two
// alternating groups of segments of the same total length
func Alternates(p [separators]int, n int) bool {
	return p[0]+(p[2]-p[1]-1)+(p[4]-p[3]-1) ==
		(p[1]-p[0]-1)+(p[3]-p[2]-1)+(n-1-p[4])
}

// K4Like identifies whether five separators at the given (ascending)
// positions have the K4 structure: alternating groups, no adjacent
// separators, and segments of at least 3 letters
func K4Like(p [separators]int, n int) bool {
	return Alternates(p, n) &&
		p[0] >= 3 &&
		p[1]-p[0]-1 >= 3 &&
		p[2]-p[1]-1 >= 3 &&
		p[3]-p[2]-1 >= 3 &&
		p[4]-p[3]-1 >= 3 &&
		n-1-p[4] >= 3
}

type dimensions struct {
	w, h, size int
}

// outerRectangle identifies at which widths the outer four separators
// (p0, p1, p3, p4) form a symmetric rectangle, and whether the rectangle
// has inset corners. The sizes (h*w) of the corresponding grids are
// appended to sizes.
func outerRectangle(
	dims []dimensions,
	p0, p1, p3, p4 int,
	sizes []int,
) ([]int, bool) {
	inset := false

	for _, d := range dims {
		if p4 >= d.size {
			continue
		}

		r0, c0 := p0/d.w, p0%d.w
		r3 := p3 / d.w
		if r0 != p1/d.w || r3 != p4/d.w ||
			r0 == r3 || r0+r3 != d.h-1 ||
			c0 != p3%d.w || p1%d.w != p4%d.w ||
			c0+p1%d.w != d.w-1 {
			continue
		}

		sizes = append(sizes, d.size)
		if r0 == 1 && c0 == 1 {
			inset = true
		}
	}

	return sizes, inset
}

func binomial5(n int) uint64 {
	var c uint64 = 1
	for i := 0; i < separators; i++ {
		c = c * uint64(n-i) / uint64(i+1)
	}
	return c
}

// ComputeScarcity enumerates every placement of 5 separators
// in a text of length n
func ComputeScarcity(n int) Scarcity {
	s := Scarcity{TextLength: n, Separators: separators}
	if n < separators {
		return s
	}
	s.Placements = binomial5(n)

	var dims []dimensions
	for w := MinWidth; w <= MaxWidth; w++ {
		if h := n / w; h >= 2 {
			dims = append(dims, dimensions{w: w, h: h, size: h * w})
		}
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup

		// first positions to process
		firsts = make(chan int, n)
	)

	for p0 := 0; p0 <= n-separators; p0++ {
		firsts <- p0
	}
	close(firsts)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var (
				local Scarcity
				sizes = make([]int, 0, MaxWidth)
			)

			for p0 := range firsts {
				for p1 := p0 + 1; p1 < n; p1++ {
					for p2 := p1 + 1; p2 < n; p2++ {
						for p3 := p2 + 1; p3 < n; p3++ {
							for p4 := p3 + 1; p4 < n; p4++ {
								p := [separators]int{p0, p1, p2, p3, p4}

								alternating := Alternates(p, n)
								k4Like := alternating && K4Like(p, n)
								if alternating {
									local.AlternatingGroups++
								}
								if k4Like {
									local.K4LikeStructure++
								}

								// necessary condition for the outer
								// separators to be mirror images
								if p0+p4 != p1+p3 {
									continue
								}

								var inset bool
								sizes, inset = outerRectangle(
									dims, p0, p1, p3, p4, sizes[:0])
								if len(sizes) == 0 {
									continue
								}

								local.Rectangle++
								if inset {
									local.RectangleInset++
								}
								for _, size := range sizes {
									d := 2*p2 - (size - 1)
									if d >= -2 && d <= 2 {
										local.RectangleCenter++
										break
									}
								}
								if alternating {
									local.Both++
								}
								if k4Like {
									local.BothK4Like++
								}
							}
						}
					}
				}
			}

			mu.Lock()
			s.AlternatingGroups += local.AlternatingGroups
			s.K4LikeStructure += local.K4LikeStructure
			s.Rectangle += local.Rectangle
			s.RectangleCenter += local.RectangleCenter
			s.RectangleInset += local.RectangleInset
			s.Both += local.Both
			s.BothK4Like += local.BothK4Like
			mu.Unlock()
		}()
	}

	wg.Wait()
	return s
}

// thousands formats a number with thousands separators
func thousands(v uint64) string {
	digits := fmt.Sprintf("%d", v)

	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}

	return b.String()
}

// String returns the statistics as a table
func (s Scarcity) String() string {
	const (
		labelWidth = 56
		countWidth = 10
	)

	row := func(label string, count, total uint64) string {
		if total == 0 || count == 0 {
			return fmt.Sprintf("%-*s%*s%13s%12s\n",
				labelWidth, label, countWidth, thousands(count), "-", "-")
		}

		return fmt.Sprintf("%-*s%*s%12.4f%%%12s\n",
			labelWidth, label, countWidth, thousands(count),
			float64(count)*100/float64(total),
			thousands(uint64(math.Round(float64(total)/float64(count)))),
		)
	}

	var b strings.Builder

	fmt.Fprintf(&b,
		"Exact statistics: %d separators placed at random among %d positions\n\n",
		s.Separators, s.TextLength,
	)
	fmt.Fprintf(&b, "%*s%*s%13s%12s\n", labelWidth, "", countWidth,
		"placements", "share", "1 in")
	fmt.Fprintf(&b, "%-*s%*s\n", labelWidth, "All placements",
		countWidth, thousands(s.Placements))
	b.WriteString(row("Alternating groups of equal size (2024 pattern)",
		s.AlternatingGroups, s.Placements))
	b.WriteString(row("  with no doublet and segments of 3+ letters",
		s.K4LikeStructure, s.Placements))
	b.WriteString(row("W rectangle (outer separators on mirror cells)",
		s.Rectangle, s.Placements))
	b.WriteString(row("  with the middle separator next to the grid center",
		s.RectangleCenter, s.Placements))
	b.WriteString(row("  with corners one cell in from the grid's corners",
		s.RectangleInset, s.Placements))
	b.WriteString(row("2024 pattern and W rectangle together",
		s.Both, s.Placements))

	// conditional shares
	conditional := func(label string, count, total uint64) {
		counts := fmt.Sprintf("%s / %s", thousands(count), thousands(total))
		if count == 0 || total == 0 {
			fmt.Fprintf(&b, "%-*s%s\n", labelWidth-countWidth-5, label+" ", counts)
			return
		}

		fmt.Fprintf(&b, "%-*s%*s%12.4f%%%12s\n",
			labelWidth-5, label, countWidth+5, counts,
			float64(count)*100/float64(total),
			thousands(uint64(math.Round(float64(total)/float64(count)))),
		)
	}

	conditional("W rectangle among placements with the 2024 pattern",
		s.Both, s.AlternatingGroups)
	conditional("  with no doublet and segments of 3+ letters",
		s.BothK4Like, s.K4LikeStructure)

	return b.String()
}
