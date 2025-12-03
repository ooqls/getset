package keydb

import "fmt"

var (
	ErrDBNotInitialized   error = fmt.Errorf("database not initialized")
	ErrIncorrectSystemKey error = fmt.Errorf("the given key is not the correct system key")
)
