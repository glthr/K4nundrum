package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/glethuillier/K4nundrum/frequencies"
	"github.com/glethuillier/K4nundrum/groups"
	"github.com/glethuillier/K4nundrum/helpers"
	"github.com/glethuillier/K4nundrum/rectangle"
)

const (
	k4 = "OBKR" +
		"UOXOGHULBSOLIFBBWFLRVQQPRNGKSSO" +
		"TWTQSJQSSEKZZWATJKLUDIAWINFBNYP" +
		"VTTMZFPKWGDKZXTJCDIGKUHUAUEKCAR"
)

type Job struct {
	ciphertext   string
	separator    rune
	simulationId uint
}

// getValidCollections returns collections of groups with
// identical letters frequency distribution shapes
func getValidCollections(
	generator *groups.GroupsGenerator,
	permutation []string,
) []*groups.Collection {
	var validCollections []*groups.Collection

	for _, collection := range generator.GetSuitableCollections(permutation) {
		if frequencies.HaveIdenticalShapes(collection) {
			validCollections = append(validCollections, collection)
		}
	}

	return validCollections
}

func runAnalysis(
	ctx context.Context,
	mu *sync.Mutex,
	job *Job,
	recorder *helpers.StatisticsRecorder,
) {

	// the separator should be immediately surrounded by nonseparators
	// (e.g., a ciphertext containing a doublet separator 'XX' should be excluded)
	for i := 0; i < len(job.ciphertext)-1; i++ {
		if job.ciphertext[i] == byte(job.separator) && job.ciphertext[i+1] == byte(job.separator) {
			return
		}
	}

	// generate permutations of segments split based on a separator
	// example: "AAXBBXC" and separator 'X':
	// "AA", "BB", "C"; "AA", "C", "BB"; etc.
	generator := groups.GetGroupsGenerator()
	for permutation := range helpers.GeneratePermutations(
		helpers.Split(job.ciphertext, job.separator),
	) {
		select {
		case <-ctx.Done():
			return
		default:
			// analyze the collections to identify groups with
			// the same letters frequency shapes
			for _, collection := range getValidCollections(generator, permutation) {
				mu.Lock()

				helpers.PrintContext(job.ciphertext, job.separator, job.simulationId)
				for j, group := range collection.Groups {
					helpers.PrintGroup(group, j)
				}

				helpers.PrintRectangle(job.ciphertext, job.separator)

				recorder.Record(job.ciphertext, job.separator, collection.Groups)
				mu.Unlock()
			}
		}
	}
}

// maxStatsLength is the longest text for which the exact statistics
// can be computed (the number of placements grows with its fifth power)
const maxStatsLength = 160

// analyzeRectangles prints, for each letter that forms at least one
// symmetric rectangle, the context and the rectangles
func analyzeRectangles(ciphertext string) {
	found := false

	for letter := 'A'; letter <= 'Z'; letter++ {
		if len(rectangle.Find(ciphertext, letter)) == 0 {
			continue
		}

		found = true
		helpers.PrintContext(ciphertext, letter, 0)
		helpers.PrintRectangle(ciphertext, letter)
	}

	if !found {
		fmt.Println("No symmetric rectangle found.")
	}
}

// describePlacement describes the five occurrences of a separator
// in the ciphertext
func describePlacement(name, ciphertext string, letter rune) string {
	var positions []int
	for i, l := range ciphertext {
		if l == letter {
			positions = append(positions, i)
		}
	}

	yesNo := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	var p [5]int
	copy(p[:], positions)
	ps := make([]string, len(positions))
	for i, q := range positions {
		ps[i] = fmt.Sprintf("%d", q+1)
	}

	rects := rectangle.FindInPositions(
		[]int{p[0], p[1], p[3], p[4]},
		len(ciphertext),
	)

	rows := make([]string, len(rects))
	center, inset := false, false
	for i, r := range rects {
		rows[i] = fmt.Sprintf("%d", r.Width)
		if _, ok := r.CenterMarkPosition(positions); ok {
			center = true
		}
		inset = inset || r.InsetCorners
	}

	description := fmt.Sprintf(
		"%s, separator %c (positions %s): 2024 pattern %s; %c rectangle %s",
		name,
		letter,
		strings.Join(ps, ", "),
		yesNo(rectangle.Alternates(p, len(ciphertext))),
		letter,
		yesNo(len(rects) > 0),
	)

	if len(rects) > 0 {
		description += ", rows of " + strings.Join(rows, ", ")
	}

	return description + fmt.Sprintf("; center mark %s; inset corners %s",
		yesNo(center), yesNo(inset))
}

// printRectangleStatistics prints the exact statistics for a text
// of the length of the ciphertext, and where the ciphertext stands
func printRectangleStatistics(ciphertext string, isK4 bool) {
	if len(ciphertext) > maxStatsLength {
		fmt.Printf(
			"The text is too long (%d letters) for exact statistics: "+
				"the maximum is %d letters.\n",
			len(ciphertext),
			maxStatsLength,
		)
		return
	}

	fmt.Print(rectangle.ComputeScarcity(len(ciphertext)))
	fmt.Println()

	if isK4 {
		fmt.Println(describePlacement("K4", ciphertext, 'W'))
		return
	}

	found := false
	for letter := 'A'; letter <= 'Z'; letter++ {
		if strings.Count(ciphertext, string(letter)) == 5 {
			found = true
			fmt.Println(describePlacement("Ciphertext", ciphertext, letter))
		}
	}

	if !found {
		fmt.Println("No letter occurs exactly five times in the ciphertext.")
	}
}

func main() {
	var (
		wg               sync.WaitGroup
		simulationsCount uint
		mu               sync.Mutex
	)

	sim := flag.Bool("sim", false, "simulation mode")
	customCiphertext := flag.String(
		"ciphertext",
		"",
		"custom analysis of an arbitrary ciphertext",
	)
	workersCount := flag.Int(
		"workers",
		20,
		"number of workers to process the analysis in parallel",
	)
	rectangleOnly := flag.Bool(
		"rectangle",
		false,
		"rectangle analysis only (with --sim: fast simulation of the rectangle metrics)",
	)
	rectangleStats := flag.Bool(
		"rectangle-stats",
		false,
		"exact statistics on the rectangle and the 2024 pattern",
	)
	flag.Parse()

	analyzedCiphertext := k4
	if *customCiphertext != "" {
		analyzedCiphertext = strings.ToUpper(*customCiphertext)
	}

	if *rectangleStats {
		printRectangleStatistics(analyzedCiphertext, *customCiphertext == "")
		return
	}

	if *rectangleOnly && !*sim {
		analyzeRectangles(analyzedCiphertext)
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	terminateAnalysis := make(chan os.Signal, 1)
	jobs := make(chan Job, 1000)

	simulation := *sim
	recorder := helpers.GetStatisticsRecorder()

	fastRectangleSimulation := *rectangleOnly && simulation

	// start workers
	// (the fast rectangle simulation does not run the groups analysis)
	for w := 1; !fastRectangleSimulation && w <= *workersCount; w++ {
		wg.Add(1)
		go func(ctx context.Context, job <-chan Job, wg *sync.WaitGroup) {
			defer wg.Done()

			for {
				select {
				case j, ok := <-job:
					if !ok {
						return
					}
					runAnalysis(ctx, &mu, &j, recorder)
				case <-ctx.Done():
					return
				}
			}
		}(ctx, jobs, &wg)
	}

	ciphertext := k4

	go func() {
		for {
			if simulation {
				// if simulation is enabled:
				// generate a random pseudo-K4
				ciphertext = helpers.GenerateRandomString(len(k4))
				simulationsCount++
				recorder.Update(simulationsCount)
				recorder.RecordRectangles(ciphertext)

				if fastRectangleSimulation {
					select {
					case <-ctx.Done():
						return
					default:
						continue
					}
				}
			} else if *customCiphertext != "" {
				ciphertext = strings.ToUpper(*customCiphertext)
			}

			// iterate over separators:
			// 'A', 'B', ..., 'Z'
			for separator := 'A'; separator <= 'Z'; separator++ {
				jobs <- Job{
					ciphertext:   ciphertext,
					separator:    separator,
					simulationId: simulationsCount,
				}
			}

			// if K4 has been analyzed:
			// exit gracefully
			if !simulation {
				// signal that all jobs have been sent
				// and wait for the workers to finish
				// their respective tasks
				close(jobs)
				wg.Wait()

				terminateAnalysis <- syscall.SIGTERM
				break
			}
		}
	}()

	signal.Notify(
		terminateAnalysis,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-terminateAnalysis
	cancelFunc()

	// make sure that the statistics file is up to date
	recorder.Save()

	rectangles, tall, center, inset, k4LikeRectangle := recorder.GetRectangleCounts()

	fmt.Println("Analysis Completed.")
	fmt.Printf("Same shapes:\t%d\n", recorder.GetSameShapesCount())
	fmt.Printf("K4-like:\t%d\n", recorder.GetK4LikeCount())
	fmt.Printf("K4-like+rect.:\t%d\n", k4LikeRectangle)

	// per-text metrics are only recorded in simulations
	if simulation {
		fmt.Printf("Rectangle:\t%d\n", rectangles)
		fmt.Printf("Rectangle 5+:\t%d\n", tall)
		fmt.Printf("Rect. center:\t%d\n", center)
		fmt.Printf("Rect. inset:\t%d\n", inset)
	}
}
