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
	words   []int
	counts  map[rune]int
	mis     int
	avgRank float64
}

func (s *Solution) clone() *Solution {
	w := make([]int, len(s.words))
	copy(w, s.words)
	c := make(map[rune]int)
	for r, cnt := range s.counts {
		c[r] = cnt
	}
	return &Solution{words: w, counts: c, mis: s.mis, avgRank: s.avgRank}
}

func newSolution(indices []int) *Solution {
	counts := combineCount(indices)
	mis := mismatch(counts)
	avgRank := 0.0
	for _, idx := range indices {
		avgRank += float64(idx)
	}
	avgRank /= float64(len(indices))
	return &Solution{words: indices, counts: counts, mis: mis, avgRank: avgRank}
}

func (s *Solution) score() float64 {
	// Only care about mismatch when we have exactly 16 words
	return float64(s.mis)*1000 + s.avgRank
}

func worker(results chan<- *Solution, done <-chan struct{}, id int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))

	temp := 500.0
	cooling := 0.99995

	// Always start with exactly 16 words
	current := make([]int, 16)
	for i := range current {
		current[i] = r.Intn(len(words))
	}
	currentSol := newSolution(current)
	best := currentSol.clone()

	for {
		select {
		case <-done:
			return
		default:
		}

		// Mutate - but keep exactly 16 words
		newWords := make([]int, 16)
		copy(newWords, current)

		// Replace 1-3 words
		numReplace := 1 + r.Intn(3)
		for i := 0; i < numReplace; i++ {
			idx := r.Intn(16)
			newWords[idx] = r.Intn(len(words))
		}

		newSol := newSolution(newWords)
		delta := newSol.score() - currentSol.score()

		if delta < 0 || r.Float64() < math.Exp(-delta/temp) {
			current = newWords
			currentSol = newSol

			if currentSol.score() < best.score() {
				best = currentSol.clone()
				if best.mis == 0 {
					select {
					case results <- best.clone():
					default:
					}
				}
			}
		}

		temp *= cooling
		if temp < 0.01 {
			temp = 500.0
			for i := range current {
				current[i] = r.Intn(len(words))
			}
			currentSol = newSolution(current)
		}
	}
}

func runTest(maxWords int, duration time.Duration) {
	loadWords(maxWords)
	fmt.Printf("\n=== Top %d words (%d usable) - searching for exactly 16 words ===\n", maxWords, len(words))

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

	go func() {
		for {
			select {
			case sol := <-results:
				key := solutionText(sol.words)
				if !seen[key] {
					seen[key] = true
					allSolutions = append(allSolutions, sol)
					fmt.Printf("  Found #%d (avgRank=%.0f): %s\n", len(allSolutions), sol.avgRank, key)
				}
			case <-done:
				return
			}
		}
	}()

	time.Sleep(duration)
	close(done)
	wg.Wait()

	if len(allSolutions) == 0 {
		fmt.Printf("  No perfect 16-word solutions found\n")
	} else {
		sort.Slice(allSolutions, func(i, j int) bool {
			return allSolutions[i].avgRank < allSolutions[j].avgRank
		})
		fmt.Printf("\n  Best (lowest avg rank): %s\n", solutionText(allSolutions[0].words))
	}
}

func main() {
	for _, size := range []int{1000, 1500, 2000, 3000, 5000} {
		runTest(size, 45*time.Second)
	}
}
