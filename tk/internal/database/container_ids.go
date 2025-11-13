package database

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// GenerateContainerID generates the next container ID for the given primitive type
// Queue: q-1, q-2, q-3, ...
// Stack: s-1, s-2, s-3, ...
// Group: g-1, g-2, g-3, ...
func (d *DB) GenerateContainerID(primitive types.ContainerPrimitive) (string, error) {
	var prefix string
	switch primitive {
	case types.PrimitiveQueue:
		prefix = "q"
	case types.PrimitiveStack:
		prefix = "s"
	case types.PrimitiveGroup:
		prefix = "g"
	default:
		return "", fmt.Errorf("unknown primitive type: %s", primitive)
	}

	// Find the highest number currently used for this primitive
	var maxNum int64
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0)
		FROM containers
		WHERE id LIKE ?
	`, len(prefix)+2) // prefix + "-" = 2 chars

	err := d.Db.QueryRow(query, prefix+"-%").Scan(&maxNum)
	if err != nil {
		return "", fmt.Errorf("failed to get max container number: %w", err)
	}

	nextNum := maxNum + 1
	return fmt.Sprintf("%s-%d", prefix, nextNum), nil
}
