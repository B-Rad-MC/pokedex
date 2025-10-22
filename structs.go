package main

type cliCommand struct {
	name        string
	description string
	callback    func(*app, []string) error
}

type config struct {
	Next     *string
	Previous *string
	Results  []resource
}

type locationArea struct {
	Name               *string
	Pokemon_encounters []pokemonEncounter
}

type pokemonEncounter struct {
	Pokemon pokemon
}

type pokemon struct {
	Name            string
	Base_experience int
	Height          int
	Weight          int
	Stats           []pokemonStat
	Types           []pokemonType
}

type pokemonStat struct {
	Stat      stat
	Base_stat int
}

type stat struct {
	Name string
}

type pokemonType struct {
	Type Type
}

type Type struct {
	Name string
}

type app struct {
	commands     map[string]cliCommand
	globalConfig config
}

type resource struct {
	Name string
}
