package helpers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glethuillier/K4nundrum/groups"
	"github.com/glethuillier/K4nundrum/rectangle"
)

type pair struct {
	key   rune
	value int
}

// PrintContext prints the ciphertext, its separator,
// and, if applicable, the simulation id
func PrintContext(ciphertext string, separator rune, simulationId uint) {
	fmt.Printf("\n> %s\n  Separator: %s",
		ciphertext,
		string(separator),
	)

	if simulationId == 0 {
		fmt.Printf("\n\n")
	} else {
		fmt.Printf("\tSimulation: #%d\n\n", simulationId)
	}
}

// PrintGroup prints a group and its letter frequency
func PrintGroup(group groups.Group, i int) {
	fmt.Printf("  Group %d:\t", i+1)

	// segments
	for _, segment := range group.Segments {
		fmt.Printf("%s ", segment)
	}
	fmt.Println()

	// letter frequency (descending order)
	pairs := make([]pair, 0)
	for k, v := range group.LetterFrequency {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	keys := make([]rune, len(pairs))
	for i, p := range pairs {
		keys[i] = p.key
	}

	fmt.Printf("  Letter Freq.:\t")
	for _, k := range keys {
		fmt.Printf("%s:%d  ", string(k), group.LetterFrequency[k])
	}
	fmt.Printf("\n\n")
}

// FormatRectangle describes a symmetric rectangle formed by the
// separator (rows and columns are 1-based)
func FormatRectangle(
	r rectangle.Rectangle,
	separator rune,
	positions []int,
) string {
	cell := func(p int) string {
		return fmt.Sprintf("(%d,%d)", p/r.Width+1, p%r.Width+1)
	}

	cells := make([]string, len(r.Positions))
	for i, p := range r.Positions {
		cells[i] = cell(p)
	}

	description := fmt.Sprintf("rows of %d (%d rows): %s at %s",
		r.Width,
		r.Height,
		string(separator),
		strings.Join(cells, " "),
	)

	if r.CenterMark {
		if q, ok := r.CenterMarkPosition(positions); ok {
			description += "; center mark at " + cell(q)
		}
	}

	return description
}

// PrintRectangle prints the symmetric rectangles formed by the
// separator in the ciphertext, if any
func PrintRectangle(ciphertext string, separator rune) {
	rectangles := rectangle.Find(ciphertext, separator)
	if len(rectangles) == 0 {
		return
	}

	var positions []int
	for i, l := range []rune(ciphertext) {
		if l == separator {
			positions = append(positions, i)
		}
	}

	for _, r := range rectangles {
		fmt.Printf("  Rectangle:\t%s\n", FormatRectangle(r, separator, positions))
	}
	fmt.Println()
}
