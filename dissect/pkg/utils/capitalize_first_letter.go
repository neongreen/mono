package utils

import (
	"strings"
)

func CapitalizeFirstLetter(targetFuncName string) string {
	if len(targetFuncName) == 0 {
		return targetFuncName
	}
	return strings.ToUpper(string(targetFuncName[0])) + targetFuncName[1:]
}
