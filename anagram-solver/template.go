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
const targetSpaces = 15

var targetCounts map[rune]int
var words []string
var wordCounts []map[rune]int
var wordsByLen map[int][]int

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

func canFit(wc map[rune]int, remaining map[rune]int) bool {
	for r, c := range wc {
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

func remaining(counts map[rune]int) int {
	total := 0
	for _, c := range counts {
		if c > 0 {
			total += c
		}
	}
	return total
}

func loadWords() {
	// Load common words
	file, _ := os.Open("google-10000.txt")
	scanner := bufio.NewScanner(file)
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
		if canFit(wc, targetCounts) {
			words = append(words, w)
			wordCounts = append(wordCounts, wc)
		}
	}
	file.Close()

	// Group by length
	wordsByLen = make(map[int][]int)
	for i, w := range words {
		wordsByLen[len(w)] = append(wordsByLen[len(w)], i)
	}

	fmt.Printf("Loaded %d words\n", len(words))
}

// Find words that could fill remaining letters
func findWords(rem map[rune]int, maxWords int) [][]int {
	var results [][]int
	remCount := remaining(rem)

	if remCount == 0 {
		return [][]int{{}}
	}
	if maxWords == 0 {
		return nil
	}

	// Try words that fit
	for i, wc := range wordCounts {
		if !canFit(wc, rem) {
			continue
		}
		wLen := len(words[i])
		if wLen > remCount {
			continue
		}

		newRem := subtract(rem, wc)
		subResults := findWords(newRem, maxWords-1)
		for _, sub := range subResults {
			results = append(results, append([]int{i}, sub...))
			if len(results) > 100 {
				return results
			}
		}
	}
	return results
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

type Result struct {
	words []string
	score int
}

func worker(id int, templates [][]string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

	for _, template := range templates {
		// Calculate used letters from template
		usedCounts := make(map[rune]int)
		for _, w := range template {
			for _, r := range strings.ToLower(w) {
				if r >= 'a' && r <= 'z' {
					usedCounts[r]++
				}
			}
		}

		// Check if template fits
		if !canFit(usedCounts, targetCounts) {
			continue
		}

		rem := subtract(targetCounts, usedCounts)
		remLetters := remaining(rem)
		neededWords := targetSpaces + 1 - len(template)

		if neededWords <= 0 || remLetters <= 0 {
			continue
		}

		// Try to find words to fill the gap
		avgLen := remLetters / neededWords

		// Try random combinations
		for attempt := 0; attempt < 1000; attempt++ {
			testRem := make(map[rune]int)
			for k, v := range rem {
				testRem[k] = v
			}

			fillers := []string{}
			for i := 0; i < neededWords && remaining(testRem) > 0; i++ {
				// Pick a random word that fits
				targetLen := avgLen + r.Intn(5) - 2
				if targetLen < 2 {
					targetLen = 2
				}
				candidates := wordsByLen[targetLen]
				if len(candidates) == 0 {
					continue
				}
				found := false
				for tries := 0; tries < 50 && !found; tries++ {
					idx := candidates[r.Intn(len(candidates))]
					wc := wordCounts[idx]
					if canFit(wc, testRem) {
						fillers = append(fillers, words[idx])
						testRem = subtract(testRem, wc)
						found = true
					}
				}
			}

			mis := mismatch(subtract(targetCounts, testRem))
			if mis == 0 && len(template)+len(fillers) == targetSpaces+1 {
				allWords := append(append([]string{}, template...), fillers...)
				results <- Result{words: allWords, score: 0}
			}
		}
	}
}

func main() {
	loadWords()

	// Templates with common sentence starters
	templates := [][]string{
		{"the"},
		{"a"},
		{"it", "is"},
		{"this", "is"},
		{"in", "the"},
		{"to", "the"},
		{"for", "the"},
		{"if", "you"},
		{"is", "not"},
		{"the", "best"},
		{"nothing", "is"},
		{"anything", "is"},
		{"programming", "is"},
		{"optimization", "is"},
		{"think", "about"},
		{"in", "terms", "of"},
		{"as", "a"},
		{"about", "the"},
	}

	results := make(chan Result, 100)
	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	chunkSize := len(templates) / numWorkers
	if chunkSize < 1 {
		chunkSize = 1
	}

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = len(templates)
		}
		if start >= len(templates) {
			break
		}
		wg.Add(1)
		go worker(i, templates[start:end], results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	seen := make(map[string]bool)
	count := 0
	for res := range results {
		// Sort words to detect duplicates
		sorted := make([]string, len(res.words))
		copy(sorted, res.words)
		sort.Strings(sorted)
		key := strings.Join(sorted, " ")

		if !seen[key] {
			seen[key] = true
			count++
			fmt.Printf("Found #%d: %s\n", count, strings.Join(res.words, " "))
		}
	}

	fmt.Printf("\nTotal unique solutions: %d\n", count)
}
