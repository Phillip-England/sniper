package sniper

import (
	"strconv"
)

// TokenType identifies the category of a token.
type TokenType int

const (
	TokenTypeRaw TokenType = iota
	TokenTypeCmd
	TokenTypeNumber
)

// Token is the interface that all token types must implement.
type Token interface {
	Type() TokenType
	Literal() string
	// Handle executes the logic specific to this token type.
	// It returns a bool indicating if execution should stop (true), and an error.
	Handle(e *Engine, index int) (bool, error)
}

// TokenFactory takes a raw string word, processes it, and returns the appropriate Token.
// UPDATED: Now accepts MouseMemory to check for dynamic spots.
func TokenFactory(word string, registry map[string]Cmd, memory *MouseMemory) Token {
	// 1. Run the number preprocessor
	numberPrep := NewNumberPreprocessor()
	processed := numberPrep.Process(word)

	// 2. Check Registry (Static Commands)
	if cmd, ok := registry[processed]; ok {
		return &CmdToken{
			cmd:     cmd,
			literal: processed,
		}
	}

	// 3. Check Mouse Memory (Dynamic Spots)
	// If the word matches a saved spot, we create a dynamic command to move there.
	if spot, ok := memory.Get(processed); ok {
		return &CmdToken{
			cmd:     NewSpotCmd(processed, spot.X, spot.Y),
			literal: processed,
		}
	}

	// 4. Check Number
	if val, err := strconv.Atoi(processed); err == nil {
		return &NumberToken{
			value:   val,
			literal: processed,
		}
	}

	// 5. Default to Raw token
	return &RawToken{
		literal: processed,
	}
}

// --- Token Implementations ---

// CmdToken represents a valid command found in the registry.
type CmdToken struct {
	cmd     Cmd
	literal string
}

func (t *CmdToken) Type() TokenType { return TokenTypeCmd }
func (t *CmdToken) Literal() string { return t.literal }
func (t *CmdToken) Command() Cmd    { return t.cmd }

func (t *CmdToken) Handle(e *Engine, index int) (bool, error) {
	// Execute the standard command once
	if err := t.cmd.Action(e, ""); err != nil {
		return false, err
	}

	// Store this as the previous command for potential repetition
	e.State.LastCmd = t.cmd
	return false, nil
}

// NumberToken represents a numeric value.
type NumberToken struct {
	value   int
	literal string
}

func (t *NumberToken) Type() TokenType { return TokenTypeNumber }
func (t *NumberToken) Literal() string { return t.literal }
func (t *NumberToken) Value() int      { return t.value }

func (t *NumberToken) Handle(e *Engine, index int) (bool, error) {
	// CASE 1: Intra-phrase Repetition (e.g., "Left 5")
	// We have a valid command in the CURRENT sequence history.
	if e.State.LastCmd != nil {
		// The command already ran once. Run it (value - 1) more times.
		if t.value > 1 {
			for k := 0; k < t.value-1; k++ {
				if err := e.State.LastCmd.Action(e, ""); err != nil {
					return false, err
				}
			}
		}
		// Consume the LastCmd so "Left 10 10" doesn't cascade.
		e.State.LastCmd = nil
		return false, nil
	}

	// CASE 2: Inter-phrase Repetition (e.g., User said "Left Down", then says "5")
	// There is no command in the current sequence, and Parse has preserved LastState.
	if e.LastState != nil && len(e.LastState.Tokens) > 0 {
		// We repeat the entire sequence 't.value' times.
		for k := 0; k < t.value; k++ {
			for _, prevToken := range e.LastState.Tokens {

				// SAFETY CHECK: Prevent infinite recursion.
				// If the previous sequence contained a Number, we only allow it
				// if it has a valid LastCmd to act on. If it's a "naked" number, skip it.
				// This prevents "2" -> "2" from exploding if State was wiped.
				if prevToken.Type() == TokenTypeNumber && e.State.LastCmd == nil {
					continue
				}

				// Execute the token.
				// We pass -1 as index because strict indexing doesn't matter for replay.
				_, err := prevToken.Handle(e, -1)
				if err != nil {
					return false, err
				}
			}
		}
	}

	return false, nil
}

// RawToken represents input that is neither a command nor a number.
type RawToken struct {
	literal string
}

func (t *RawToken) Type() TokenType { return TokenTypeRaw }
func (t *RawToken) Literal() string { return t.literal }

func (t *RawToken) Handle(e *Engine, index int) (bool, error) {
	// Currently, raw input that isn't a command or number is ignored
	// to preserve original functionality, but this handler exists for future expansion.
	return false, nil
}
