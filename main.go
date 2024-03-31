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

				recorder.Record(job.ciphertext, collection.Groups)
				mu.Unlock()
			}
		}
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
	flag.Parse()

	ctx, cancelFunc := context.WithCancel(context.Background())

	terminateAnalysis := make(chan os.Signal, 1)
	jobs := make(chan Job, 1000)

	simulation := *sim
	recorder := helpers.GetStatisticsRecorder()

	// start workers
	for w := 1; w <= *workersCount; w++ {
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
	fmt.Println("Analysis Completed.")
	fmt.Printf("Same shapes:\t%d\n", recorder.GetSameShapesCount())
	fmt.Printf("K4-like:\t%d\n", recorder.GetK4LikeCount())
	cancelFunc()
}
