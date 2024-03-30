package helpers

import (
	"fmt"
	"sort"

	"github.com/glethuillier/K4nundrum/groups"
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
