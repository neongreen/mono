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
	"i": 100, "a": 100, "am": 90, "an": 90, "as": 90, "at": 90, "be": 90, "by": 90, "do": 90, "go": 90,
	"he": 90, "if": 90, "in": 90, "is": 90, "it": 90, "me": 90, "my": 90, "no": 90, "of": 90, "on": 90,
	"or": 90, "so": 90, "to": 90, "up": 90, "us": 90, "we": 90, "all": 80, "and": 80, "any": 80, "are": 80,
	"but": 80, "can": 80, "did": 80, "for": 80, "get": 80, "got": 80, "had": 80, "has": 80, "her": 80,
	"him": 80, "his": 80, "its": 80, "let": 80, "may": 80, "not": 80, "now": 80, "one": 80, "our": 80,
	"out": 80, "say": 80, "she": 80, "the": 80, "too": 80, "two": 80, "was": 80, "who": 80, "you": 80,
	"your": 75, "they": 75, "this": 75, "that": 75, "with": 75, "have": 75, "will": 75, "from": 75,
	"been": 75, "just": 75, "only": 75, "over": 75, "such": 75, "take": 75, "into": 75, "than": 75,
	"them": 75, "then": 75, "some": 75, "what": 75, "when": 75, "make": 75, "like": 75, "time": 75,
	"very": 75, "know": 75, "want": 75, "give": 75, "most": 75, "also": 75, "back": 75, "come": 75,
	"about": 70, "after": 70, "first": 70, "think": 70, "could": 70, "would": 70, "there": 70, "their": 70,
	"which": 70, "these": 70, "other": 70, "being": 70, "those": 70, "still": 70, "while": 70, "where": 70,
	"since": 70, "under": 70, "right": 70, "never": 70, "every": 70, "going": 70, "might": 70,
	"should": 65, "people": 65, "before": 65, "really": 65, "always": 65, "things": 65, "little": 65,
	"something": 60, "anything": 60, "everything": 60, "nothing": 60, "tonight": 60, "promise": 60,
	"important": 55, "absolutely": 50, "humiliating": 45,
	"tell": 70, "guys": 65, "spill": 60, "mixing": 55, "publicly": 50, "story": 65, "secret": 60,
	"must": 75, "trust": 60, "truly": 55, "stupid": 55, "mix": 60,
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
	// Exact match?
	if words, ok := d.signatures[sortedSig]; ok {
		bonus := d.commonSigs[sortedSig]
		return 0, words, bonus
	}

	// If the signature is too long, cap the penalty
	if len(sortedSig) > 20 {
		return len(sortedSig), nil, 0
	}

	bestDist := len(sortedSig) + 5 // Default penalty
	var bestWords []string
	bestBonus := 0
	targetLen := len(sortedSig)

	// Search nearby lengths first
	for offset := 0; offset <= min(targetLen, 10); offset++ {
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

				// Prefer common words even if slightly worse distance
				effectiveScore := dist*10 - bonus

				bestEffective := bestDist*10 - bestBonus
				if effectiveScore < bestEffective || (effectiveScore == bestEffective && dist < bestDist) {
					bestDist = dist
					bestWords = d.signatures[sig]
					bestBonus = bonus
				}

				if bestDist == 0 && bestBonus >= 50 {
					return bestDist, bestWords, bestBonus
				}
			}
		}
		if bestDist <= 1 && bestBonus >= 50 {
			break
		}
	}

	return bestDist, bestWords, bestBonus
}

type WordMatch struct {
	Signature string
	Distance  int
	Words     []string
	Bonus     int
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
	totalError := 0
	matches := make([]WordMatch, len(sigs))

	for i, sig := range sigs {
		dist, words, bonus := dict.findNearest(sig)
		// Score: edit distance penalty minus commonality bonus
		score := dist*10 - bonus
		totalError += score
		matches[i] = WordMatch{Signature: sig, Distance: dist, Words: words, Bonus: bonus}
	}

	return totalError, matches
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
	totalDist := 0
	for i, m := range sol.Matches {
		word := "???"
		if len(m.Words) > 0 {
			word = m.Words[0]
		}
		words[i] = word
		totalDist += m.Distance
		commonMark := ""
		if m.Bonus > 0 {
			commonMark = fmt.Sprintf(" [common:%d]", m.Bonus)
		}
		fmt.Printf("  '%s' -> '%s' (dist=%d%s)\n", m.Signature, word, m.Distance, commonMark)
	}
	fmt.Printf("\nTotal edit distance: %d\n", totalDist)
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
