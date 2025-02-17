package main

import (
	"fmt"
	"strings"
)

func cleanInput(text string) []string {
	// TODO: add err handling
	sanitised := strings.Split(strings.TrimSpace(text), " ")
	for i := 0; i < len(sanitised); i++ {
		sanitised[i] = strings.ToLower(sanitised[i])
	}

	return sanitised
}

func main() {
	fmt.Printf("You is well intentioned, you have %d big booms.\n", 5)
}
