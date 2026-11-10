package helpers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/glethuillier/K4nundrum/groups"
	"github.com/glethuillier/K4nundrum/rectangle"
)

type StatisticsRecorder struct {
	mu       sync.Mutex
	saveFile chan struct{}

	// number of generated pseudo-K4
	simulationsCount uint

	// pseudo-K4s with the same letter frequency
	// distribution shapes
	sameDistributionShapesCount uint

	// pseudo-K4s with the same shapes AND
	// appropriately sized (i.e., no tiny group)
	appropriatelySizedGroupsCount uint

	// pseudo-K4s with the same shapes AND
	// alternating groups (e.g., A|B|A|B|A|B)
	alternatingGroupsCount uint

	// K4-like pseudo-K4s
	// (same shapes, non tiny, alternating)
	k4LikeGroupsCount uint

	// pseudo-K4s in which at least one letter forms
	// a symmetric rectangle (any width)
	rectangleCount uint

	// ... with at least one rectangle of 5 rows or more
	rectangleTallCount uint

	// ... with at least one rectangle with a center mark
	rectangleCenterCount uint

	// ... with at least one rectangle with a center mark
	// and inset corners (the K4 shape)
	rectangleInsetCount uint

	// K4-like collections whose separator also
	// forms a symmetric rectangle
	k4LikeRectangleCount uint
}

const filename = "stats.txt"

func GetStatisticsRecorder() *StatisticsRecorder {
	stats := &StatisticsRecorder{
		saveFile: make(chan struct{}),
	}

	go func() {
		// regularly save the statistics
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				stats.save()
			case <-stats.saveFile:
				stats.save()
			}
		}
	}()

	return stats
}

// segmentsAreAppropriatelySized identifies whether segments are
// "appropriately sized" (i.e., its length > 2) or not
func segmentsAreAppropriatelySized(gs []groups.Group) bool {
	for _, g := range gs {
		for _, segment := range g.Segments {
			if len(segment) < 3 {
				return false
			}
		}
	}
	return true
}

// groupsAlternate identifies whether groups are alternating in
// the ciphertext or not
// (this function supports an arbirtrary number of groups)
func groupsAlternate(ciphertext string, gs []groups.Group) bool {
	segmentsPerGroup := make(map[string]int)
	var (
		allSegments []string

		// groups in the order they appear in the ciphertext
		ciphertextOrderedGroups []int
	)

	// first, label the groups
	for i, group := range gs {
		for _, segment := range group.Segments {
			segmentsPerGroup[segment] = i
			allSegments = append(allSegments, segment)
		}
	}

	// then, create a slice that contains the labels of the groups
	// in the order they appear in the ciphertext
	// (e.g., { 0, 1, 0, 1 })
	for {
		groupId, ciphertextLen, found := func() (int, int, bool) {
			for _, segment := range allSegments {
				if strings.HasPrefix(ciphertext, segment) {
					ciphertext = strings.Replace(ciphertext, segment, "", 1)
					return segmentsPerGroup[segment], len(ciphertext), true
				}
			}

			if len(ciphertext) > 0 {
				// remove separator
				ciphertext = ciphertext[1:]
			}

			return -1, len(ciphertext), false
		}()

		if !found {
			if ciphertextLen == 0 {
				// done
				break
			} else {
				// segment was not found because of separator:
				// continue
				continue
			}
		} else {
			ciphertextOrderedGroups = append(ciphertextOrderedGroups, groupId)
		}
	}

	gsLen := len(gs)

	// a) check the first groups
	// (they should all differ)
	//
	// example: { 0, 1, x, x }
	//            |  :
	for i := 0; i < gsLen; i++ {
		for j := 0; j < gsLen; j++ {
			if i != j && ciphertextOrderedGroups[i] == ciphertextOrderedGroups[j] {
				return false
			}
		}
	}

	// b) check that groups are alternating
	//
	// example: { 0, 1, 0, 1 }
	//            |  :  |  :
	for i := 0; i < len(ciphertextOrderedGroups); i++ {
		if ciphertextOrderedGroups[i] != ciphertextOrderedGroups[i%gsLen] {
			return false
		}
	}

	return true
}

func formatStatistics(statsType string, count, totalCount uint) string {
	return fmt.Sprintf("%-25s\t%.2f%%\t%10d/%d\n",
		statsType,
		float64(count*100)/float64(totalCount),
		count,
		totalCount,
	)
}

// formatRareStatistics is formatStatistics with more decimals,
// for events that are rare
func formatRareStatistics(statsType string, count, totalCount uint) string {
	return fmt.Sprintf("%-25s\t%.4f%%\t%10d/%d\n",
		statsType,
		float64(count*100)/float64(totalCount),
		count,
		totalCount,
	)
}

// Save writes the statistics to the file
// (it is safe to call it from any goroutine)
func (s *StatisticsRecorder) Save() {
	s.save()
}

func (s *StatisticsRecorder) save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// do not save if the original K4 is analyzed
	if s.simulationsCount == 0 {
		return
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error when closing file: %s", err.Error())
		}
	}()

	statistics := formatStatistics(
		"Same distribution shapes",
		s.sameDistributionShapesCount,
		s.simulationsCount,
	)

	statistics += formatStatistics(
		"Groups length > 2",
		s.appropriatelySizedGroupsCount,
		s.simulationsCount,
	)

	statistics += formatStatistics(
		"Alternating groups",
		s.alternatingGroupsCount,
		s.simulationsCount,
	)

	statistics += formatStatistics(
		"K4-like groups",
		s.k4LikeGroupsCount,
		s.simulationsCount,
	)

	statistics += formatRareStatistics(
		"Symmetric rectangle",
		s.rectangleCount,
		s.simulationsCount,
	)

	statistics += formatRareStatistics(
		"Rectangle, 5+ rows",
		s.rectangleTallCount,
		s.simulationsCount,
	)

	statistics += formatRareStatistics(
		"Rectangle + center",
		s.rectangleCenterCount,
		s.simulationsCount,
	)

	statistics += formatRareStatistics(
		"Rectangle, inset",
		s.rectangleInsetCount,
		s.simulationsCount,
	)

	statistics += formatRareStatistics(
		"K4-like + rectangle",
		s.k4LikeRectangleCount,
		s.simulationsCount,
	)

	if _, err = file.WriteString(statistics); err != nil {
		fmt.Printf("error writing file: %s", err.Error())
	}
}

func (s *StatisticsRecorder) Update(simulationsCount uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.simulationsCount = simulationsCount
}

// Record records a collection of groups with the same shapes, obtained
// by splitting the ciphertext with the separator
func (s *StatisticsRecorder) Record(
	ciphertext string,
	separator rune,
	gs []groups.Group,
) {
	// same distribution shapes AND groups > 2
	segmentsAreAppropriatelySized := segmentsAreAppropriatelySized(gs)

	// same distribution shapes AND alternating groups
	groupsAlternate := groupsAlternate(ciphertext, gs)

	k4Like := segmentsAreAppropriatelySized && groupsAlternate
	k4LikeRectangle := k4Like &&
		len(rectangle.Find(ciphertext, separator)) > 0

	s.mu.Lock()
	s.sameDistributionShapesCount++

	if segmentsAreAppropriatelySized {
		s.appropriatelySizedGroupsCount++
	}

	if groupsAlternate {
		s.alternatingGroupsCount++
	}

	// same distribution shapes AND groups > 2 AND alternates
	// (K4-like pseudo-K4s)
	if k4Like {
		s.k4LikeGroupsCount++
	}

	if k4LikeRectangle {
		s.k4LikeRectangleCount++
	}
	s.mu.Unlock()

	s.saveFile <- struct{}{}
}

// RecordRectangles records the symmetric rectangles of a pseudo-K4
// (all letters are considered; the pseudo-K4 is counted once per metric)
func (s *StatisticsRecorder) RecordRectangles(ciphertext string) {
	var anyRectangle, tall, center, inset bool

	// positions of each letter
	var positions [26][]int
	for i := 0; i < len(ciphertext); i++ {
		if l := ciphertext[i]; l >= 'A' && l <= 'Z' {
			positions[l-'A'] = append(positions[l-'A'], i)
		}
	}

	for _, ps := range positions {
		for _, r := range rectangle.FindInPositions(ps, len(ciphertext)) {
			anyRectangle = true
			tall = tall || r.Height >= 5
			center = center || r.CenterMark
			inset = inset || (r.InsetCorners && r.CenterMark)
		}
	}

	if !anyRectangle {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rectangleCount++
	if tall {
		s.rectangleTallCount++
	}
	if center {
		s.rectangleCenterCount++
	}
	if inset {
		s.rectangleInsetCount++
	}
}

func (s *StatisticsRecorder) GetSameShapesCount() uint {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sameDistributionShapesCount
}

func (s *StatisticsRecorder) GetK4LikeCount() uint {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.k4LikeGroupsCount
}

// GetRectangleCounts returns the number of pseudo-K4s with a symmetric
// rectangle, with a rectangle of 5+ rows, with a rectangle with a center
// mark, with a rectangle with a center mark and inset corners, and the
// number of K4-like collections whose separator forms a rectangle
func (s *StatisticsRecorder) GetRectangleCounts() (rectangles, tall, center, inset, k4Like uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rectangleCount,
		s.rectangleTallCount,
		s.rectangleCenterCount,
		s.rectangleInsetCount,
		s.k4LikeRectangleCount
}
