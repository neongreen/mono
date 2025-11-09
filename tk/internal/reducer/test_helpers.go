package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// Shared test helper functions for property-based and exhaustive tests

// getEventOrder formats event order for error messages
func getEventOrder(events []types.Event) string {
	result := "["
	for i, e := range events {
		var payload types.TaskStatusSetPayload
		json.Unmarshal(e.Payload, &payload)
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s(TS=%d)", payload.State, e.TS)
	}
	return result + "]"
}

// mustMarshal marshals a value to JSON or panics
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// generateAllPermutations returns all permutations of the given slice
// For n items, generates n! permutations
func generateAllPermutations[T any](items []T) [][]T {
	var result [][]T

	var permute func([]T, int)
	permute = func(arr []T, n int) {
		if n == 1 {
			// Make a copy of the current permutation
			perm := make([]T, len(arr))
			copy(perm, arr)
			result = append(result, perm)
			return
		}

		for i := 0; i < n; i++ {
			permute(arr, n-1)
			if n%2 == 1 {
				arr[0], arr[n-1] = arr[n-1], arr[0]
			} else {
				arr[i], arr[n-1] = arr[n-1], arr[i]
			}
		}
	}

	// Make a copy to avoid modifying the input
	arrCopy := make([]T, len(items))
	copy(arrCopy, items)
	permute(arrCopy, len(arrCopy))

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
