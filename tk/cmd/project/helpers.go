package project

import (
	"github.com/neongreen/mono/tk/internal/utils"
)

func getCurrentUser() (string, error) {
	return utils.GetCurrentUser()
}
