package helpers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/glethuillier/K4nundrum/groups"
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

func (s *StatisticsRecorder) save() {
	// do not save if the original K4 is analyzed
	if s.simulationsCount == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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

	if _, err = file.WriteString(statistics); err != nil {
		fmt.Printf("error writing file: %s", err.Error())
	}
}

func (s *StatisticsRecorder) Update(simulationsCount uint) {
	s.simulationsCount = simulationsCount
}

func (s *StatisticsRecorder) Record(ciphertext string, gs []groups.Group) {
	// same distribution shapes
	s.sameDistributionShapesCount++

	// same distribution shapes AND groups > 3
	segmentsAreAppropriatelySized := segmentsAreAppropriatelySized(gs)
	if segmentsAreAppropriatelySized {
		s.appropriatelySizedGroupsCount++
	}

	// same distribution shapes AND alternating groups
	groupsAlternate := groupsAlternate(ciphertext, gs)
	if groupsAlternate {
		s.alternatingGroupsCount++
	}

	// same distribution shapes AND groups > 2 AND alternates
	// (K4-like pseudo-K4s)
	if segmentsAreAppropriatelySized && groupsAlternate {
		s.k4LikeGroupsCount++
	}

	s.saveFile <- struct{}{}
}

func (s *StatisticsRecorder) GetSameShapesCount() uint {
	return s.sameDistributionShapesCount
}

func (s *StatisticsRecorder) GetK4LikeCount() uint {
	return s.k4LikeGroupsCount
}
