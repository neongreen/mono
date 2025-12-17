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

func combineCount(indices []int) map[rune]int {
	counts := make(map[rune]int)
	for _, idx := range indices {
		for r, c := range wordCounts[idx] {
			counts[r] += c
		}
	}
	return counts
}

func loadWords() {
	// Load only the top N most common words that can fit
	file, _ := os.Open("google-10000.txt")
	scanner := bufio.NewScanner(file)
	rank := 0
	for scanner.Scan() && rank < 1500 { // Only top 1500 words
		w := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(w) < 2 {
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
	fmt.Printf("Loaded %d common words that fit\n", len(words))
}

type Solution struct {
	words  []int
	counts map[rune]int
	score  float64
}

func (s *Solution) clone() *Solution {
	w := make([]int, len(s.words))
	copy(w, s.words)
	c := make(map[rune]int)
	for r, cnt := range s.counts {
		c[r] = cnt
	}
	return &Solution{words: w, counts: c, score: s.score}
}

func score(indices []int, counts map[rune]int) float64 {
	mis := mismatch(counts)

	// Word count penalty (16 is target)
	wcDiff := len(indices) - (targetSpaces + 1)
	if wcDiff < 0 {
		wcDiff = -wcDiff
	}

	// Frequency score (lower index = more common)
	freq := 0.0
	for _, idx := range indices {
		freq += float64(idx) / float64(len(words))
	}
	freq /= float64(len(indices))

	return float64(mis)*1000 + float64(wcDiff)*100 + freq*50
}

func randomSolution(r *rand.Rand) *Solution {
	n := 14 + r.Intn(5) // 14-18 words
	indices := make([]int, n)
	for i := range indices {
		indices[i] = r.Intn(len(words))
	}
	counts := combineCount(indices)
	return &Solution{
		words:  indices,
		counts: counts,
		score:  score(indices, counts),
	}
}

func mutate(sol *Solution, r *rand.Rand) *Solution {
	newWords := make([]int, len(sol.words))
	copy(newWords, sol.words)

	op := r.Intn(100)
	if op < 70 && len(newWords) > 0 {
		// Replace one word
		idx := r.Intn(len(newWords))
		newWords[idx] = r.Intn(len(words))
	} else if op < 85 {
		// Add a word
		newWords = append(newWords, r.Intn(len(words)))
	} else if len(newWords) > 10 {
		// Remove a word
		idx := r.Intn(len(newWords))
		newWords = append(newWords[:idx], newWords[idx+1:]...)
	}

	counts := combineCount(newWords)
	return &Solution{
		words:  newWords,
		counts: counts,
		score:  score(newWords, counts),
	}
}

func solutionText(sol *Solution) string {
	ws := make([]string, len(sol.words))
	for i, idx := range sol.words {
		ws[i] = words[idx]
	}
	return strings.Join(ws, " ")
}

func worker(id int, wg *sync.WaitGroup, bestChan chan<- *Solution, done <-chan struct{}) {
	defer wg.Done()
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))

	current := randomSolution(r)
	best := current
	lastReport := best.score

	temp := 500.0
	cooling := 0.99995

	for {
		select {
		case <-done:
			return
		default:
		}

		neighbor := mutate(current, r)
		delta := neighbor.score - current.score
		if delta < 0 || r.Float64() < math.Exp(-delta/temp) {
			current = neighbor
		}

		if current.score < best.score {
			best = current.clone()
			if best.score < lastReport-5 {
				select {
				case bestChan <- best.clone():
					lastReport = best.score
				default:
				}
			}
		}

		temp *= cooling
		if temp < 0.01 {
			temp = 500.0
			current = randomSolution(r)
		}
	}
}

func main() {
	loadWords()
	fmt.Printf("Target: %d letters, %d words\n", len(target), targetSpaces+1)

	// Show some sample words
	fmt.Println("\nSample common words that fit:")
	for i := 0; i < min(30, len(words)); i++ {
		fmt.Printf("  %s\n", words[i])
	}

	numWorkers := runtime.NumCPU()
	fmt.Printf("\nStarting %d workers\n", numWorkers)

	bestChan := make(chan *Solution, 100)
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, bestChan, done)
	}

	var globalBest *Solution
	start := time.Now()
	timeout := 60 * time.Second
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case sol := <-bestChan:
			if globalBest == nil || sol.score < globalBest.score {
				globalBest = sol
				mis := mismatch(sol.counts)
				fmt.Printf("\n[%.1fs] score=%.2f, mismatch=%d, words=%d\n",
					time.Since(start).Seconds(), sol.score, mis, len(sol.words))
				fmt.Printf("  %s\n", solutionText(sol))
			}
		case <-ticker.C:
			if globalBest != nil {
				fmt.Printf("[%.1fs] Current best: score=%.2f\n",
					time.Since(start).Seconds(), globalBest.score)
			}
		case <-time.After(timeout):
			fmt.Println("\nDone")
			close(done)
			wg.Wait()

			if globalBest != nil {
				fmt.Println("\n=== BEST SOLUTION ===")
				mis := mismatch(globalBest.counts)
				fmt.Printf("Mismatch: %d\n", mis)
				fmt.Printf("Words: %d\n", len(globalBest.words))
				fmt.Printf("Solution: %s\n", solutionText(globalBest))

				// Sort and show letters
				var letters []rune
				for _, idx := range globalBest.words {
					for _, r := range words[idx] {
						letters = append(letters, r)
					}
				}
				sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
				fmt.Printf("Letters: %s\n", string(letters))
				fmt.Printf("Target:  %s\n", target)
			}
			return
		}
	}
}
