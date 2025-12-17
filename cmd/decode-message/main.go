package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	sortedLetters = "aaaaaaaaaabbcdeeefgggghhhhhiiiiiiiiiiiikllllllmmmmmnnnnnnnnnnoooooooppppprrrrssssssttttttttttuuuuuuuuxyyyyyz"
	numWords      = 16
	totalLetters  = 108
)

// Common English words - prioritize these heavily
var commonWords = map[string]int{
	// Single letters
	"i": 200, "a": 200,
	// Very common 2-letter words
	"am": 180, "an": 180, "as": 180, "at": 180, "be": 180, "by": 180, "do": 180, "go": 180,
	"he": 180, "if": 180, "in": 180, "is": 180, "it": 180, "me": 180, "my": 180, "no": 180, "of": 180, "on": 180,
	"or": 180, "so": 180, "to": 180, "up": 180, "us": 180, "we": 180,
	// Common 3-letter words
	"all": 160, "and": 160, "any": 160, "are": 160, "but": 160, "can": 160, "did": 160, "for": 160, "get": 160, "got": 160,
	"had": 160, "has": 160, "her": 160, "him": 160, "his": 160, "its": 160, "let": 160, "may": 160, "not": 160, "now": 160,
	"one": 160, "our": 160, "out": 160, "say": 160, "she": 160, "the": 160, "too": 160, "two": 160, "was": 160, "who": 160, "you": 160,
	// Common 4-letter words
	"your": 150, "they": 150, "this": 150, "that": 150, "with": 150, "have": 150, "will": 150, "from": 150,
	"been": 150, "just": 150, "only": 150, "over": 150, "such": 150, "take": 150, "into": 150, "than": 150,
	"them": 150, "then": 150, "some": 150, "what": 150, "when": 150, "make": 150, "like": 150, "time": 150,
	"very": 150, "know": 150, "want": 150, "give": 150, "most": 150, "also": 150, "back": 150, "come": 150,
	"tell": 150, "guys": 150, "must": 150,
	// Common 5-letter words
	"about": 140, "after": 140, "first": 140, "think": 140, "could": 140, "would": 140, "there": 140, "their": 140,
	"which": 140, "these": 140, "other": 140, "being": 140, "those": 140, "still": 140, "while": 140, "where": 140,
	"since": 140, "under": 140, "right": 140, "never": 140, "every": 140, "going": 140, "might": 140,
	"trust": 140, "truly": 140, "story": 140, "spill": 140,
	// Common 6-letter words
	"should": 130, "people": 130, "before": 130, "really": 130, "always": 130, "things": 130, "little": 130,
	"mixing": 130, "secret": 130, "stupid": 130,
	// Longer common words - likely candidates
	"tonight": 150, "promise": 140, "something": 130, "anything": 130, "everything": 130, "nothing": 130,
	"important": 120, "absolutely": 110, "humiliating": 200, "publicly": 150,
}

// SortedDict maps sorted letter signatures to original words
type SortedDict struct {
	signatures map[string][]string
	byLength   map[int][]string
	commonSigs map[string]int // sorted sig -> commonality score
}

func sortString(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	return string(runes)
}

func editDistance(s1, s2 string) int {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}
	if len(s2) == 0 {
		return len(s1)
	}

	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for i := range prev {
		prev[i] = i
	}

	for i := range s1 {
		curr[0] = i + 1
		for j := range s2 {
			cost := 0
			if s1[i] != s2[j] {
				cost = 1
			}
			curr[j+1] = min(curr[j]+1, min(prev[j+1]+1, prev[j]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[len(s2)]
}

func loadDictionary() *SortedDict {
	dict := &SortedDict{
		signatures: make(map[string][]string),
		byLength:   make(map[int][]string),
		commonSigs: make(map[string]int),
	}

	// First, add all common words
	for word, score := range commonWords {
		sig := sortString(word)
		dict.signatures[sig] = append(dict.signatures[sig], word)
		if score > dict.commonSigs[sig] {
			dict.commonSigs[sig] = score
		}
	}

	// Try to load from web
	resp, err := http.Get("https://raw.githubusercontent.com/dwyl/english-words/master/words_alpha.txt")
	if err != nil {
		fmt.Println("Failed to download dictionary, using local")
		file, err := os.Open("/usr/share/dict/words")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if len(word) >= 1 && len(word) <= 20 && isAlpha(word) {
				sig := sortString(word)
				dict.signatures[sig] = append(dict.signatures[sig], word)
			}
		}
	} else {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			word := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if len(word) >= 1 && len(word) <= 20 && isAlpha(word) {
				sig := sortString(word)
				dict.signatures[sig] = append(dict.signatures[sig], word)
			}
		}
	}

	// Index by length
	seen := make(map[string]bool)
	for sig := range dict.signatures {
		if !seen[sig] {
			dict.byLength[len(sig)] = append(dict.byLength[len(sig)], sig)
			seen[sig] = true
		}
	}

	// Sort words in each signature group by commonality
	for sig, words := range dict.signatures {
		sort.Slice(words, func(i, j int) bool {
			scoreI := commonWords[words[i]]
			scoreJ := commonWords[words[j]]
			if scoreI != scoreJ {
				return scoreI > scoreJ
			}
			return len(words[i]) < len(words[j])
		})
		dict.signatures[sig] = words
	}

	return dict
}

func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func (d *SortedDict) findNearest(sortedSig string) (int, []string, int) {
	// Returns: (edit_distance, words, commonality_bonus)
	// Find the dictionary word whose sorted signature is closest to this partition

	// Exact match?
	if words, ok := d.signatures[sortedSig]; ok {
		bonus := d.commonSigs[sortedSig]
		return 0, words, bonus
	}

	// Search for closest match
	bestDist := len(sortedSig) + 10 // Default: very bad
	var bestWords []string
	bestBonus := 0
	targetLen := len(sortedSig)

	// Search signatures of similar length
	for offset := 0; offset <= min(targetLen+5, 15); offset++ {
		for _, length := range []int{targetLen - offset, targetLen + offset} {
			if length < 1 || length > 20 {
				continue
			}
			sigs, ok := d.byLength[length]
			if !ok {
				continue
			}

			for _, sig := range sigs {
				dist := editDistance(sortedSig, sig)
				bonus := d.commonSigs[sig]

				// Prefer lower edit distance, break ties with commonality
				if dist < bestDist || (dist == bestDist && bonus > bestBonus) {
					bestDist = dist
					bestWords = d.signatures[sig]
					bestBonus = bonus
				}

				if bestDist == 0 {
					return bestDist, bestWords, bestBonus
				}
			}
		}
		// Early exit if we found a good match
		if bestDist <= 2 {
			break
		}
	}

	return bestDist, bestWords, bestBonus
}

type WordMatch struct {
	Signature string
	Distance  int      // edit distance to nearest dictionary word (0 = exact match)
	Words     []string // matching or nearest dictionary words
	Bonus     int      // commonality bonus
}

type Solution struct {
	Splits  []int
	Error   int
	Matches []WordMatch
}

func partitionToSigs(letters string, splits []int) []string {
	sigs := make([]string, 0, len(splits)+1)
	start := 0
	for _, sp := range splits {
		if sp > start {
			sigs = append(sigs, letters[start:sp])
		}
		start = sp
	}
	if start < len(letters) {
		sigs = append(sigs, letters[start:])
	}
	return sigs
}

func scoreSolution(dict *SortedDict, letters string, splits []int) (int, []WordMatch) {
	sigs := partitionToSigs(letters, splits)

	// Verify all letters are used
	totalLettersUsed := 0
	for _, sig := range sigs {
		totalLettersUsed += len(sig)
	}
	if totalLettersUsed != len(letters) {
		return 1000000, nil // Invalid partition
	}

	// Score = sum of edit distances (lower is better)
	// Subtract commonality bonus to prefer common words
	score := 0
	matches := make([]WordMatch, len(sigs))

	for i, sig := range sigs {
		dist, words, bonus := dict.findNearest(sig)
		matches[i] = WordMatch{Signature: sig, Distance: dist, Words: words, Bonus: bonus}

		// Add edit distance (weighted heavily)
		score += dist * 100

		// Subtract bonus for common words (prefer common words at same distance)
		score -= bonus
	}

	return score, matches
}

func randomSplits(length, numSplits int) []int {
	splits := make(map[int]bool)
	for len(splits) < numSplits {
		p := rand.Intn(length-1) + 1
		splits[p] = true
	}

	result := make([]int, 0, numSplits)
	for p := range splits {
		result = append(result, p)
	}
	sort.Ints(result)
	return result
}

func mutateSplits(splits []int, length int, temperature float64) []int {
	newSplits := make([]int, len(splits))
	copy(newSplits, splits)

	nMutations := max(1, int(temperature*5))

	for i := 0; i < nMutations; i++ {
		if len(newSplits) == 0 {
			break
		}

		r := rand.Float64()
		if r < 0.6 {
			// Move a split point
			idx := rand.Intn(len(newSplits))
			delta := rand.Intn(11) - 5
			newPos := newSplits[idx] + delta
			if newPos >= 1 && newPos < length {
				dup := false
				for j, s := range newSplits {
					if j != idx && s == newPos {
						dup = true
						break
					}
				}
				if !dup {
					newSplits[idx] = newPos
				}
			}
		} else if r < 0.8 && len(newSplits) >= 2 {
			// Swap two splits
			i := rand.Intn(len(newSplits))
			j := rand.Intn(len(newSplits))
			newSplits[i], newSplits[j] = newSplits[j], newSplits[i]
		}
	}

	sort.Ints(newSplits)

	// Remove duplicates
	unique := make([]int, 0, len(newSplits))
	seen := make(map[int]bool)
	for _, s := range newSplits {
		if !seen[s] {
			unique = append(unique, s)
			seen[s] = true
		}
	}
	newSplits = unique

	// Adjust count
	for len(newSplits) < numWords-1 {
		p := rand.Intn(length-1) + 1
		if !seen[p] {
			newSplits = append(newSplits, p)
			seen[p] = true
		}
	}
	for len(newSplits) > numWords-1 {
		idx := rand.Intn(len(newSplits))
		newSplits = append(newSplits[:idx], newSplits[idx+1:]...)
	}

	sort.Ints(newSplits)
	return newSplits
}

func runWorker(dict *SortedDict, letters string, resultChan chan<- Solution, stopped *atomic.Bool, wg *sync.WaitGroup) {
	defer wg.Done()

	splits := randomSplits(len(letters), numWords-1)
	currentError, currentMatches := scoreSolution(dict, letters, splits)

	bestError := currentError
	bestSplits := make([]int, len(splits))
	copy(bestSplits, splits)
	bestMatches := currentMatches

	maxIterations := 10000000
	initialTemp := 1.0
	finalTemp := 0.00001

	for iteration := 0; iteration < maxIterations; iteration++ {
		if stopped.Load() {
			break
		}

		temperature := initialTemp * math.Pow(finalTemp/initialTemp, float64(iteration)/float64(maxIterations))

		newSplits := mutateSplits(splits, len(letters), temperature)
		newError, newMatches := scoreSolution(dict, letters, newSplits)

		delta := float64(newError - currentError)
		if delta < 0 || rand.Float64() < math.Exp(-delta/(temperature*100+0.001)) {
			splits = newSplits
			currentError = newError
			currentMatches = newMatches

			if currentError < bestError {
				bestError = currentError
				bestSplits = make([]int, len(splits))
				copy(bestSplits, splits)
				bestMatches = currentMatches

				select {
				case resultChan <- Solution{Splits: bestSplits, Error: bestError, Matches: bestMatches}:
				default:
				}
			}
		}
	}

	select {
	case resultChan <- Solution{Splits: bestSplits, Error: bestError, Matches: bestMatches}:
	default:
	}
}

func printSolution(sol Solution) {
	fmt.Printf("\n=== Solution (score=%d) ===\n", sol.Error)
	words := make([]string, len(sol.Matches))
	exactMatches := 0
	totalDist := 0
	for i, m := range sol.Matches {
		totalDist += m.Distance
		if len(m.Words) > 0 {
			words[i] = m.Words[0]
		} else {
			words[i] = "???"
		}

		if m.Distance == 0 {
			exactMatches++
			commonMark := ""
			if m.Bonus > 0 {
				commonMark = fmt.Sprintf(" [common:%d]", m.Bonus)
			}
			fmt.Printf("  '%s' -> '%s' (EXACT%s)\n", m.Signature, words[i], commonMark)
		} else {
			commonMark := ""
			if m.Bonus > 0 {
				commonMark = fmt.Sprintf(" [common:%d]", m.Bonus)
			}
			fmt.Printf("  '%s' -> '%s' (dist=%d%s)\n", m.Signature, words[i], m.Distance, commonMark)
		}
	}
	fmt.Printf("\nExact matches: %d/%d\n", exactMatches, len(sol.Matches))
	fmt.Printf("Total edit distance: %d\n", totalDist)
	if exactMatches == len(sol.Matches) {
		fmt.Printf("*** PERFECT SOLUTION - ALL PARTITIONS ARE EXACT MATCHES! ***\n")
	}
	fmt.Printf("Words: %s\n", strings.Join(words, " "))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Loading dictionary...")
	dict := loadDictionary()
	fmt.Printf("Loaded %d unique signatures\n", len(dict.signatures))

	letters := sortedLetters
	fmt.Printf("\nSearching for best partition of %d letters into %d words...\n", len(letters), numWords)

	numWorkers := runtime.NumCPU()
	fmt.Printf("Using %d parallel workers\n\n", numWorkers)

	resultChan := make(chan Solution, numWorkers*100)
	stopped := &atomic.Bool{}
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go runWorker(dict, letters, resultChan, stopped, &wg)
	}

	// Collect results and print progress
	var bestSolution Solution
	bestSolution.Error = math.MaxInt

	startTime := time.Now()
	checkpoints := []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second, 60 * time.Second, 120 * time.Second}
	checkpointIdx := 0
	maxRuntime := 180 * time.Second

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

loop:
	for {
		select {
		case sol := <-resultChan:
			if sol.Error < bestSolution.Error {
				bestSolution = sol
				elapsed := time.Since(startTime)
				fmt.Printf("\n[%.1fs] New best: score=%d\n", elapsed.Seconds(), sol.Error)
				printSolution(sol)
			}

		case <-ticker.C:
			elapsed := time.Since(startTime)

			// Print checkpoint
			if checkpointIdx < len(checkpoints) && elapsed >= checkpoints[checkpointIdx] {
				fmt.Printf("\n========== CHECKPOINT at %.0f seconds ==========\n", checkpoints[checkpointIdx].Seconds())
				if bestSolution.Error < math.MaxInt {
					printSolution(bestSolution)
				} else {
					fmt.Println("No solution found yet")
				}
				fmt.Println("================================================\n")
				checkpointIdx++
			}

			if elapsed >= maxRuntime {
				stopped.Store(true)
				break loop
			}
		}
	}

	// Give workers time to finish
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("\n\n==================== FINAL RESULT ====================\n")
	fmt.Printf("Runtime: %.1f seconds\n", time.Since(startTime).Seconds())
	if bestSolution.Error < math.MaxInt {
		printSolution(bestSolution)
	} else {
		fmt.Println("No solution found")
	}

	wg.Wait()
}
