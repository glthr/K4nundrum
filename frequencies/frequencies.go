package frequencies

import (
	"reflect"
	"sort"

	"github.com/glethuillier/K4nundrum/groups"
)

func computeLetterFrequency(group *groups.Group) {
	frequency := make(map[rune]int)

	for _, segment := range group.Segments {
		for _, c := range segment {
			if _, ok := frequency[c]; !ok {
				frequency[c] = 1
			} else {
				frequency[c]++
			}
		}
	}

	group.LetterFrequency = frequency
}

// HaveIdenticalShapes identifies whether groups in a given collection
// have the same letter frequency distribution shapes or not
func HaveIdenticalShapes(collection *groups.Collection) bool {
	// first, compute the letter frequency for each group
	for i := 0; i < len(collection.Groups); i++ {
		computeLetterFrequency(&collection.Groups[i])
	}

	// then compare the values, abstracting away the letters
	freqValues := make([][]int, len(collection.Groups))
	for i := range freqValues {
		freqValues[i] = make([]int, 26)
	}

	for i, group := range collection.Groups {
		for _, v := range group.LetterFrequency {
			freqValues[i] = append(freqValues[i], v)
		}
	}

	// sort the values for comparison purposes
	for _, v := range freqValues {
		sort.Ints(v)
	}

	for i := 0; i < len(freqValues)-1; i += 1 {
		if !reflect.DeepEqual(freqValues[i], freqValues[i+1]) {
			return false
		}
	}

	// if the values match, the letter frequencies have
	// the same distribution shapes
	return true
}
