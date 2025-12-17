package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const target = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"

var targetCount map[rune]int
var words []string
var wordCounts []map[rune]int
var wordFreq map[string]int

func init() {
	targetCount = make(map[rune]int)
	for _, r := range target {
		targetCount[r]++
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

func canUseWord(remaining, word map[rune]int) bool {
	for r, c := range word {
		if remaining[r] < c {
			return false
		}
	}
	return true
}

func subtract(a, b map[rune]int) map[rune]int {
	result := make(map[rune]int)
	for r, c := range a {
		result[r] = c - b[r]
	}
	return result
}

func add(a, b map[rune]int) map[rune]int {
	result := make(map[rune]int)
	for r := 'a'; r <= 'z'; r++ {
		result[r] = a[r] + b[r]
	}
	return result
}

func remaining(counts map[rune]int) int {
	total := 0
	for _, c := range counts {
		if c > 0 {
			total += c
		}
	}
	return total
}

func mismatch(a, b map[rune]int) int {
	diff := 0
	for r := 'a'; r <= 'z'; r++ {
		d := a[r] - b[r]
		if d < 0 {
			d = -d
		}
		diff += d
	}
	return diff
}

type Solution struct {
	words  []int
	counts map[rune]int
	score  float64
}

func (s *Solution) clone() *Solution {
	newWords := make([]int, len(s.words))
	copy(newWords, s.words)
	newCounts := make(map[rune]int)
	for r, c := range s.counts {
		newCounts[r] = c
	}
	return &Solution{words: newWords, counts: newCounts, score: s.score}
}

func scoreSol(indices []int, counts map[rune]int) float64 {
	mis := mismatch(counts, targetCount)

	// Frequency score
	freq := 0.0
	for _, idx := range indices {
		f := wordFreq[words[idx]]
		freq += float64(f)
	}
	freq /= float64(len(indices))

	// Word count (target ~16)
	wcDiff := len(indices) - 16
	if wcDiff < 0 {
		wcDiff = -wcDiff
	}

	return float64(mis)*1000 + freq*0.1 + float64(wcDiff)*10
}

func main() {
	wordFreq = make(map[string]int)

	// Load frequency data
	file, _ := os.Open("google-10000.txt")
	scanner := bufio.NewScanner(file)
	rank := 0
	for scanner.Scan() {
		w := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(w) >= 1 {
			wordFreq[w] = rank
			rank++
		}
	}
	file.Close()

	// Load large word list but only keep usable ones
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
		if canUseWord(targetCount, wc) {
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

	// Update wordCounts order
	newWordCounts := make([]map[rune]int, len(words))
	for i, w := range words {
		newWordCounts[i] = letterCount(w)
	}
	wordCounts = newWordCounts

	fmt.Printf("Loaded %d usable words\n", len(words))

	// Print some top words with rare letters
	fmt.Println("\nWords with rare letters (x, z, k, f):")
	count := 0
	for _, w := range words {
		hasRare := false
		for _, r := range w {
			if r == 'x' || r == 'z' || r == 'k' {
				hasRare = true
				break
			}
		}
		if hasRare && count < 30 {
			fmt.Printf("  %s (freq: %d)\n", w, wordFreq[w])
			count++
		}
	}

	// Run parallel search
	numWorkers := runtime.NumCPU()
	fmt.Printf("\nStarting %d parallel searchers\n", numWorkers)

	var mu sync.Mutex
	var bestSol *Solution
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))
			localBest := search(r, done)

			mu.Lock()
			if localBest != nil && (bestSol == nil || localBest.score < bestSol.score) {
				bestSol = localBest
			}
			mu.Unlock()
		}(i)
	}

	// Run for 60 seconds
	time.Sleep(60 * time.Second)
	close(done)
	wg.Wait()

	if bestSol != nil {
		ws := make([]string, len(bestSol.words))
		for i, idx := range bestSol.words {
			ws[i] = words[idx]
		}
		fmt.Printf("\n=== BEST SOLUTION ===\n")
		fmt.Printf("Mismatch: %d\n", mismatch(bestSol.counts, targetCount))
		fmt.Printf("Words: %d\n", len(bestSol.words))
		fmt.Printf("Solution: %s\n", strings.Join(ws, " "))
	}
}

func search(r *rand.Rand, done <-chan struct{}) *Solution {
	// Start with random common words
	current := randomSolution(r)
	best := current.clone()

	temp := 500.0
	cooling := 0.9999

	iter := 0
	for {
		select {
		case <-done:
			return best
		default:
		}

		// Mutate
		neighbor := mutate(current, r)

		delta := neighbor.score - current.score
		if delta < 0 || r.Float64() < temp/500.0 {
			current = neighbor
		}

		if current.score < best.score {
			best = current.clone()
			mis := mismatch(best.counts, targetCount)
			if mis <= 2 {
				ws := make([]string, len(best.words))
				for i, idx := range best.words {
					ws[i] = words[idx]
				}
				fmt.Printf("[iter %d] mismatch=%d words=%d: %s\n",
					iter, mis, len(best.words), strings.Join(ws, " "))
			}
		}

		temp *= cooling
		if temp < 1 {
			temp = 500.0
			current = randomSolution(r)
		}
		iter++
	}
}

func randomSolution(r *rand.Rand) *Solution {
	indices := make([]int, 0, 16)
	counts := make(map[rune]int)

	// Pick 14-18 random common words
	numWords := 14 + r.Intn(5)
	for i := 0; i < numWords; i++ {
		idx := r.Intn(min(2000, len(words)))
		indices = append(indices, idx)
		counts = add(counts, wordCounts[idx])
	}

	return &Solution{
		words:  indices,
		counts: counts,
		score:  scoreSol(indices, counts),
	}
}

func mutate(sol *Solution, r *rand.Rand) *Solution {
	newIndices := make([]int, len(sol.words))
	copy(newIndices, sol.words)

	op := r.Intn(100)

	if op < 70 && len(newIndices) > 0 {
		// Replace a word
		replaceIdx := r.Intn(len(newIndices))
		newIdx := r.Intn(min(3000, len(words)))
		newIndices[replaceIdx] = newIdx
	} else if op < 85 {
		// Add a word
		newIdx := r.Intn(min(2000, len(words)))
		newIndices = append(newIndices, newIdx)
	} else if op < 100 && len(newIndices) > 10 {
		// Remove a word
		removeIdx := r.Intn(len(newIndices))
		newIndices = append(newIndices[:removeIdx], newIndices[removeIdx+1:]...)
	}

	newCounts := make(map[rune]int)
	for _, idx := range newIndices {
		newCounts = add(newCounts, wordCounts[idx])
	}

	return &Solution{
		words:  newIndices,
		counts: newCounts,
		score:  scoreSol(newIndices, newCounts),
	}
}
