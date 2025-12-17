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
var wordFreq map[string]int

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

func loadWords() {
	wordFreq = make(map[string]int)

	file, _ := os.Open("google-10000.txt")
	scanner := bufio.NewScanner(file)
	rank := 0
	for scanner.Scan() {
		w := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(w) >= 2 {
			wordFreq[w] = rank
			rank++
		}
	}
	file.Close()

	file, _ = os.Open("words_alpha.txt")
	scanner = bufio.NewScanner(file)
	seen := make(map[string]bool)
	for scanner.Scan() {
		w := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(w) < 2 || seen[w] {
			continue
		}
		seen[w] = true

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
			if _, ok := wordFreq[w]; !ok {
				wordFreq[w] = 100000
			}
		}
	}
	file.Close()

	// Sort by frequency
	sort.Slice(words, func(i, j int) bool {
		return wordFreq[words[i]] < wordFreq[words[j]]
	})

	newCounts := make([]map[rune]int, len(words))
	for i, w := range words {
		newCounts[i] = letterCount(w)
	}
	wordCounts = newCounts
}

type Solution struct {
	words  []int
	counts map[rune]int
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

func solutionKey(indices []int) string {
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, idx := range sorted {
		parts[i] = words[idx]
	}
	return strings.Join(parts, "|")
}

func avgFreq(indices []int) float64 {
	sum := 0.0
	for _, idx := range indices {
		sum += float64(wordFreq[words[idx]])
	}
	return sum / float64(len(indices))
}

func worker(id int, solutions chan<- []int, done <-chan struct{}) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))

	temp := 500.0
	cooling := 0.99995

	// Start random
	current := make([]int, 14+r.Intn(5))
	for i := range current {
		current[i] = r.Intn(min(3000, len(words)))
	}
	currentCounts := combineCount(current)

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
			newWords[idx] = r.Intn(min(5000, len(words)))
		} else if op < 85 {
			newWords = append(newWords, r.Intn(min(3000, len(words))))
		} else if len(newWords) > 10 {
			idx := r.Intn(len(newWords))
			newWords = append(newWords[:idx], newWords[idx+1:]...)
		}

		newCounts := combineCount(newWords)
		newMis := mismatch(newCounts)
		oldMis := mismatch(currentCounts)

		// Score includes mismatch and word count
		newScore := float64(newMis)*100 + float64(abs(len(newWords)-(targetSpaces+1)))*10
		oldScore := float64(oldMis)*100 + float64(abs(len(current)-(targetSpaces+1)))*10

		delta := newScore - oldScore
		if delta < 0 || r.Float64() < math.Exp(-delta/temp) {
			current = newWords
			currentCounts = newCounts

			// Report perfect solutions
			if newMis == 0 && len(newWords) == targetSpaces+1 {
				sol := make([]int, len(newWords))
				copy(sol, newWords)
				select {
				case solutions <- sol:
				default:
				}
			}
		}

		temp *= cooling
		if temp < 0.01 {
			temp = 500.0
			current = make([]int, 14+r.Intn(5))
			for i := range current {
				current[i] = r.Intn(min(3000, len(words)))
			}
			currentCounts = combineCount(current)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	loadWords()
	fmt.Printf("Loaded %d words\n", len(words))

	solutions := make(chan []int, 1000)
	done := make(chan struct{})
	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, solutions, done)
		}(i)
	}

	seen := make(map[string]bool)
	results := make([][]int, 0)
	start := time.Now()
	timeout := 45 * time.Second

	go func() {
		for {
			select {
			case sol := <-solutions:
				key := solutionKey(sol)
				if !seen[key] {
					seen[key] = true
					results = append(results, sol)
					freq := avgFreq(sol)
					fmt.Printf("[%.1fs] #%d (freq=%.0f): %s\n",
						time.Since(start).Seconds(), len(results), freq, solutionText(sol))
				}
			case <-done:
				return
			}
		}
	}()

	time.Sleep(timeout)
	close(done)
	wg.Wait()

	// Sort results by average frequency (prefer common words)
	sort.Slice(results, func(i, j int) bool {
		return avgFreq(results[i]) < avgFreq(results[j])
	})

	fmt.Printf("\n=== TOP 10 SOLUTIONS (by word commonality) ===\n")
	for i := 0; i < min(10, len(results)); i++ {
		fmt.Printf("%d. (freq=%.0f) %s\n", i+1, avgFreq(results[i]), solutionText(results[i]))
	}
}
