package sniper

import (
	"strings"

	"github.com/go-vgo/robotgo"
)


type Engine struct {
	// We hold pointers to the types so that the Engine references
	// the specific instances created by their "New" functions.
	// This preserves the state (active modifiers, mutexes) of each.
	StickyKeyboard *StickyKeyboard

	// numberPreprocessor is the internal instance responsible for converting text to digits.
	numberPreprocessor *NumberPreprocessor

	// registry maps string command names (e.g., "left") to their Cmd implementation.
	registry map[string]Cmd

	// Tokens holds the slice of Commands after the input string has been
	// split, processed, and matched against the registry.
	Tokens []Cmd

	// TokenIndices maps the index of the Token in the Tokens slice
	// to its original index in the RawWords slice.
	TokenIndices []int

	// RawWords holds the processed string slice of the input.
	// We keep this so "Isolated" commands can access the text following them.
	RawWords []string

	Mouse Mouse
}

// NewEngine initializes the Engine and sets up the internal keyboard types
// with their default safety delays and memory allocation.
func NewEngine() *Engine {
	e := &Engine{
		StickyKeyboard:     NewStickyKeyboard(),
		numberPreprocessor: NewNumberPreprocessor(),
		registry:           make(map[string]Cmd),
		Tokens:             make([]Cmd, 0),
		TokenIndices:       make([]int, 0),
		RawWords:           make([]string, 0),
	}

	// Register available commands
	e.registerCommands()

	return e
}

// registerCommands populates the internal map with available command structs.
func (e *Engine) registerCommands() {
	// We map the string result of Name() or a specific trigger word to the struct.
	e.registry["left"] = Left{}
	e.registry["right"] = Right{}
	e.registry["up"] = Up{}
	e.registry["down"] = Down{}
	// Assuming you have a Phrase command implementation:
	// e.registry["phrase"] = Phrase{} 
}

// Parse accepts a raw input string, converts it to lowercase, splits it,
// processes numbers, maps strings to Commands, and stores the result in Tokens.
func (e *Engine) Parse(input string) {
	// 0. Ensure all input is lowercase as requested.
	input = strings.ToLower(input)

	// 1. Split the input into individual pieces by spaces.
	rawInput := strings.Fields(input)

	// 2. Initialize the slices to store processed tokens and words.
	e.Tokens = make([]Cmd, 0, len(rawInput))
	e.TokenIndices = make([]int, 0, len(rawInput))
	e.RawWords = make([]string, 0, len(rawInput))

	// 3. Run the pre-processor and convert to Command objects.
	for i, word := range rawInput {
		// Convert text to digits if applicable (e.g., "one" -> "1")
		processedWord := e.numberPreprocessor.Process(word)
		
		// We store the processed word in RawWords so that if we capture a phrase,
		// we get "1" instead of "one" if that's what the preprocessor did.
		e.RawWords = append(e.RawWords, processedWord)

		// Look up the string in our command registry
		if cmd, found := e.registry[processedWord]; found {
			e.Tokens = append(e.Tokens, cmd)
			// Track which index in RawWords this command corresponds to
			e.TokenIndices = append(e.TokenIndices, i)
		}
	}
}

// Execute iterates over the parsed Tokens and runs the Action for each one.
// It returns the first error encountered, if any.
func (e *Engine) Execute() error {
	for i, cmd := range e.Tokens {
		// If the command is Isolated (like "phrase"), it consumes the rest of the input.
		if cmd.Mode() == ModeIsolated {
			// Find where this command was located in the raw word list
			rawIndex := e.TokenIndices[i]

			// Join everything after this command into a single string
			payload := ""
			if rawIndex+1 < len(e.RawWords) {
				payload = strings.Join(e.RawWords[rawIndex+1:], " ")
			}

			// Execute the action with the payload
			if err := cmd.Action(e, payload); err != nil {
				return err
			}

			// Since this command consumed the rest of the sentence, we stop executing tokens.
			return nil
		}

		// Execute standard commands (ModeCommand), passing empty string as payload
		if err := cmd.Action(e, ""); err != nil {
			return err
		}
		robotgo.Sleep(5)
	}
	return nil
}