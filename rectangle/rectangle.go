// Package rectangle looks for "symmetric rectangles": sets of four
// occurrences of a letter that stand on mirror-symmetric cells of a grid
// obtained by writing the text in rows of a given width.
package rectangle

import "sort"

const (
	// MinWidth is the smallest row width considered
	MinWidth = 3

	// MaxWidth is the largest row width considered
	MaxWidth = 48
)

// Rectangle is a symmetric rectangle found in a grid
// (the text written in rows of Width letters)
type Rectangle struct {
	Width, Height int // rows of Width letters; Height complete rows
	Row, Col      int // 0-based top-left corner cell; the other three corners are its mirror images

	// 0-based positions in the text, ascending
	Positions [4]int

	// another position stands next to the center of the grid
	CenterMark bool

	// the corners are one cell in from the corners of the grid
	InsetCorners bool
}

// CenterMarkPosition returns the position, among the given positions,
// that stands next to the center of the grid of the rectangle
// (and is not one of its four corners), if any.
// If several positions qualify, the nearest to the center is returned.
func (r Rectangle) CenterMarkPosition(positions []int) (int, bool) {
	var (
		best  int
		bestD = 3
	)

	for _, q := range positions {
		if q == r.Positions[0] || q == r.Positions[1] ||
			q == r.Positions[2] || q == r.Positions[3] {
			continue
		}

		// compare 2*q to h*w-1 to stay in integers
		d := 2*q - (r.Height*r.Width - 1)
		if d < 0 {
			d = -d
		}

		if d <= 2 && d < bestD {
			best, bestD = q, d
		}
	}

	return best, bestD <= 2
}

func contains(sorted []int, p int) bool {
	i := sort.SearchInts(sorted, p)
	return i < len(sorted) && sorted[i] == p
}

// FindInPositions returns every symmetric rectangle formed by the given
// positions in a text of length n (all widths from MinWidth to MaxWidth).
// The results are ordered by width, then row, then column.
func FindInPositions(positions []int, n int) []Rectangle {
	if len(positions) < 4 {
		return nil
	}

	// sorted, without duplicates
	ps := append([]int(nil), positions...)
	sort.Ints(ps)
	uniq := ps[:0]
	for i, p := range ps {
		if p >= 0 && (i == 0 || p != ps[i-1]) {
			uniq = append(uniq, p)
		}
	}
	ps = uniq

	var found []Rectangle

	for w := MinWidth; w <= MaxWidth; w++ {
		h := n / w
		if h < 2 {
			break
		}

		// positions inside the grid
		size := h * w
		in := ps[:sort.SearchInts(ps, size)]
		if len(in) < 4 {
			continue
		}

		// candidates for the top-left corner, in row-major order
		for _, p := range in {
			r, c := p/w, p%w
			if 2*r >= h-1 || 2*c >= w-1 {
				continue
			}

			topRight := r*w + (w - 1 - c)
			bottomLeft := (h-1-r)*w + c
			bottomRight := (h-1-r)*w + (w - 1 - c)
			if !contains(in, topRight) ||
				!contains(in, bottomLeft) ||
				!contains(in, bottomRight) {
				continue
			}

			rect := Rectangle{
				Width:        w,
				Height:       h,
				Row:          r,
				Col:          c,
				Positions:    [4]int{p, topRight, bottomLeft, bottomRight},
				InsetCorners: r == 1 && c == 1,
			}
			sort.Ints(rect.Positions[:])

			_, rect.CenterMark = rect.CenterMarkPosition(in)
			found = append(found, rect)
		}
	}

	return found
}

// Find returns the symmetric rectangles formed by
// the occurrences of letter in text.
func Find(text string, letter rune) []Rectangle {
	var positions []int
	for i, l := range []rune(text) {
		if l == letter {
			positions = append(positions, i)
		}
	}

	return FindInPositions(positions, len([]rune(text)))
}
