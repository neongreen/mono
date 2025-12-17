package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Target letters (sorted)
const target = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"

var targetCount map[rune]int
var words []string
var wordCounts []map[rune]int
var wordFreq map[string]int
var commonPairs map[string]bool

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

func subtractWord(remaining, word map[rune]int) map[rune]int {
	result := make(map[rune]int)
	for r, c := range remaining {
		result[r] = c - word[r]
	}
	return result
}

func remainingLetters(counts map[rune]int) int {
	total := 0
	for _, c := range counts {
		total += c
	}
	return total
}

func loadCommonPairs() {
	// Common English word pairs/transitions
	pairs := []string{
		"the_", "_the", "of_", "_of", "and_", "_and", "to_", "_to", "a_", "_a",
		"in_", "_in", "is_", "_is", "it_", "_it", "that_", "_that",
		"for_", "_for", "you_", "_you", "on_", "_on", "with_", "_with",
		"this_", "_this", "be_", "_be", "are_", "_are", "as_", "_as",
		"at_", "_at", "by_", "_by", "from_", "_from", "or_", "_or",
		"an_", "_an", "not_", "_not", "but_", "_but", "all_", "_all",
		"will_", "_will", "have_", "_have", "your_", "_your",
		"programming_", "_programming", "thinking_", "_thinking",
		"something_", "_something", "nothing_", "_nothing",
		"anything_", "_anything", "everything_", "_everything",
	}
	commonPairs = make(map[string]bool)
	for _, p := range pairs {
		commonPairs[p] = true
	}
}

type State struct {
	words     []string
	remaining map[rune]int
	score     float64
}

func main() {
	loadCommonPairs()

	// Load words
	wordFreq = make(map[string]int)
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

	// Filter words that can fit
	for w := range wordFreq {
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
		}
	}

	// Sort by frequency
	sort.Slice(words, func(i, j int) bool {
		return wordFreq[words[i]] < wordFreq[words[j]]
	})

	fmt.Printf("Loaded %d usable words\n", len(words))
	fmt.Printf("Target: %d letters\n", len(target))

	// Just find perfect matches using common words
	// Try greedy + backtracking
	findSolutions(targetCount, nil, 0)
}

var bestSolutions [][]string

func findSolutions(remaining map[rune]int, current []string, depth int) {
	rem := remainingLetters(remaining)
	if rem == 0 {
		// Found a perfect match!
		sol := make([]string, len(current))
		copy(sol, current)
		bestSolutions = append(bestSolutions, sol)
		fmt.Printf("FOUND: %s\n", strings.Join(sol, " "))
		if len(bestSolutions) >= 10 {
			return
		}
	}

	if depth > 20 || rem < 0 || len(bestSolutions) >= 10 {
		return
	}

	// Try adding words (prefer common ones)
	maxTry := min(500, len(words))
	for i := 0; i < maxTry; i++ {
		w := words[i]
		wc := wordCounts[i]
		if canUseWord(remaining, wc) {
			newRemaining := subtractWord(remaining, wc)
			newCurrent := append(current, w)
			findSolutions(newRemaining, newCurrent, depth+1)
		}
	}
}
