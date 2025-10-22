package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/B-RAD-MC/pokedex/internal"
)

var pokecache internal.Cache
var nationaldex map[string]pokemon

func main() {
	functions := setup()
	pokecache = internal.NewCache(time.Second * 5)
	nationaldex = map[string]pokemon{}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("read error: %v\n", err)
			}
			break
		}
		input := cleanInput(scanner.Text())
		var command string
		if len(input) < 2 {
			input = append(input, "", "")
		}
		command = input[0]
		toRun, ok := functions.commands[command]
		if ok {
			err := toRun.callback(functions, input[1:])
			if err != nil {
				fmt.Printf("command error: %v\n", err)
			}
		} else {
			fmt.Print("Unknown command\n")
		}
	}
}
