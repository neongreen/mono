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

// Target anagram with "hexa" as the first word
const targetLetters = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"
const targetSpaces = 15
const firstWord = "hexa"

var targetCounts map[rune]int

// Word represents a word with its properties
type Word struct {
	text      string
	letters   string
	frequency int // rank (lower = more common)
	counts    map[rune]int
	posTag    string // simplified POS tag (noun, verb, adj, adv, etc.)
}

// BigramScore stores bigram frequency data
type BigramScore map[string]float64

// TrigramScore stores trigram frequency data
type TrigramScore map[string]float64

// Solution represents a candidate anagram solution
type Solution struct {
	words       []int // indices into wordList (first is always "hexa")
	score       float64
	letterScore float64 // letter mismatch score
	bigramScore float64 // bigram fluency score
	trigramScore float64 // trigram fluency score
	posScore    float64 // POS sequence score
	mismatch    int
}

var wordList []Word
var bigramScores BigramScore
var trigramScores TrigramScore
var firstWordIndex int

func init() {
	targetCounts = letterCounts(targetLetters)
	bigramScores = make(BigramScore)
	trigramScores = make(TrigramScore)
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

func canFit(wordCounts map[rune]int) bool {
	for r, count := range wordCounts {
		if targetCounts[r] < count {
			return false
		}
	}
	return true
}

// Simple POS tagger based on word endings and common patterns
func inferPOSTag(word string) string {
	word = strings.ToLower(word)
	
	// Common verbs
	verbs := []string{"is", "are", "was", "were", "be", "been", "being", "have", "has", "had", "do", "does", "did",
		"can", "could", "will", "would", "should", "may", "might", "must"}
	for _, v := range verbs {
		if word == v {
			return "VERB"
		}
	}
	
	// Articles and determiners
	if word == "the" || word == "a" || word == "an" {
		return "DET"
	}
	
	// Pronouns
	pronouns := []string{"i", "you", "he", "she", "it", "we", "they", "me", "him", "her", "us", "them"}
	for _, p := range pronouns {
		if word == p {
			return "PRON"
		}
	}
	
	// Prepositions
	preps := []string{"in", "on", "at", "to", "for", "of", "with", "from", "by", "about", "through"}
	for _, p := range preps {
		if word == p {
			return "ADP"
		}
	}
	
	// Adjectives (common endings)
	if strings.HasSuffix(word, "ful") || strings.HasSuffix(word, "less") || 
	   strings.HasSuffix(word, "ous") || strings.HasSuffix(word, "ive") || 
	   strings.HasSuffix(word, "able") || strings.HasSuffix(word, "ible") {
		return "ADJ"
	}
	
	// Adverbs (common ending)
	if strings.HasSuffix(word, "ly") && len(word) > 3 {
		return "ADV"
	}
	
	// Verbs (common endings)
	if strings.HasSuffix(word, "ing") || strings.HasSuffix(word, "ed") || 
	   strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") {
		return "VERB"
	}
	
	// Default to NOUN
	return "NOUN"
}

func loadWords(freqFile string) error {
	// First, manually add "hexa" since it's a name and not in the word list
	hexaCounts := letterCounts(firstWord)
	if canFit(hexaCounts) {
		wordList = append(wordList, Word{
			text:      firstWord,
			letters:   sortLetters(firstWord),
			frequency: 0, // Give it highest priority
			counts:    hexaCounts,
			posTag:    "NOUN", // It's a name
		})
		firstWordIndex = 0
	}
	
	file, err := os.Open(freqFile)
	if err != nil {
		return err
	}
	defer file.Close()

	seen := make(map[string]bool)
	seen[firstWord] = true // Don't add hexa again
	scanner := bufio.NewScanner(file)
	rank := 1 // Start from 1 since hexa is 0
	
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

		// Skip very short words except common ones
		if len(text) == 1 && text != "a" && text != "i" {
			continue
		}

		counts := letterCounts(text)
		if !canFit(counts) {
			continue
		}

		wordList = append(wordList, Word{
			text:      text,
			letters:   sortLetters(text),
			frequency: rank,
			counts:    counts,
			posTag:    inferPOSTag(text),
		})
		
		// Store index of first word "hexa"
		if text == firstWord {
			firstWordIndex = len(wordList) - 1
		}
		
		rank++
		if rank >= 5000 { // Limit to top 5000 words
			break
		}
	}

	return scanner.Err()
}

// Load bigram frequencies from en_freq.txt
func loadBigrams() error {
	// For simplicity, we'll generate bigram scores from word frequency
	// In a real implementation, we'd use actual bigram corpus data
	for i := 0; i < len(wordList) && i < 1000; i++ {
		for j := 0; j < len(wordList) && j < 1000; j++ {
			w1, w2 := wordList[i].text, wordList[j].text
			// Score based on both words' frequencies
			score := 1.0 / (1.0 + math.Log(float64(i+j+2)))
			bigramScores[w1+" "+w2] = score
		}
	}
	return nil
}

// Load trigram frequencies
func loadTrigrams() error {
	// For simplicity, we'll generate trigram scores from word frequency
	// In a real implementation, we'd use actual trigram corpus data
	for i := 0; i < len(wordList) && i < 500; i++ {
		for j := 0; j < len(wordList) && j < 500; j++ {
			for k := 0; k < len(wordList) && k < 500; k++ {
				w1, w2, w3 := wordList[i].text, wordList[j].text, wordList[k].text
				score := 1.0 / (1.0 + math.Log(float64(i+j+k+3)))
				trigramScores[w1+" "+w2+" "+w3] = score
			}
		}
	}
	return nil
}

// Calculate bigram score for a sequence of words
func calculateBigramScore(words []string) float64 {
	if len(words) < 2 {
		return 0.0
	}
	
	score := 0.0
	for i := 0; i < len(words)-1; i++ {
		bigram := words[i] + " " + words[i+1]
		if s, ok := bigramScores[bigram]; ok {
			score += s
		} else {
			// Penalize unknown bigrams
			score -= 0.1
		}
	}
	return score / float64(len(words)-1) // Average score
}

// Calculate trigram score for a sequence of words
func calculateTrigramScore(words []string) float64 {
	if len(words) < 3 {
		return 0.0
	}
	
	score := 0.0
	count := 0
	for i := 0; i < len(words)-2; i++ {
		trigram := words[i] + " " + words[i+1] + " " + words[i+2]
		if s, ok := trigramScores[trigram]; ok {
			score += s
			count++
		}
	}
	if count == 0 {
		return -0.1
	}
	return score / float64(count)
}

// Calculate POS sequence score
func calculatePOSScore(words []string) float64 {
	if len(words) < 2 {
		return 0.0
	}
	
	score := 0.0
	tags := make([]string, len(words))
	for i, w := range words {
		tags[i] = inferPOSTag(w)
	}
	
	// Reward common POS sequences
	for i := 0; i < len(tags)-1; i++ {
		// Common patterns get positive scores
		pair := tags[i] + "-" + tags[i+1]
		switch pair {
		case "DET-NOUN", "DET-ADJ", "ADJ-NOUN", "NOUN-VERB", 
		     "ADP-DET", "ADP-NOUN", "VERB-ADV", "ADV-VERB":
			score += 0.5
		case "VERB-VERB", "DET-VERB", "NOUN-DET":
			score -= 0.3 // Penalize unlikely sequences
		}
	}
	
	return score / float64(len(words)-1)
}

func letterMismatch(solCounts map[rune]int) int {
	mismatch := 0
	for r := 'a'; r <= 'z'; r++ {
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

func scoreSolution(indices []int) *Solution {
	counts := combinedCounts(indices)
	mismatch := letterMismatch(counts)
	
	// Get word texts
	words := make([]string, len(indices))
	for i, idx := range indices {
		words[i] = wordList[idx].text
	}
	
	// Calculate individual scores
	bigramScore := calculateBigramScore(words)
	trigramScore := calculateTrigramScore(words)
	posScore := calculatePOSScore(words)
	
	// Frequency score (prefer common words)
	freqScore := 0.0
	for _, idx := range indices {
		freq := wordList[idx].frequency
		freqScore += math.Log(float64(freq + 1))
		if freq > 2000 {
			freqScore += float64(freq-2000) / 50
		}
	}
	
	// Word count penalty - target exactly 16 words
	wordCountDiff := len(indices) - (targetSpaces + 1)
	if wordCountDiff < 0 {
		wordCountDiff = -wordCountDiff
	}
	
	// Combined score with weights
	letterScore := float64(mismatch) * 1000
	totalScore := letterScore + 
	             float64(wordCountDiff)*100 + 
	             freqScore*2 - 
	             bigramScore*50 -  // Negative because higher bigram score is better
	             trigramScore*30 - 
	             posScore*20
	
	return &Solution{
		words:        indices,
		score:        totalScore,
		letterScore:  letterScore,
		bigramScore:  bigramScore,
		trigramScore: trigramScore,
		posScore:     posScore,
		mismatch:     mismatch,
	}
}

func createInitialSolution(r *rand.Rand) *Solution {
	// Start with "hexa" as first word
	numWords := targetSpaces + 1
	indices := make([]int, numWords)
	indices[0] = firstWordIndex // "hexa"
	
	// Fill remaining words randomly from common words
	for i := 1; i < numWords; i++ {
		indices[i] = r.Intn(min(2000, len(wordList)))
	}
	
	return scoreSolution(indices)
}

func mutateSolution(sol *Solution, r *rand.Rand) *Solution {
	newIndices := make([]int, len(sol.words))
	copy(newIndices, sol.words)
	
	mutationType := r.Intn(100)
	
	// Never mutate the first word (it's always "hexa")
	if mutationType < 60 && len(newIndices) > 1 {
		// Replace one word (not the first)
		replaceIdx := 1 + r.Intn(len(newIndices)-1)
		newIndices[replaceIdx] = r.Intn(min(3000, len(wordList)))
	} else if mutationType < 80 && len(newIndices) > 1 {
		// Swap two words (not the first)
		if len(newIndices) > 2 {
			idx1 := 1 + r.Intn(len(newIndices)-1)
			idx2 := 1 + r.Intn(len(newIndices)-1)
			newIndices[idx1], newIndices[idx2] = newIndices[idx2], newIndices[idx1]
		}
	} else if mutationType < 90 {
		// Add a common word (not at beginning)
		newWordIdx := r.Intn(min(2000, len(wordList)))
		insertPos := 1 + r.Intn(len(newIndices))
		newIndices = append(newIndices[:insertPos], append([]int{newWordIdx}, newIndices[insertPos:]...)...)
	} else if len(newIndices) > 10 {
		// Remove a word (not the first)
		removeIdx := 1 + r.Intn(len(newIndices)-1)
		newIndices = append(newIndices[:removeIdx], newIndices[removeIdx+1:]...)
	}
	
	return scoreSolution(newIndices)
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

	current := createInitialSolution(r)
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

		if current.score < best.score {
			best = current
			if best.score < lastReportedScore-5 || (best.mismatch == 0 && lastReportedScore > best.score) {
				select {
				case bestChan <- best:
					lastReportedScore = best.score
				default:
				}
			}
		}

		temp *= coolingRate
		if temp < finalTemp {
			temp = initialTemp * 0.1
			current = createInitialSolution(r)
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
	if err := loadWords("google-10000.txt"); err != nil {
		fmt.Printf("Error loading words: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d usable words\n", len(wordList))
	
	if firstWordIndex == 0 && wordList[0].text != firstWord {
		fmt.Printf("Error: Could not find first word '%s' in word list\n", firstWord)
		return
	}
	fmt.Printf("First word '%s' is at index %d\n", firstWord, firstWordIndex)
	
	fmt.Println("Loading bigram data...")
	if err := loadBigrams(); err != nil {
		fmt.Printf("Error loading bigrams: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d bigrams\n", len(bigramScores))
	
	fmt.Println("Loading trigram data...")
	if err := loadTrigrams(); err != nil {
		fmt.Printf("Error loading trigrams: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d trigrams\n", len(trigramScores))
	
	fmt.Printf("Target: %d letters, ~%d words, starting with '%s'\n", 
		len(targetLetters), targetSpaces+1, firstWord)

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
				fmt.Printf("  Text: %s\n", solutionText(sol))
				fmt.Printf("  Bigram: %.3f, Trigram: %.3f, POS: %.3f\n", 
					sol.bigramScore, sol.trigramScore, sol.posScore)
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
		fmt.Printf("Total score: %.2f\n", globalBest.score)
		fmt.Printf("  Letter mismatch: %d (weight: 1000x)\n", globalBest.mismatch)
		fmt.Printf("  Bigram score: %.3f (weight: -50x)\n", globalBest.bigramScore)
		fmt.Printf("  Trigram score: %.3f (weight: -30x)\n", globalBest.trigramScore)
		fmt.Printf("  POS score: %.3f (weight: -20x)\n", globalBest.posScore)
		fmt.Printf("Words: %d (target: %d)\n", len(globalBest.words), targetSpaces+1)
		fmt.Printf("\nSolution: %s\n", solutionText(globalBest))
		
		// Show POS tags
		words := make([]string, len(globalBest.words))
		tags := make([]string, len(globalBest.words))
		for i, idx := range globalBest.words {
			words[i] = wordList[idx].text
			tags[i] = wordList[idx].posTag
		}
		fmt.Printf("\nPOS tags: %s\n", strings.Join(tags, " "))
	}
}
