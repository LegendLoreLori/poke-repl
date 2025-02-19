package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	next     string
	previous string
}

var cfg config
var commands map[string]cliCommand
var cache Cache

func commandExit(config *config) error {
	println("Closing...")
	os.Exit(0)
	return nil
}
func commandHelp(config *config) error {
	print("Poke-REPL provides a CLI to query pokeAPI to retreive pokedex information\n\nUsage:\n")
	for _, v := range commands {
		fmt.Printf("%s: %s\n", v.name, v.description)
	}
	return nil
}
func commandMap(config *config) error {
	var url string
	if config.next == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	} else {
		url = config.next
	}

	body, ok := cache.Get(url)
	if !ok { // cache miss
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with code: %d and body: %s", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	cache.Add(url, body)

	var pokemap PokeMap
	if err := json.Unmarshal(body, &pokemap); err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}
	cfg.next = pokemap.Next
	cfg.previous = pokemap.Previous
	for i := 0; i < len(pokemap.Results); i++ {
		println(pokemap.Results[i].Name)
	}
	return nil
}
func commandMapB(config *config) error {
	var url string
	if config.previous == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	} else {
		url = config.previous
	}

	body, ok := cache.Get(url)
	if !ok { // cache miss
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with code: %d and body: %s", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	cache.Add(url, body)

	var pokemap PokeMap
	if err := json.Unmarshal(body, &pokemap); err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}
	cfg.next = pokemap.Next
	cfg.previous = pokemap.Previous
	for i := 0; i < len(pokemap.Results); i++ {
		println(pokemap.Results[i].Name)
	}
	return nil
}

func cleanInput(text string) []string {
	if text == "" {
		return make([]string, 1)
	}
	sanitised := strings.Fields(strings.TrimSpace(text))
	for i := 0; i < len(sanitised); i++ {
		sanitised[i] = strings.ToLower(sanitised[i])
	}

	return sanitised
}

func main() {
	cache = NewCache(20 * time.Second)
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
		"map": {
			name:        "map",
			description: "Display a list of pokemon locations, ordered by ID, 20 at a time, each subsequent call of map will display the next 20 maps, use mapb to display the previous 20 maps",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Companion command to map, display a the previous list of pokemon location, ordered by ID, 20 at a time. If map hasn't been called prior this prints out the default 20 locations",
			callback:    commandMapB,
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
		if err := command.callback(&cfg); err != nil {
			log.Fatal("error calling command: %w", err)
		}

	}
}
