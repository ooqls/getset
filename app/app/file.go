package app

import (
	"fmt"
	"os"
)

func fileExists(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return false
	}

	return true
}

func mustReadFile(p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		panic(fmt.Errorf("[Failure] failed to read file %s: %v", p, err))
	}

	return b
}


