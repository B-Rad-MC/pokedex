package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func setup() *app {
	supportedCommands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays a list of areas in the Pokemon world. Shows 20 areas at a time. Execute this command again to view the next page.",
			callback:    commandMapA,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous page of areas in the Pokemon world. Shows 20 areas at a time. Does not work if you are on the first page.",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "Explores the specified area and returns a list of pokemon found there. Eg: explore <area-name>",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Throw a pokeball at a pokemon of your choice and attempt to catch it. Eg: catch <pokemon-name>",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays information on a pokemon you have caught before. Eg: inspect <pokemon-name>",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays the names of all the pokemon you've caught.",
			callback:    commandPokedex,
		},
	}
	sharedConfig := &config{
		Next: nullableStr("https://pokeapi.co/api/v2/location-area"),
	}
	return &app{
		commands:     supportedCommands,
		globalConfig: *sharedConfig,
	}
}

func cleanInput(text string) []string {
	prefixOk := true
	suffixOk := true
	for prefixOk {
		text, prefixOk = strings.CutPrefix(text, " ")
		if text == "" || text == " " {
			prefixOk = false
		}
	}
	for suffixOk {
		text, suffixOk = strings.CutSuffix(text, " ")
		if text == "" || text == " " {
			suffixOk = false
		}
	}
	return strings.Split(strings.ToLower(text), " ")
}

func nullableStr(s string) *string {
	return &s
}

func commandExit(parent *app, _ []string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(parent *app, _ []string) error {
	fmt.Print("Welcome to the Pokedex!\n")
	for _, command := range parent.commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(pointer *config, URL *string, word string) error {
	if URL == nil {
		fmt.Printf("you're on the %s page\n", word)
		return nil
	}
	var data []byte
	if entry, exists := pokecache.Get(*URL); exists {
		data = entry
	} else {
		res, err := http.Get(*URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("request status code not OK")
		}

		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}
	var locationList config
	if err := json.Unmarshal(data, &locationList); err != nil {
		return err
	}
	for _, locationArea := range locationList.Results {
		fmt.Printf("%s\n", locationArea.Name)
	}
	pointer.Next = locationList.Next
	pointer.Previous = locationList.Previous
	pokecache.Add(*URL, data)
	return nil
}

func commandMapA(parent *app, _ []string) error {
	return commandMap(&parent.globalConfig, parent.globalConfig.Next, "last")
}

func commandMapB(parent *app, _ []string) error {
	return commandMap(&parent.globalConfig, parent.globalConfig.Previous, "first")
}

func commandExplore(parent *app, param []string) error {
	if param[0] == "" {
		fmt.Print("Please input the area you wish to explore. Eg. explore <area-name>\n")
		return nil
	}
	URL := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", param[0])
	var data []byte
	if entry, exists := pokecache.Get(URL); exists {
		data = entry
	} else {
		res, err := http.Get(URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("request status code not OK")
		}

		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Exploring %s...\n", param[0])
	var area locationArea
	if err := json.Unmarshal(data, &area); err != nil {
		return err
	}
	fmt.Print("Found Pokemon:\n")
	for _, foundPokemon := range area.Pokemon_encounters {
		fmt.Printf(" - %s\n", foundPokemon.Pokemon.Name)
	}
	return nil
}

func commandCatch(parent *app, param []string) error {
	//base xp 50 = 51%; base xp 306 = 0.75%
	//probability = -0.196289 * basexp + 60.81445
	if param[0] == "" {
		fmt.Print("Please input the name of the pokemon you want to catch. Eg. catch <pokemon-name>\n")
		return nil
	}
	URL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", param[0])
	var data []byte
	if entry, exists := pokecache.Get(URL); exists {
		data = entry
	} else {
		res, err := http.Get(URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("request status code not OK")
		}

		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}
	var wildMon pokemon
	if err := json.Unmarshal(data, &wildMon); err != nil {
		return err
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", param[0])
	if rand.Float64() > (-0.196289*float64(wildMon.Base_experience)+60.81445)*0.01 {
		fmt.Printf("%s escaped!\n", param[0])
	} else {
		nationaldex[param[0]] = wildMon
		fmt.Printf("%s was caught!\n", param[0])
	}
	return nil
}

func commandInspect(parent *app, param []string) error {
	caughtMon, exists := nationaldex[param[0]]
	if !exists {
		fmt.Print("you have not caught that pokemon\n")
	} else {
		fmt.Printf("Name: %s\nHeight: %v\nWeight: %v\nStats: \n", caughtMon.Name, caughtMon.Height, caughtMon.Weight)
		for _, stat := range caughtMon.Stats {
			fmt.Printf("  -%s: %v\n", stat.Stat.Name, stat.Base_stat)
		}
		fmt.Printf("Types:\n")
		for _, caughtType := range caughtMon.Types {
			fmt.Printf("  - %s\n", caughtType.Type.Name)
		}
	}
	return nil
}

func commandPokedex(parent *app, param []string) error {
	fmt.Print("Your Pokedex:\n")
	if len(nationaldex) == 0 {
		fmt.Print("...is empty!\n")
	}
	for _, caughtMon := range nationaldex {
		fmt.Printf(" - %s\n", caughtMon.Name)
	}
	return nil
}
