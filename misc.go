package manageLedger2

import (
	"fmt"
	"log"
)

// Abort logs an error, then panics
func Abort(err error, msg string) {
	fmt.Println()
	log.Printf("[ERROR] %s", err)
	panic(msg)
}
