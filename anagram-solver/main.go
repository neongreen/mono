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

// Target letters (sorted)
const targetLetters = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"
const targetSpaces = 15

var targetCounts map[rune]int

type Word struct {
	text      string
	letters   string // sorted letters
	frequency int    // rank (lower = more common)
	counts    map[rune]int
}

type Solution struct {
	words    []int // indices into wordList
	score    float64
	letters  string
	mismatch int
}

func (s *Solution) clone() *Solution {
	newWords := make([]int, len(s.words))
	copy(newWords, s.words)
	return &Solution{
		words:    newWords,
		score:    s.score,
		letters:  s.letters,
		mismatch: s.mismatch,
	}
}

var wordList []Word
var wordsByLen map[int][]int

// Common short words that are allowed
var allowedShortWords = map[string]bool{
	"a": true, "i": true, "an": true, "as": true, "at": true, "be": true,
	"by": true, "do": true, "go": true, "he": true, "if": true, "in": true,
	"is": true, "it": true, "me": true, "my": true, "no": true, "of": true,
	"on": true, "or": true, "so": true, "to": true, "up": true, "us": true,
	"we": true, "am": true, "ok": true,
}

func init() {
	targetCounts = letterCounts(targetLetters)
}

func sortLetters(s string) string {
	s = strings.ToLower(s)
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	return string(runes)
}

func letterCounts(s string) map[rune]int {
	counts := make(map[rune]int)
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			counts[r]++
		}
	}
	return counts
}

// Check if word can possibly fit (uses only letters available in target)
func canFit(wordCounts map[rune]int) bool {
	for r, count := range wordCounts {
		if targetCounts[r] < count {
			return false
		}
	}
	return true
}

func loadWords(freqFile, largeFile string) error {
	// Load frequency-ranked words first
	freqWords := make(map[string]int)
	file, err := os.Open(freqFile)
	if err == nil {
		scanner := bufio.NewScanner(file)
		rank := 0
		for scanner.Scan() {
			text := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if len(text) > 0 {
				freqWords[text] = rank
				rank++
			}
		}
		file.Close()
	}

	// Load large word list
	file, err = os.Open(largeFile)
	if err != nil {
		return err
	}
	defer file.Close()

	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if len(text) == 0 || seen[text] {
			continue
		}
		seen[text] = true

		// Only keep words with letters a-z
		valid := true
		for _, r := range text {
			if r < 'a' || r > 'z' {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		// Filter out short words that aren't common
		if len(text) <= 2 && !allowedShortWords[text] {
			continue
		}

		counts := letterCounts(text)
		// Skip words that can't possibly fit
		if !canFit(counts) {
			continue
		}

		freq := 100000 // default for unknown words
		if f, ok := freqWords[text]; ok {
			freq = f
		}

		wordList = append(wordList, Word{
			text:      text,
			letters:   sortLetters(text),
			frequency: freq,
			counts:    counts,
		})
	}

	// Sort by frequency (most common first)
	sort.Slice(wordList, func(i, j int) bool {
		return wordList[i].frequency < wordList[j].frequency
	})

	// Re-assign frequency ranks
	for i := range wordList {
		wordList[i].frequency = i
	}

	// Group words by length
	wordsByLen = make(map[int][]int)
	for i, w := range wordList {
		wordsByLen[len(w.text)] = append(wordsByLen[len(w.text)], i)
	}

	return scanner.Err()
}

func letterMismatch(solCounts map[rune]int) int {
	mismatch := 0
	allLetters := make(map[rune]bool)
	for r := range targetCounts {
		allLetters[r] = true
	}
	for r := range solCounts {
		allLetters[r] = true
	}

	for r := range allLetters {
		diff := targetCounts[r] - solCounts[r]
		if diff < 0 {
			diff = -diff
		}
		mismatch += diff
	}
	return mismatch
}

func combinedCounts(indices []int) map[rune]int {
	counts := make(map[rune]int)
	for _, idx := range indices {
		for r, c := range wordList[idx].counts {
			counts[r] += c
		}
	}
	return counts
}

func combinedLetters(counts map[rune]int) string {
	var sb strings.Builder
	for r := 'a'; r <= 'z'; r++ {
		for i := 0; i < counts[r]; i++ {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func scoreSolution(indices []int, counts map[rune]int) (float64, int, string) {
	mismatch := letterMismatch(counts)
	letters := combinedLetters(counts)

	totalLen := 0
	for r, c := range counts {
		if r >= 'a' && r <= 'z' {
			totalLen += c
		}
	}

	// Frequency score: VERY HEAVILY prefer common words (top 2000)
	freqScore := 0.0
	for _, idx := range indices {
		freq := wordList[idx].frequency
		freqScore += math.Log(float64(freq + 1))
		// Heavily penalize uncommon words
		if freq > 2000 {
			freqScore += float64(freq-2000) / 50
		}
		if freq > 5000 {
			freqScore += float64(freq-5000) / 20
		}
	}

	// Word count penalty - target exactly 16 words
	wordCountDiff := len(indices) - (targetSpaces + 1)
	if wordCountDiff < 0 {
		wordCountDiff = -wordCountDiff
	}

	// Penalize very short average word length
	avgWordLen := float64(totalLen) / float64(len(indices))
	shortWordPenalty := 0.0
	if avgWordLen < 5 {
		shortWordPenalty = (5 - avgWordLen) * 30
	}

	// Primary: letter match, Secondary: word count, Tertiary: frequency
	score := float64(mismatch)*1000 + float64(wordCountDiff)*100 + freqScore*2 + shortWordPenalty

	return score, mismatch, letters
}

func createRandomSolution(r *rand.Rand) *Solution {
	numWords := targetSpaces + 1 + r.Intn(3) - 1
	targetLen := len(targetLetters)
	avgLen := targetLen / numWords

	indices := make([]int, 0, numWords)
	totalLen := 0

	for i := 0; i < numWords && totalLen < targetLen+10; i++ {
		targetWordLen := avgLen + r.Intn(5) - 2
		if targetWordLen < 3 {
			targetWordLen = 3
		}
		if targetWordLen > 12 {
			targetWordLen = 12
		}

		candidates := wordsByLen[targetWordLen]
		if len(candidates) == 0 {
			for delta := 1; delta <= 5; delta++ {
				candidates = wordsByLen[targetWordLen+delta]
				if len(candidates) > 0 {
					break
				}
				candidates = wordsByLen[targetWordLen-delta]
				if len(candidates) > 0 {
					break
				}
			}
		}

		if len(candidates) > 0 {
			// Prefer more common words
			maxIdx := min(200, len(candidates))
			idx := candidates[r.Intn(maxIdx)]
			indices = append(indices, idx)
			totalLen += len(wordList[idx].text)
		}
	}

	counts := combinedCounts(indices)
	score, mismatch, letters := scoreSolution(indices, counts)
	return &Solution{
		words:    indices,
		score:    score,
		mismatch: mismatch,
		letters:  letters,
	}
}

func mutateSolution(sol *Solution, r *rand.Rand) *Solution {
	newIndices := make([]int, len(sol.words))
	copy(newIndices, sol.words)

	mutationType := r.Intn(100)

	// Heavily favor common words (top 3000)
	commonLimit := min(3000, len(wordList))

	if mutationType < 60 {
		// Replace one word with another of similar length from common words
		if len(newIndices) > 0 {
			replaceIdx := r.Intn(len(newIndices))
			oldWord := wordList[newIndices[replaceIdx]]
			targetLen := len(oldWord.text) + r.Intn(3) - 1
			if targetLen < 2 {
				targetLen = 2
			}

			candidates := wordsByLen[targetLen]
			if len(candidates) > 0 {
				// Only use top common words by length
				maxIdx := min(300, len(candidates))
				newIndices[replaceIdx] = candidates[r.Intn(maxIdx)]
			}
		}
	} else if mutationType < 80 {
		// Replace one word with any common word
		if len(newIndices) > 0 {
			replaceIdx := r.Intn(len(newIndices))
			newIndices[replaceIdx] = r.Intn(commonLimit)
		}
	} else if mutationType < 90 {
		// Add a common word
		newWordIdx := r.Intn(commonLimit)
		newIndices = append(newIndices, newWordIdx)
	} else if mutationType < 98 {
		// Remove a word
		if len(newIndices) > 8 {
			removeIdx := r.Intn(len(newIndices))
			newIndices = append(newIndices[:removeIdx], newIndices[removeIdx+1:]...)
		}
	} else {
		// Swap two words with common ones
		if len(newIndices) >= 2 {
			idx1 := r.Intn(len(newIndices))
			idx2 := r.Intn(len(newIndices))
			newIndices[idx1] = r.Intn(commonLimit)
			newIndices[idx2] = r.Intn(commonLimit)
		}
	}

	counts := combinedCounts(newIndices)
	score, mismatch, letters := scoreSolution(newIndices, counts)
	return &Solution{
		words:    newIndices,
		score:    score,
		mismatch: mismatch,
		letters:  letters,
	}
}

func solutionText(sol *Solution) string {
	words := make([]string, len(sol.words))
	for i, idx := range sol.words {
		words[i] = wordList[idx].text
	}
	return strings.Join(words, " ")
}

func runAnnealing(id int, wg *sync.WaitGroup, bestChan chan<- *Solution, done <-chan struct{}) {
	defer wg.Done()

	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id*1000)))

	initialTemp := 1000.0
	finalTemp := 0.01
	coolingRate := 0.99995

	current := createRandomSolution(r)
	best := current
	lastReportedScore := best.score

	temp := initialTemp
	iter := 0

	for {
		select {
		case <-done:
			return
		default:
		}

		neighbor := mutateSolution(current, r)

		delta := neighbor.score - current.score
		if delta < 0 || r.Float64() < math.Exp(-delta/temp) {
			current = neighbor
		}

		// Report improvements (significant improvements only to avoid flooding)
		if current.score < best.score {
			best = current.clone()
			// Report if significant improvement or perfect match
			if best.score < lastReportedScore-5 || (best.mismatch == 0 && lastReportedScore > best.score) {
				select {
				case bestChan <- best.clone():
					lastReportedScore = best.score
				default:
					// Channel full, skip reporting
				}
			}
		}

		temp *= coolingRate
		if temp < finalTemp {
			temp = initialTemp * 0.1 // reheat
			current = createRandomSolution(r)
		}

		iter++
		if iter%200000 == 0 && id == 0 {
			fmt.Printf("[worker %d, iter %dk, temp=%.1f] best: score=%.2f, mismatch=%d\n",
				id, iter/1000, temp, best.score, best.mismatch)
		}
	}
}

func main() {
	fmt.Println("Loading word lists...")
	if err := loadWords("google-10000.txt", "words_alpha.txt"); err != nil {
		fmt.Printf("Error loading words: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d usable words\n", len(wordList))
	fmt.Printf("Target: %d letters, ~%d words\n", len(targetLetters), targetSpaces+1)

	numWorkers := runtime.NumCPU()
	fmt.Printf("Starting %d parallel workers\n", numWorkers)

	bestChan := make(chan *Solution, 100)
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go runAnnealing(i, &wg, bestChan, done)
	}

	var globalBest *Solution
	startTime := time.Now()
	timeout := 90 * time.Second
	solutionCount := 0

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case sol := <-bestChan:
			if globalBest == nil || sol.score < globalBest.score {
				globalBest = sol
				solutionCount++
				fmt.Printf("\n[%.1fs] NEW BEST #%d: score=%.2f, mismatch=%d, words=%d\n",
					time.Since(startTime).Seconds(), solutionCount, sol.score, sol.mismatch, len(sol.words))
				fmt.Printf("  %s\n", solutionText(sol))
			}
		case <-ticker.C:
			if globalBest != nil {
				fmt.Printf("[%.1fs] Current best: score=%.2f, mismatch=%d, words=%d\n",
					time.Since(startTime).Seconds(), globalBest.score, globalBest.mismatch, len(globalBest.words))
			}
		case <-timeoutCh:
			fmt.Println("\nTimeout reached - stopping workers")
			close(done)
			goto finish
		}
	}

finish:
	wg.Wait()

	if globalBest != nil {
		fmt.Println("\n=== BEST SOLUTION ===")
		fmt.Printf("Score: %.2f\n", globalBest.score)
		fmt.Printf("Letter mismatch: %d\n", globalBest.mismatch)
		fmt.Printf("Words: %d (target: %d)\n", len(globalBest.words), targetSpaces+1)
		fmt.Printf("Solution: %s\n", solutionText(globalBest))
		fmt.Printf("Letters: %s\n", globalBest.letters)
		fmt.Printf("Target:  %s\n", targetLetters)
	}
}
