package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func([]string, *config) error
}

type config struct {
	next     string
	previous string
}

var cfg config
var commands map[string]cliCommand
var cache Cache

func commandExit(options []string, config *config) error {
	println("closing...")
	os.Exit(0)
	return nil
}
func commandHelp(options []string, config *config) error {
	if len(options) > 0 {
		if v, ok := commands[options[0]]; ok {
			fmt.Printf("%s: %s\n", v.name, v.description)
			return nil
		}
		return fmt.Errorf("no help entry found for command: %s", options[0])
	}

	var keyOrder []string
	for k := range commands {
		keyOrder = append(keyOrder, k)
	}
	sort.Strings(keyOrder) // crappy sort implementation for now while i figure out a better way for ordering, probably just manual lol
	for _, k := range keyOrder {
		fmt.Printf("%s: %s\n", k, commands[k].description)
	}
	return nil
}
func commandMap(options []string, config *config) error {
	if len(options) > 0 {
		return fmt.Errorf("too many arguments, expecting 0, found: %v", options)
	}

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
func commandMapB(options []string, config *config) error {
	if len(options) > 0 {
		return fmt.Errorf("too many arguments, expecting 0, found: %v", options)
	}

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
func commandExplore(options []string, config *config) error { // maybe update config?
	if len(options) == 0 {
		return errors.New("missing location argument")
	}

	url := "https://pokeapi.co/api/v2/location-area/"
	location := strings.Join(options, "-")
	url += location

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
			description: "exit the program",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "display this information",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "display a list of pokemon locations, 20 at a time, each subsequent call of map will display the next 20 maps",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "companion command to map, display a the previous list of pokemon location, 20 at a time. If map hasn't been called prior this prints out the default 20 locations",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "display a list of encountered pokemon in a given location",
			callback:    commandExplore,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		var args []string
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := cleanInput(scanner.Text())
		if len(text) > 1 {
			args = append(args, text[1:]...)
		}
		command, ok := commands[text[0]]
		if !ok {
			fmt.Printf("invalid command: '%s'\n", text[0])
			continue
		}
		if err := command.callback(args, &cfg); err != nil {
			fmt.Printf("error calling command %s: %s\n", text[0], err)
		}

	}
}
