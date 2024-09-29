package main

//go build && ./pokedexcli
//for quick save and recompile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Crimsonchamp/pokedexcli/internal/pokecache"
)

// Pokemon location struct from JSON, used for listing different areas, taken from PokeAPI
type Location struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// Pokemon area struct from JSon, used for listing different pokemon in chosen area, taken from PokeAPI
type Area struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

// Command struct for functions below.
type cliCommand struct {
	name        string
	description string
	callback    func(cache *pokecache.Cache, storage *Storage, arg1 any)
}

// Struct for the Pokemon themselves, used in Catch command, taken from PokeAPI
type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Cries          struct {
		Latest string `json:"latest"`
		Legacy string `json:"legacy"`
	} `json:"cries"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height                 int    `json:"height"`
	HeldItems              []any  `json:"held_items"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	PastAbilities []any  `json:"past_abilities"`
	PastTypes     []any  `json:"past_types"`
	Species       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       any    `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  any    `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      any    `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale any    `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string `json:"front_default"`
				FrontFemale  any    `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string `json:"back_default"`
				BackFemale       any    `json:"back_female"`
				BackShiny        string `json:"back_shiny"`
				BackShinyFemale  any    `json:"back_shiny_female"`
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string `json:"back_default"`
						BackFemale       any    `json:"back_female"`
						BackShiny        string `json:"back_shiny"`
						BackShinyFemale  any    `json:"back_shiny_female"`
						FrontDefault     string `json:"front_default"`
						FrontFemale      any    `json:"front_female"`
						FrontShiny       string `json:"front_shiny"`
						FrontShinyFemale any    `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

// Struct for Pokemon Storage,
type Storage struct {
	box map[string]*Pokemon
}

// Help function
func commandHelp(_ *pokecache.Cache, _ *Storage, _ any) {
	fmt.Println("\nCommand list:")
	fmt.Println("-help:Prints this list")
	fmt.Println("-exit:Exits the Pokedex")
	fmt.Println("-mapf:Shows next 20 areas")
	fmt.Println("-mapb:Shows last 20 areas")
	fmt.Println("-explore area: Replace area with area name from map commands ")
	fmt.Println("-catch pokemon: Use it's name instead of typing pokemon")
	fmt.Println("-remove pokemon: Use it's name instead of typing pokemon")
	fmt.Println("-inspect pokemon: Use it's name instead of typing pokemon")
	fmt.Println("-pokedex: Lists pokemon you have caught")
}

// Exit function
func commandExit(_ *pokecache.Cache, _ *Storage, _ any) {
	fmt.Println("Exiting Pokedex!")
	os.Exit(0)
}

// Base Url and location pointer for M(ove)F(orward) and M(ove)B(ack)
var baseURL = "https://pokeapi.co/api/v2/location-area"
var currentLocation *Location

// Reads and prints location info, then updates page pointers.
func commandMF(cache *pokecache.Cache, _ *Storage, _ any) {

	//Check if first call, regular use or last page.
	var url string
	if currentLocation == nil {
		url = baseURL
	} else if currentLocation.Next != nil {
		url = *currentLocation.Next
	} else {
		fmt.Println("\nLast Page!")
		return
	}

	//initialize data for cache check
	var data []byte
	var err error

	//If data found in cache, use cached data as data, bypass http get, else use http get and add to cache.
	cacheData, found := cache.Get(url)
	if found {
		fmt.Println("\nUsing Cached Data")
		data = cacheData
	} else {
		fmt.Println("\nFetching New Data")
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		//turn url info into json
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		cache.Add(url, data)
	}

	var locations Location

	//turn json into Location struct
	if err = json.Unmarshal(data, &locations); err != nil {
		fmt.Println(err)
		return
	}

	//Print each Result of location
	fmt.Println("\nAreas:")
	fmt.Println("--------------")
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	//Updates location marker
	currentLocation = &locations
}

// Same as above, but going to previous page.
func commandMB(cache *pokecache.Cache, _ *Storage, _ any) {
	var url string

	if currentLocation == nil {
		url = baseURL
	} else if currentLocation.Previous != nil {
		url = *currentLocation.Previous
	} else {
		fmt.Println("\nFirst Page!")
		return
	}

	//initialize data for cache check
	var data []byte
	var err error

	//If data found in cache, use cached data as data, bypass http get, else use http get and add to cache.
	cacheData, found := cache.Get(url)
	if found {
		fmt.Println("\nUsing Cached Data")
		data = cacheData
	} else {
		fmt.Println("\nFetching New Data")
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		//turn url info into json
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		cache.Add(url, data)
	}

	var locations Location

	//turn json into Location struct
	if err = json.Unmarshal(data, &locations); err != nil {
		fmt.Println(err)
		return
	}

	//Print each Result of location
	fmt.Println("\nAreas:")
	fmt.Println("--------------")
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	//Updates location marker
	currentLocation = &locations
}

// Reads and prints regional pokemon info, then updates page pointers.
func commandExplore(cache *pokecache.Cache, _ *Storage, answer any) {

	query, ok := answer.(string)
	if !ok {
		fmt.Println("Error, Incorrect Format - Use: Explore Area")
		return
	}

	//initialize data for cache check
	var data []byte
	var err error

	url := "https://pokeapi.co/api/v2/location-area/" + query + "/"
	fmt.Println(url)

	//If data found in cache, use cached data as data, bypass http get, else use http get and add to cache.
	cacheData, found := cache.Get(url)
	if found {
		fmt.Println("\nUsing Cached Data")
		data = cacheData
	} else {
		fmt.Println("\nFetching New Data")
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Get Error:", err)
			return
		}
		defer resp.Body.Close()

		//turn url info into json
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Data Read Error:", err)
			return
		}

		//Checks if call is empty
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API returned non-OK status: %v\n", resp.Status)
			fmt.Println("Check for Typo!")
			return
		}

		cache.Add(url, data)
	}

	var area Area

	//turn json into area struct
	if err = json.Unmarshal(data, &area); err != nil {
		fmt.Println("Unmarshal Error:", err)
		return
	}

	//Print each pokemon in area
	fmt.Println("\nLocal Pokemon:")
	fmt.Println("--------------")
	for _, encounter := range area.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}

}

// Attempts to 'catch' pokemon, if successful, adds to storage
func commandCatch(_ *pokecache.Cache, s *Storage, answer any) {

	//Make sure arg is string
	query, ok := answer.(string)
	if !ok {
		fmt.Println("Error, Incorrect Format - Use: catch pokemon")
		return
	}

	//Checks if storage already contains said pokemon
	_, exists := s.box[query]
	if exists {
		fmt.Println("Don't be greedy! One per trainer")
		return
	}

	//Search url+query
	url := "https://pokeapi.co/api/v2/pokemon/" + query

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error Fetching URL: ", err)
	}

	//Url > data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error Reading URL: ", err)
	}

	//Checks if call is empty
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned non-OK status: %v\n", resp.Status)
		fmt.Println("Check for Typo!")
		return
	}

	//Initiate pokemon variable
	var mon *Pokemon

	//Fill mon with unmarshalled data
	if err = json.Unmarshal(data, &mon); err != nil {
		fmt.Println("Error Reading Json: ", err)
		return
	}

	// Establish random seed

	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	// Generate a random integer in the range [0, 500)
	rN := randGenerator.Intn(500)

	fmt.Println("Throwing Pokeball!")
	if rN >= mon.BaseExperience {
		fmt.Println(".")
		fmt.Println(".")
		fmt.Println(".")
		fmt.Printf("%v was caught!\n", mon.Name)
		s.box[mon.Name] = mon

	} else if rN < mon.BaseExperience && rN > (mon.BaseExperience/2) {
		fmt.Println(".")
		fmt.Println(".")
		fmt.Printf("%v escaped! So close!\n", mon.Name)
	} else {
		fmt.Println(".")
		fmt.Printf("%v immediately escaped!\n", mon.Name)
	}
}

// Removes pokemon from storage
func commandRelease(_ *pokecache.Cache, s *Storage, answer any) {
	query, ok := answer.(string)
	if !ok {
		fmt.Println("Error, Incorrect Format - Use: remove pokemon")
		return
	}
	_, exists := s.box[query]
	if !exists {
		fmt.Println("You have not caught this pokemon!")
		return
	}
	delete(s.box, query)
}

// Prints pokemon stats
func commandInspect(_ *pokecache.Cache, s *Storage, answer any) {
	query, ok := answer.(string)
	if !ok {
		fmt.Println("Error, Incorrect Format - Use: catch pokemon")
		return
	}

	pokemon, exists := s.box[query]
	if !exists {
		fmt.Println("You have not caught this pokemon!")
		return
	}

	fmt.Println("Name: ", pokemon.Name)
	fmt.Println("Height: ", pokemon.Height)
	fmt.Println("Weight: ", pokemon.Weight)
	fmt.Println("Stats: ")
	fmt.Println(" -hp: ", pokemon.Stats[0].BaseStat)
	fmt.Println(" -attack: ", pokemon.Stats[1].BaseStat)
	fmt.Println(" -defense: ", pokemon.Stats[2].BaseStat)
	fmt.Println(" -special attack: ", pokemon.Stats[3].BaseStat)
	fmt.Println(" -special defense: ", pokemon.Stats[4].BaseStat)
	fmt.Println(" -speed: ", pokemon.Stats[5].BaseStat)
	fmt.Println("Types: ")
	fmt.Println(" - ", pokemon.Types[0].Type.Name)
	if len(pokemon.Types) > 1 {
		fmt.Println(" - ", pokemon.Types[1].Type.Name)
	}
}

// Prints list of pokemon in storage
func commandPokedex(_ *pokecache.Cache, s *Storage, _ any) {
	fmt.Println("Current Box:")
	for _, pokemon := range s.box {
		fmt.Println("-", pokemon.ID, " ", pokemon.Name)
	}
}

// Initializes storage for pokemon catching
func getStorage() *Storage {
	return &Storage{
		box: make(map[string]*Pokemon),
	}
}

// List of commands to pull from, takes cache for map commands.
func getCommandMap() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit pokedex",
			callback:    commandExit,
		},
		"mapf": {
			name:        "mapf",
			description: "Show the next 20 locations",
			callback:    commandMF,
		},
		"mapb": {
			name:        "mapb",
			description: "Shows the previous 20 locations",
			callback:    commandMB,
		},
		"explore": {
			name:        "explore",
			description: "Shows the pokemon in associated area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "explore",
			description: "Attempts to catch pokemon",
			callback:    commandCatch,
		},
		"release": {
			name:        "release",
			description: "Remove Pokemon from storage",
			callback:    commandRelease,
		},
		"inspect": {
			name:        "inspect",
			description: "Print Pokemon Stats",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Prints pokemon in storage",
			callback:    commandPokedex,
		},
	}
}

func main() {
	fmt.Println("Welcome to a Pokedex!\nType 'help' if you need guidance!")

	cache := pokecache.NewCache(5 * time.Minute)

	storage := getStorage()

	commands := getCommandMap()

	//Scanner set up to read inputs
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("pokedex > ")

		// If input is scanned, enter
		if scanner.Scan() {
			input := scanner.Text()
			cmdln := strings.Fields(input)
			cmdName := ""
			arg1 := ""

			if len(cmdln) > 0 {
				cmdName = cmdln[0]
			}
			if len(cmdln) > 1 {
				arg1 = cmdln[1]
			}

			//Parse Input to command map, trigger input's callback command.
			if cmd, exists := commands[cmdName]; exists {
				cmd.callback(cache, storage, arg1)
			} else {
				fmt.Println("Sorry I don't understand, type 'help' for commands")
			}
		}
	}
}
