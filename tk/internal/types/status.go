package types

// Role authority levels (higher is more authoritative)
var roleAuthority = map[string]int{
	"human": 5,
	"qa":    4,
	"rel":   3,
	"agent": 2,
	"bot":   1,
}

// Claim represents a status assertion by an actor
type Claim struct {
	State     string `json:"state"`
	Role      string `json:"role"`
	Tentative bool   `json:"tentative"`
	TS        int64  `json:"ts"`
}

// AxisStatus represents the status claims for a single axis
type AxisStatus struct {
	Effective string  `json:"effective"`
	Claims    []Claim `json:"claims"`
}

// GetRoleAuthority returns the authority level for a role
func GetRoleAuthority(role string) int {
	if auth, ok := roleAuthority[role]; ok {
		return auth
	}
	return 0
}
