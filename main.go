package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands map[string]cliCommand

func commandExit() error {
	println("Closing...")
	os.Exit(0)
	return nil
}
func commandHelp() error {
	print("Poke-REPL provides a CLI to query pokeAPI to retreive pokedex information\n\nUsage:\n")
	for _, v := range commands {
		fmt.Printf("%s: %s\n", v.name, v.description)
	}
	return nil
}

func cleanInput(text string) []string {
	// TODO: add err handling
	sanitised := strings.Fields(strings.TrimSpace(text))
	for i := 0; i < len(sanitised); i++ {
		sanitised[i] = strings.ToLower(sanitised[i])
	}

	return sanitised
}

func main() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the program",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display this information",
			callback:    commandHelp,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := cleanInput(scanner.Text())
		command, ok := commands[text[0]]
		if !ok {
			fmt.Printf("Invalid command: '%s'\n", text[0])
			continue
		}
		if err := command.callback(); err != nil {
			log.Fatal("error calling command: %w", err)
		}

	}
}
