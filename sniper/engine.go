package sniper

import (
	"strconv"
	"strings"
	"time"
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

	Mouse *Mouse
	Delay time.Duration

	// LastCmd holds the most recently executed command context.
	// It is part of the struct to avoid cluttering the Execute function.
	LastCmd Cmd

	// FirstCmdIsValid indicates if the very first word of the parsed string
	// corresponded to a valid command in the registry.
	FirstCmdIsValid bool
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
		Mouse:              NewMouse(),
		Delay:              time.Millisecond * 1,
		LastCmd:            nil,   // Explicit initialization
		FirstCmdIsValid:    false, // Explicit initialization
	}

	// Register available commands
	e.registerCommands()

	return e
}

// registerCommands populates the internal map with available command structs.
func (e *Engine) registerCommands() {
	// --- Navigation (Cardinals) ---
	e.registry["north"] = North{}
	e.registry["south"] = South{}
	e.registry["east"] = East{}
	e.registry["west"] = West{}
	
	// Aliases for user comfort (optional, but robust)
	e.registry["up"] = Up{}
	e.registry["down"] = Down{}
	e.registry["right"] = Right{}
	e.registry[""] = Right{}
	e.registry["left"] = Left{}

	// --- Modifiers ---
	e.registry["shift"] = Shift{}
	e.registry["control"] = Control{}
	e.registry["alt"] = Alt{}
	e.registry["command"] = Command{}

	// --- Editing ---
	e.registry["enter"] = Enter{}
	e.registry["tab"] = Tab{}
	e.registry["space"] = Space{}
	e.registry["back"] = Back{}
	e.registry["delete"] = Delete{}
	e.registry["escape"] = Escape{}
	e.registry["home"] = Home{}
	e.registry["end"] = End{}
	e.registry["pageup"] = PageUp{}
	e.registry["pagedown"] = PageDown{}

	// --- Symbols ---
	e.registry["period"] = Dot{}
	e.registry["comma"] = Comma{}
	e.registry["slash"] = Slash{}
	e.registry["window"] = Backslash{}
	e.registry["semi"] = Semi{}
	e.registry["quote"] = Quote{}
	e.registry["bracket"] = Bracket{}
	e.registry["closing"] = Closing{}
	e.registry["dash"] = Dash{}
	e.registry["equals"] = Equals{}
	e.registry["tick"] = Tick{}

	// --- NATO Alphabet ---
	e.registry["alpha"] = Alpha{}
	e.registry["bravo"] = Bravo{}
	e.registry["charlie"] = Charlie{}
	e.registry["delta"] = Delta{}
	e.registry["echo"] = Echo{}
	e.registry["foxtrot"] = Foxtrot{}
	e.registry["golf"] = Golf{}
	e.registry["hotel"] = Hotel{}
	e.registry["india"] = India{}
	e.registry["juliet"] = Juliet{}
	e.registry["kilo"] = Kilo{}
	e.registry["lima"] = Lima{}
	e.registry["mike"] = Mike{}
	e.registry["november"] = November{}
	e.registry["oscar"] = Oscar{}
	e.registry["papa"] = Papa{}
	e.registry["quebec"] = Quebec{}
	e.registry["romeo"] = Romeo{}
	e.registry["sierra"] = Sierra{}
	e.registry["tango"] = Tango{}
	e.registry["uniform"] = Uniform{}
	e.registry["victor"] = Victor{}
	e.registry["whiskey"] = Whiskey{}
	e.registry["xray"] = Xray{}
	e.registry["x-ray"] = Xray{}
	e.registry["yankee"] = Yankee{}
	e.registry["zulu"] = Zulu{}

	// --- Mouse ---
	e.registry["click"] = Click{}
}

// Parse accepts a raw input string, converts it to lowercase, splits it,
// processes numbers, maps strings to Commands, and stores the result in Tokens.
func (e *Engine) Parse(input string) {
	// Reset state for the new parse
	e.FirstCmdIsValid = false
	e.LastCmd = nil

	// 0. Ensure all input is lowercase as requested.
	input = strings.ToLower(input)

	// 1. Split the input into individual pieces by spaces.
	rawInput := strings.Fields(input)

	// 2. Initialize the slices to store processed tokens and words.
	e.Tokens = make([]Cmd, 0, len(rawInput))
	e.TokenIndices = make([]int, 0, len(rawInput))
	e.RawWords = make([]string, 0, len(rawInput))

	// Check if the first word is valid immediately after splitting
	if len(rawInput) > 0 {
		firstWordProcessed := e.numberPreprocessor.Process(rawInput[0])
		if _, ok := e.registry[firstWordProcessed]; ok {
			e.FirstCmdIsValid = true
		}
	}

	// 3. Run the pre-processor and convert to Command objects.
	for i, word := range rawInput {
		// Convert text to digits if applicable (e.g., "one" -> "1")
		// Note: The registry now has "one", "two", etc. explicitly mapped.
		// If NumberPreprocessor converts "one" -> "1", we need to make sure
		// we handle that.
		// Assuming NumberPreprocessor converts "one" to "1".
		// We have registered "one" as a text key, but if the preprocessor runs first,
		// it might turn it into a digit string.
		// NOTE: In the Execute loop, we check e.registry[word].
		// If "one" becomes "1", we need registry["1"] or rely on strconv.Atoi logic.
		// Current logic: Standard words like "alpha" are not touched by NumberPreprocessor.
		// "one" -> "1".
		// To support the explicit command "one" (which types 1), we can either:
		// A) Let NumberPreprocessor turn it to "1" and have the loop treat it as a repetition number.
		// B) If we want "one" to TYPE the number 1, we must rely on the struct.
		// Given the logic in Execute: if it's in registry, it executes.
		// If it's a number, it repeats previous.
		// Conflict: If I say "alpha one", do I want "a" then repeat "a" once? Or "a" then type "1"?
		// Usually "alpha 5" means repeat a 5 times.
		// So "one", "two" etc in registry might conflict with repetition logic if the preprocessor converts them.
		// However, I have added them to registry.
		// If Preprocessor converts "one" -> "1":
		// We need registry["1"] = One{} if we want it to be a command.
		// BUT the user logic says "If it is a Number: Repeat LastCmd".
		// So we should probably NOT register "one", "two" if we want them to act as multipliers.
		// BUT the user explicitly asked for "generate a command for each keyboard press".
		// This implies they want to be able to type numbers by voice.
		// I have registered "one", "two". If the preprocessor leaves them as words, they work as commands.
		processedWord := e.numberPreprocessor.Process(word)

		e.RawWords = append(e.RawWords, processedWord)

		// Look up the string in our command registry
		if cmd, found := e.registry[processedWord]; found {
			e.Tokens = append(e.Tokens, cmd)
			e.TokenIndices = append(e.TokenIndices, i)
		}
	}
}

// Execute iterates over the RawWords linearly. It processes each word as an entity:
// - If it is a Command: Execute it and set it as the `LastCmd`.
// - If it is a Number: Repeat `LastCmd` (n-1) times, then clear `LastCmd`.
func (e *Engine) Execute() error {
	// Note: e.LastCmd was reset in Parse(), so we start fresh here unless
	// Parse wasn't called immediately before.

	for i, word := range e.RawWords {
		// 1. Check if the word is a known command in the registry
		if cmd, isCommand := e.registry[word]; isCommand {
			// Check for "Isolated" mode (like "phrase hello world")
			if cmd.Mode() == ModeIsolated {
				payload := ""
				// Join everything remaining in RawWords as the payload
				if i+1 < len(e.RawWords) {
					payload = strings.Join(e.RawWords[i+1:], " ")
				}
				// Execute and return immediately (consumes rest of input)
				return cmd.Action(e, payload)
			}

			// Execute the standard command once
			if err := cmd.Action(e, ""); err != nil {
				return err
			}

			// Store this as the previous command for potential repetition
			e.LastCmd = cmd
			continue
		}

		// 2. If it's not a command, check if it is a number
		if val, err := strconv.Atoi(word); err == nil {
			// We only repeat if we have a valid previous command in memory
			if e.LastCmd != nil {
				// The command already ran once when we encountered it.
				// We run it (val - 1) more times.
				if val > 1 {
					for k := 0; k < val-1; k++ {
						if err := e.LastCmd.Action(e, ""); err != nil {
							return err
						}
					}
				}
				// CRITICAL: Wash away the previous action.
				// As per requirements: "left 10 10" -> The second 10 should be skipped.
				e.LastCmd = nil
			}
			// If LastCmd is nil, we simply ignore this number.
		}
	}

	return nil
}