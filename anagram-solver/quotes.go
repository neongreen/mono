package main

import (
	"fmt"
	"sort"
	"strings"
)

func sortedLetters(s string) string {
	var letters []rune
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			letters = append(letters, r)
		}
	}
	sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
	return string(letters)
}

func main() {
	target := "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"

	quotes := []string{
		// Try longer sentences with high i, a, n, t counts
		"obtaining maximum utility from optimizing things is about maintaining simplicity in algorithms and using thoughtful design patterns",
		"programming optimizations multiply initialization simplicity but nothing is still zero in multiplication algorithms",
		"multiplication using optimization algorithms is a math programming skill that simplicity thinking naturally brings about",
		"alphabetization of multiply initialization programming is nothing but an optimizing algorithm simplicity thinking ability",
		"manipulation of initialization algorithms multiplying optimization is simply about programming utility thinking ability",
		"multiply nothing by optimization and initialization using algorithms simply brings about programming thinking ability",
		"multiply using initialization optimization algorithms is programming thinking about nothing but simplicity in ability",
		"obtaining utility multiplication initialization simplifying optimization algorithms thinking ability programming is about nothing",
		"optimization multiplying utility initialization simplifying programming algorithms is about thinking ability nothing",
		"utilizing optimization multiplication initialization algorithms simplifying programming is about ability thinking nothing",
	}

	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Target length: %d\n\n", len(target))

	for _, q := range quotes {
		sorted := sortedLetters(q)

		// Compare letter by letter
		targetCounts := make(map[rune]int)
		for _, r := range target {
			targetCounts[r]++
		}
		sortedCounts := make(map[rune]int)
		for _, r := range sorted {
			sortedCounts[r]++
		}

		diff := 0
		var missing, extra []string
		for r := 'a'; r <= 'z'; r++ {
			d := targetCounts[r] - sortedCounts[r]
			if d > 0 {
				missing = append(missing, fmt.Sprintf("%c:%d", r, d))
			} else if d < 0 {
				extra = append(extra, fmt.Sprintf("%c:%d", r, -d))
			}
			if d < 0 {
				d = -d
			}
			diff += d
		}

		spaces := strings.Count(q, " ")
		fmt.Printf("Quote: %s\n", q)
		fmt.Printf("  Length: %d, Spaces: %d, Diff: %d\n", len(sorted), spaces, diff)
		if diff > 0 {
			fmt.Printf("  Missing: %v\n", missing)
			fmt.Printf("  Extra: %v\n", extra)
		}
		if diff == 0 && spaces == 15 {
			fmt.Printf("  *** PERFECT MATCH! ***\n")
		}
		fmt.Println()
	}
}
