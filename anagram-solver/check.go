package main

import (
	"fmt"
	"sort"
	"strings"
)

func sortLetters(s string) string {
	s = strings.ToLower(s)
	var letters []rune
	for _, r := range s {
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
		"the important thing is not to stop questioning",
		"imagination is more important than knowledge",
		"nothing is particularly hard if you break it into small tasks",
		"simplicity is the ultimate sophistication",
		"talk is cheap show me the code",
		"programming is thinking not typing",
		"first do it then do it right then do it better",
		"any fool can use a computer most fools do",
		"debugging is like hunting a bug in a haystack",
		"optimizing too early is the root of all programming sin",
		"multiprogramming is the spitting image of multitasking",
	}
	
	fmt.Printf("Target letters: %s\n", target)
	fmt.Printf("Target length: %d\n\n", len(target))
	
	for _, q := range quotes {
		sorted := sortLetters(q)
		match := sorted == target
		fmt.Printf("Quote: %s\n", q)
		fmt.Printf("  Sorted: %s\n", sorted)
		fmt.Printf("  Length: %d, Match: %v\n\n", len(sorted), match)
	}
}
