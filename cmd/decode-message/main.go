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

// wordFrequency stores frequency data loaded from external source
var wordFrequency = make(map[string]int)

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

func loadWordFrequencies() {
	fmt.Println("Loading word frequencies...")
	resp, err := http.Get("https://raw.githubusercontent.com/hermitdave/FrequencyWords/master/content/2018/en/en_50k.txt")
	if err != nil {
		fmt.Println("Warning: could not load word frequencies:", err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	rank := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			word := strings.ToLower(parts[0])
			if isAlpha(word) {
				// Score based on rank: top words get highest scores
				// Use inverse rank scaled to give meaningful bonuses
				// Top 100 words: 500-400, Top 1000: 400-300, Top 10000: 300-100, rest: 100-0
				var score int
				if rank < 100 {
					score = 500 - rank
				} else if rank < 1000 {
					score = 400 - (rank-100)/10
				} else if rank < 10000 {
					score = 300 - (rank-1000)/100
				} else {
					score = max(0, 100-(rank-10000)/500)
				}
				wordFrequency[word] = score
				rank++
			}
		}
	}
	fmt.Printf("Loaded %d word frequencies\n", len(wordFrequency))
}

func loadDictionary() *SortedDict {
	dict := &SortedDict{
		signatures: make(map[string][]string),
		byLength:   make(map[int][]string),
		commonSigs: make(map[string]int),
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
				// Update commonSigs with frequency score
				if freq := wordFrequency[word]; freq > dict.commonSigs[sig] {
					dict.commonSigs[sig] = freq
				}
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
				// Update commonSigs with frequency score
				if freq := wordFrequency[word]; freq > dict.commonSigs[sig] {
					dict.commonSigs[sig] = freq
				}
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

	// Sort words in each signature group by frequency (highest first)
	for sig, words := range dict.signatures {
		sort.Slice(words, func(i, j int) bool {
			scoreI := wordFrequency[words[i]]
			scoreJ := wordFrequency[words[j]]
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
	// Add penalty for obscure/unknown words
	score := 0
	matches := make([]WordMatch, len(sigs))

	for i, sig := range sigs {
		dist, words, bonus := dict.findNearest(sig)
		matches[i] = WordMatch{Signature: sig, Distance: dist, Words: words, Bonus: bonus}

		// Add edit distance (weighted heavily)
		score += dist * 100

		// Subtract bonus for common words (prefer common words at same distance)
		score -= bonus

		// PENALTY: Words not in frequency list (outside top 50k) are suspicious
		// Penalize more for longer words (long obscure words are unlikely in real messages)
		if bonus == 0 && len(sig) > 2 {
			// For words not in frequency list, add a penalty based on length
			// Short words (3-4 chars) might be abbreviations: small penalty
			// Medium words (5-8 chars): moderate penalty
			// Long words (9+ chars): heavy penalty - probably not real message words
			if len(sig) <= 4 {
				score += 50
			} else if len(sig) <= 8 {
				score += 100
			} else {
				score += 200 + (len(sig)-8)*50 // Heavy penalty for long obscure words
			}
		}
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

	// Load word frequencies first (used for scoring)
	loadWordFrequencies()

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
