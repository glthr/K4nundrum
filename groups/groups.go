package groups

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"sort"
	"strings"
)

type GroupsGenerator struct {
	knownCollections []string
}

type Group struct {
	Segments        []string
	LetterFrequency map[rune]int
}

// suitable groups of segments
type Collection struct {
	Groups []Group
}

func GetGroupsGenerator() *GroupsGenerator {
	return &GroupsGenerator{}
}

// isNewCollection ensures that collections of groups already processed
// are not processed again
func (g *GroupsGenerator) isNewCollection(segments map[uint][]string) bool {
	var allSegments []string

	// sort segments in each group
	for _, segmentsPerGroup := range segments {
		sort.Strings(segmentsPerGroup)
		allSegments = append(allSegments,
			strings.Join(segmentsPerGroup, "."),
		)
	}

	// sort groups
	// (because A|B ⇔ B|A)
	sort.Strings(allSegments)

	h := sha256.New()
	h.Write([]byte(strings.Join(allSegments, "/")))
	allGroups := hex.EncodeToString(h.Sum(nil))

	if slices.Contains(g.knownCollections, allGroups) {
		return false
	}

	g.knownCollections = append(g.knownCollections, allGroups)
	return true
}

// GetSuitableCollections returns collections of groups that can _potentially_
// have the same letter frequency distribution shapes. More specifically, it
// ensures that the groups have the same number of letters.
func (g *GroupsGenerator) GetSuitableCollections(permutation []string) []*Collection {
	var suitableCollections []*Collection

	totalSegmentsLength := func(permutation []string) int {
		size := 0
		for _, segment := range permutation {
			size += len(segment)
		}
		return size
	}(permutation)

	for collectionSize := 2; collectionSize <= len(permutation); collectionSize++ {
		if totalSegmentsLength%collectionSize != 0 {
			continue
		}

		expectedGroupLength := totalSegmentsLength / collectionSize
		segments := make(map[uint][]string)

		var (
			validCollection   bool
			indexMap          uint
			actualGroupLength int
			i                 int
		)

		for j, p := range permutation {
			actualGroupLength += len(p)
			validCollection = true

			if actualGroupLength > expectedGroupLength {
				// groups cannot be suitable (different lengths):
				// skip
				validCollection = false
				break
			} else if actualGroupLength == expectedGroupLength {
				// group length corresponds to the expected size of a group:
				// continue
				segments[indexMap] = permutation[i : j+1]

				actualGroupLength = 0
				i = j + 1
				indexMap++
			}
		}

		if validCollection && g.isNewCollection(segments) {
			var groups []Group

			for _, v := range segments {
				groups = append(groups, Group{
					Segments: v,
				})
			}

			suitableCollections = append(suitableCollections, &Collection{
				Groups: groups,
			})
		}
	}

	return suitableCollections
}
