package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const target = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"
const targetSpaces = 15

var targetCounts map[rune]int
var words []string
var wordCounts []map[rune]int

func init() {
	targetCounts = make(map[rune]int)
	for _, r := range target {
		targetCounts[r]++
	}
}

func letterCount(s string) map[rune]int {
	m := make(map[rune]int)
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			m[r]++
		}
	}
	return m
}

func canFit(wc map[rune]int) bool {
	for r, c := range wc {
		if targetCounts[r] < c {
			return false
		}
	}
	return true
}

func mismatch(counts map[rune]int) int {
	diff := 0
	for r := 'a'; r <= 'z'; r++ {
		d := targetCounts[r] - counts[r]
		if d < 0 {
			d = -d
		}
		diff += d
	}
	return diff
}

func loadWords(maxWords int) {
	words = nil
	wordCounts = nil

	file, _ := os.Open("google-10000.txt")
	scanner := bufio.NewScanner(file)
	rank := 0
	for scanner.Scan() && rank < maxWords {
		w := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(w) < 2 {
			rank++
			continue
		}
		valid := true
		for _, r := range w {
			if r < 'a' || r > 'z' {
				valid = false
				break
			}
		}
		if !valid {
			rank++
			continue
		}
		wc := letterCount(w)
		if canFit(wc) {
			words = append(words, w)
			wordCounts = append(wordCounts, wc)
		}
		rank++
	}
	file.Close()
}

func combineCount(indices []int) map[rune]int {
	counts := make(map[rune]int)
	for _, idx := range indices {
		for r, c := range wordCounts[idx] {
			counts[r] += c
		}
	}
	return counts
}

func solutionText(indices []int) string {
	ws := make([]string, len(indices))
	for i, idx := range indices {
		ws[i] = words[idx]
	}
	return strings.Join(ws, " ")
}

type Solution struct {
	words    []int
	counts   map[rune]int
	mismatch int
	wordDiff int
	avgRank  float64
}

func (s *Solution) clone() *Solution {
	w := make([]int, len(s.words))
	copy(w, s.words)
	c := make(map[rune]int)
	for r, cnt := range s.counts {
		c[r] = cnt
	}
	return &Solution{words: w, counts: c, mismatch: s.mismatch, wordDiff: s.wordDiff, avgRank: s.avgRank}
}

func newSolution(indices []int) *Solution {
	counts := combineCount(indices)
	mis := mismatch(counts)
	wcDiff := len(indices) - (targetSpaces + 1)
	if wcDiff < 0 {
		wcDiff = -wcDiff
	}
	avgRank := 0.0
	for _, idx := range indices {
		avgRank += float64(idx)
	}
	avgRank /= float64(len(indices))

	return &Solution{
		words:    indices,
		counts:   counts,
		mismatch: mis,
		wordDiff: wcDiff,
		avgRank:  avgRank,
	}
}

func (s *Solution) score() float64 {
	return float64(s.mismatch)*1000 + float64(s.wordDiff)*100 + s.avgRank
}

func worker(results chan<- *Solution, done <-chan struct{}, id int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))

	temp := 500.0
	cooling := 0.99995

	current := make([]int, 14+r.Intn(5))
	for i := range current {
		current[i] = r.Intn(len(words))
	}
	currentSol := newSolution(current)
	best := currentSol

	reportThreshold := 5.0 // Report when score improves by this much

	for {
		select {
		case <-done:
			return
		default:
		}

		// Mutate
		newWords := make([]int, len(current))
		copy(newWords, current)

		op := r.Intn(100)
		if op < 70 && len(newWords) > 0 {
			idx := r.Intn(len(newWords))
			newWords[idx] = r.Intn(len(words))
		} else if op < 85 {
			newWords = append(newWords, r.Intn(len(words)))
		} else if len(newWords) > 10 {
			idx := r.Intn(len(newWords))
			newWords = append(newWords[:idx], newWords[idx+1:]...)
		}

		newSol := newSolution(newWords)
		delta := newSol.score() - currentSol.score()

		if delta < 0 || r.Float64() < math.Exp(-delta/temp) {
			current = newWords
			currentSol = newSol

			if currentSol.score() < best.score()-reportThreshold {
				best = currentSol.clone()
				select {
				case results <- best.clone():
				default:
				}
			}
		}

		temp *= cooling
		if temp < 0.01 {
			temp = 500.0
			current = make([]int, 14+r.Intn(5))
			for i := range current {
				current[i] = r.Intn(len(words))
			}
			currentSol = newSolution(current)
		}
	}
}

type DictResult struct {
	size      int
	solutions []*Solution
}

func runWithDictSize(maxWords int, duration time.Duration) DictResult {
	loadWords(maxWords)
	fmt.Printf("\n=== Testing with top %d words (%d usable) ===\n", maxWords, len(words))

	results := make(chan *Solution, 1000)
	done := make(chan struct{})
	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(results, done, id)
		}(i)
	}

	var allSolutions []*Solution
	seen := make(map[string]bool)
	_ = time.Now() // track start

	go func() {
		for {
			select {
			case sol := <-results:
				// Create unique key
				key := fmt.Sprintf("%d-%d-%.0f", sol.mismatch, len(sol.words), sol.avgRank)
				if !seen[key] {
					seen[key] = true
					allSolutions = append(allSolutions, sol)
				}
			case <-done:
				return
			}
		}
	}()

	time.Sleep(duration)
	close(done)
	wg.Wait()

	// Sort by score (best first)
	sort.Slice(allSolutions, func(i, j int) bool {
		return allSolutions[i].score() < allSolutions[j].score()
	})

	// Print top 5
	fmt.Printf("  Top 5 solutions found:\n")
	for i := 0; i < min(5, len(allSolutions)); i++ {
		sol := allSolutions[i]
		fmt.Printf("  %d. mismatch=%d, words=%d, avgRank=%.0f\n",
			i+1, sol.mismatch, len(sol.words), sol.avgRank)
		fmt.Printf("     %s\n", solutionText(sol.words))
	}

	return DictResult{size: maxWords, solutions: allSolutions}
}

func main() {
	dictSizes := []int{1000, 1500, 2000, 3000}
	allResults := make([]DictResult, 0)

	for _, size := range dictSizes {
		result := runWithDictSize(size, 40*time.Second)
		allResults = append(allResults, result)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("FINAL SUMMARY - Best solution for each dictionary size")
	fmt.Println(strings.Repeat("=", 70))

	for _, result := range allResults {
		fmt.Printf("\nTop %d words:\n", result.size)
		if len(result.solutions) > 0 {
			best := result.solutions[0]
			status := ""
			if best.mismatch == 0 && best.wordDiff == 0 {
				status = " *** PERFECT ***"
			}
			fmt.Printf("  Best: mismatch=%d, words=%d%s\n", best.mismatch, len(best.words), status)
			fmt.Printf("  %s\n", solutionText(best.words))
		} else {
			fmt.Printf("  No solutions found\n")
		}
	}
}
