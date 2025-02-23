package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func([]string, *config) error
}

type config struct {
	baseCatchRate int
	next          string
	previous      string
	locale        string
	currentMap    PokeMap
	pokedex       map[string]Pokemon
	selectedBall  Pokeball
	pokeballs     [4]Pokeball
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
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 3, 1, ' ', 0)

	if len(options) > 0 {
		if v, ok := commands[options[0]]; ok {
			fmt.Fprintf(w, "%s:\t%s\t\n", v.name, v.description)
			w.Flush()
			return nil
		}
		return fmt.Errorf("no help entry found for command: %s", options[0])
	}

	var keyOrder []string
	for k := range commands {
		keyOrder = append(keyOrder, k)
	}
	sort.Strings(keyOrder)
	for _, k := range keyOrder {
		fmt.Fprintf(w, "%s:\t%s\t\n", k, commands[k].description)
	}
	w.Flush()

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
			return fmt.Errorf("response failed with code: %d %s", res.StatusCode, body)
		}
		if err != nil {
			return fmt.Errorf("failed to read response body with error: %w", err)
		}
	}
	cache.Add(url, body)

	var pokeMapRes PokeMapResponse
	if err := json.Unmarshal(body, &pokeMapRes); err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}
	config.next = pokeMapRes.Next
	config.previous = pokeMapRes.Previous
	for i := 0; i < len(pokeMapRes.Results); i++ {
		println(pokeMapRes.Results[i].Name)
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
			return fmt.Errorf("response failed with code: %d %s", res.StatusCode, body)
		}
		if err != nil {
			return fmt.Errorf("failed to read response body with error: %w", err)
		}
	}
	cache.Add(url, body)

	var pokeMapRes PokeMapResponse
	if err := json.Unmarshal(body, &pokeMapRes); err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}
	config.next = pokeMapRes.Next
	config.previous = pokeMapRes.Previous
	for i := 0; i < len(pokeMapRes.Results); i++ {
		println(pokeMapRes.Results[i].Name)
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

	body, ok := cache.Get(location) // using location for now, maybe its better for consistency to use the built url?
	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with code: %d %s for location: %s", res.StatusCode, body, location)
		}
		if err != nil {
			return fmt.Errorf("failed to read response body with error: %w", err)
		}
	}
	cache.Add(location, body)

	var pokeMapData PokeMap
	if err := json.Unmarshal(body, &pokeMapData); err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}
	config.currentMap = pokeMapData
	fmt.Printf("Pokemon found in %s...\n", location)
	for _, pokemon := range pokeMapData.PokemonEncounters {
		fmt.Printf(" - %s\n", pokemon.Pokemon.Name)
	}
	return nil
}
func commandCatch(options []string, config *config) error {
	var url string
	var pokeName string
	var latestVersion string
	if len(options) != 1 {
		if len(options) > 1 {
			return fmt.Errorf("too many arguments provided, expecting 1 found: %s", options)
		} else {
			return errors.New("missing pokemon name argument")
		}
	} else {
		if config.currentMap.Location.Name == "" {
			return errors.New("no location has been explored yet")
		}
		pokeName = options[0]
		for _, p := range config.currentMap.PokemonEncounters {
			if pokeName == p.Pokemon.Name {
				url = p.Pokemon.URL
				latestVersion = p.VersionDetails[len(p.VersionDetails)-1].Version.Name
			}
		}
		if url == "" {
			return fmt.Errorf("%s isn't found in %s", pokeName, config.currentMap.Location.Name)
		}
	}

	// check if caught already
	encounter, ok := config.pokedex[pokeName]
	if !ok {
		res, err := http.Get(url) // else get pokemon encounter information
		if err != nil {
			return err
		}
		pokeBody, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with code: %d %s for url: %s", res.StatusCode, pokeBody, url)
		}
		if err != nil {
			return fmt.Errorf("failed to read response body with error: %w", err)
		}
		if err := json.Unmarshal(pokeBody, &encounter); err != nil {
			return fmt.Errorf("error unmarshalling data: %w", err)
		}
		config.pokedex[pokeName] = encounter
	}
	encounter.encountered++

	catchRate := config.baseCatchRate + config.selectedBall.catchModifier
	catchDifficulty := int(math.Min(float64(encounter.BaseExperience)*float64(255)/float64(340), 255.0)) // normalise catch rate to 0-255, yes this does result in blissey being as hard to catch as the legendaries
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	success := 0
	fmt.Printf("throwing %s at %s!\n", config.selectedBall.name, pokeName)
	for i := 0; i < 3; i++ {
		if r.Intn(catchDifficulty) <= catchRate {
			success++
		}
		time.Sleep(1 * time.Second)
		print(".")
	}

	if success >= 3 {
		fmt.Printf("\ncaught %s!\n", encounter.Species.Name)
		if encounter.caught == 0 {
			println("adding to pokedex...")
			res, err := http.Get(encounter.Species.URL) // get pokedex information
			if err != nil {
				return err
			}
			dexBody, err := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode > 299 {
				return fmt.Errorf("response failed with code: %d %s for url: %s", res.StatusCode, dexBody, url)
			}
			if err != nil {
				return fmt.Errorf("failed to retrieve pokedex information: %w", err)
			}

			var dexEntry PokedexEntry
			if err := json.Unmarshal(dexBody, &dexEntry); err != nil {
				return fmt.Errorf("error unmarshalling data: %w", err)
			}
			encounter.AddDexEntry(config.locale, &dexEntry)

			str, err := encounter.GetFlavorText(latestVersion)
			if err != nil {
				println("Error getting pokedex entry")
			} else {
				fmt.Printf("%q\n", str)
			}
		}
		encounter.caught++
	} else {
		fmt.Printf("\nit got away...\n")
	}
	config.pokedex[pokeName] = encounter

	return nil
}
func commandSelect(options []string, config *config) error {
	if len(options) < 1 || len(options) > 2 {
		if len(options) > 2 {
			return fmt.Errorf("too many arguments provided, expecting 1-2 found: %s", options)
		} else {
			return errors.New("missing pokeball argument")
		}
	} else {
		for _, ball := range config.pokeballs {
			arg := strings.Join(options[0:], " ")
			if arg == ball.name {
				config.selectedBall = ball
				fmt.Printf("%s equipped\n", arg)
				return nil
			}
		}
	}
	return fmt.Errorf("cannot understand argument, please select from:\n%s\n%s\n%s\n%s", config.pokeballs[0].name, config.pokeballs[1].name, config.pokeballs[2].name, config.pokeballs[3].name)
}
func commandPokedex(options []string, config *config) error {
	if len(options) == 0 {
		var pokedexEntries []Pokemon
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 3, 1, ' ', 0)

		for _, v := range config.pokedex {
			pokedexEntries = append(pokedexEntries, v)
		}
		slices.SortFunc(pokedexEntries, func(a, b Pokemon) int {
			if a.ID < b.ID {
				return -1
			}
			if a.ID > b.ID {
				return 1
			}
			return 0
		})
		for _, entry := range pokedexEntries {
			fmt.Fprintf(w, "#%04d\t%s\tencountered: %d\tcaught: %d\t\n", entry.ID, entry.Species.Name, entry.encountered, entry.caught)
		}

		w.Flush()

		return nil
	}

	if len(options) > 1 {
		return errors.New("too many arguments, expecting 0-1")
	}
	name := options[0]
	pokemon, ok := config.pokedex[name]
	if !ok {
		return fmt.Errorf("missing entry for %s", name)
	}

	fmt.Printf("#%04d %s the %s\n", pokemon.ID, pokemon.Species.Name, pokemon.Genera[0].Genus)
	fmt.Printf("encountered: %d caught: %d\n", pokemon.encountered, pokemon.caught)
	println("stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("\t- %s %d\n", stat.Stat.Name, stat.BaseStat)
	}
	version := pokemon.FlavorTextEntries[len(pokemon.FlavorTextEntries)-1].Version.Name // I should come up with a better way to do this
	fmt.Printf("pokedex entry for version: %s\n", version)
	entry, _ := pokemon.GetFlavorText(version)
	println(entry)

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
	cfg = config{
		locale:        "en",
		baseCatchRate: 20,
		pokedex:       make(map[string]Pokemon),
		pokeballs: [4]Pokeball{
			{20, "pokeball"},
			{35, "great ball"},
			{55, "ultra ball"},
			{255, "master ball"},
		},
	}
	cfg.selectedBall = cfg.pokeballs[0]
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
			description: "display the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "display the previous 20 locations",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "display a list of encountered pokemon in a given location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "attempt to catch a pokemon found in the current map",
			callback:    commandCatch,
		},
		"select": {
			name:        "select",
			description: "select a pokeball to use when attempting to catch pokemon",
			callback:    commandSelect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "get the pokedex entry of a selected pokemon",
			callback:    commandPokedex,
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
