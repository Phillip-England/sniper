package sniper

import (
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

	// Tokens holds the master slice of Tokens after parsing.
	// This is the immutable sequence for the current execution cycle.
	Tokens []Token

	// RemainingTokens tracks which tokens have NOT yet been executed.
	// This slice shrinks as Execute() iterates.
	RemainingTokens []Token

	// HandledTokens tracks which tokens HAVE been executed (or are about to be).
	// This slice grows as Execute() iterates.
	HandledTokens []Token

	// RemainingRawWords tracks the unexecuted text as a single string.
	// It is updated prior to token handling so commands can see what text remains.
	RemainingRawWords string

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

	// IsOperating determines if the engine is currently allowed to execute actions.
	// If false, the engine should halt operations.
	IsOperating bool
}

// NewEngine initializes the Engine and sets up the internal keyboard types
// with their default safety delays and memory allocation.
func NewEngine() *Engine {
	e := &Engine{
		StickyKeyboard:     NewStickyKeyboard(),
		numberPreprocessor: NewNumberPreprocessor(),
		registry:           make(map[string]Cmd),
		Tokens:             make([]Token, 0),
		RemainingTokens:    make([]Token, 0),
		HandledTokens:      make([]Token, 0),
		RemainingRawWords:  "",
		TokenIndices:       make([]int, 0),
		RawWords:           make([]string, 0),
		Mouse:              NewMouse(),
		Delay:              time.Microsecond * 800,
		LastCmd:            nil,   // Explicit initialization
		FirstCmdIsValid:    false, // Explicit initialization
		IsOperating:        true,  // Defaults to true so the engine runs
	}

	// Register available commands dynamically from the global Registry
	e.registerCommands()

	return e
}

// registerCommands populates the internal map by iterating over the global Registry slice.
// It maps the command to all of its 'CalledBy' aliases.
func (e *Engine) registerCommands() {
	for _, cmd := range Registry {
		// Iterate over the slice of triggers returned by CalledBy
		for _, trigger := range cmd.CalledBy() {
			key := strings.ToLower(trigger)
			e.registry[key] = cmd
		}
	}
}

// Parse accepts a raw input string, converts it to lowercase, splits it,
// processes numbers, maps strings to Tokens, and stores the result.
func (e *Engine) Parse(input string) {
	// Reset state for the new parse
	e.FirstCmdIsValid = false
	e.LastCmd = nil

	// 0. Ensure all input is lowercase as requested.
	input = strings.ToLower(input)

	// 1. Split the input into individual pieces by spaces.
	rawInput := strings.Fields(input)

	// 2. Initialize the slices to store processed tokens and words.
	e.Tokens = make([]Token, 0, len(rawInput))
	e.TokenIndices = make([]int, 0, len(rawInput))
	e.RawWords = make([]string, 0, len(rawInput))

	// 3. Process words into Tokens using the Factory
	for i, word := range rawInput {
		// Use the TokenFactory to create the specific token type (Cmd, Number, or Raw)
		// This handles the number preprocessing internally.
		token := TokenFactory(word, e.registry, e.numberPreprocessor)

		// Store the token and the raw word (processed literal)
		e.Tokens = append(e.Tokens, token)
		e.RawWords = append(e.RawWords, token.Literal())
		e.TokenIndices = append(e.TokenIndices, i)

		// Check validity of first command (legacy logic support)
		if i == 0 && token.Type() == TokenTypeCmd {
			e.FirstCmdIsValid = true
		}
	}

	// 4. Initialize the tracking slices.
	// At the start of execution, Handled is empty, and Remaining is a copy of all Tokens.
	e.HandledTokens = make([]Token, 0, len(e.Tokens))
	e.RemainingTokens = make([]Token, len(e.Tokens))
	copy(e.RemainingTokens, e.Tokens)

	// 5. Initialize RemainingRawWords with the full sentence at the start
	e.RemainingRawWords = strings.Join(e.RawWords, " ")
}

// Execute iterates over the Tokens linearly. It delegates logic to the
// Handle function of each token.
func (e *Engine) Execute() error {
	// Note: e.LastCmd was reset in Parse(), so we start fresh here unless
	// Parse wasn't called immediately before.

	for i, token := range e.Tokens {
		if !e.IsOperating {
			break
		}
		// Update the internal state (tracking slices, remaining words string)
		// before we execute the logic for this token.
		e.UpdateInternalState(i, token)

		// Execute logic
		stop, err := token.Handle(e, i)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	e.IsOperating = true
	return nil
}

// UpdateInternalState handles the maintenance of the tracking slices and strings
// (RemainingRawWords, HandledTokens, RemainingTokens) prior to a token's execution.
func (e *Engine) UpdateInternalState(i int, token Token) {
	// 1. Update RemainingRawWords
	// We want the words AFTER the current one.
	// If this is the last token, the remaining string is empty.
	if i+1 < len(e.RawWords) {
		e.RemainingRawWords = strings.Join(e.RawWords[i+1:], " ")
	} else {
		e.RemainingRawWords = ""
	}

	// 2. Add to Handled list
	e.HandledTokens = append(e.HandledTokens, token)

	// 3. Remove from Remaining list (pop from front)
	if len(e.RemainingTokens) > 0 {
		e.RemainingTokens = e.RemainingTokens[1:]
	}
}
