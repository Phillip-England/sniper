package sniper

import (
	"strconv"
	"strings"
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
func TokenFactory(word string, registry map[string]Cmd, numberPrep *NumberPreprocessor) Token {
	// 1. Run the number preprocessor (e.g. converts "one" -> "1")
	processed := numberPrep.Process(word)

	// 2. Check if the processed word exists in the command registry
	if cmd, ok := registry[processed]; ok {
		return &CmdToken{
			cmd:     cmd,
			literal: processed,
		}
	}

	// 3. Check if it is a number
	if val, err := strconv.Atoi(processed); err == nil {
		return &NumberToken{
			value:   val,
			literal: processed,
		}
	}

	// 4. Default to Raw token
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
	// Check for "Isolated" mode (like "phrase hello world")
	if t.cmd.Mode() == ModeIsolated {
		payload := ""
		// Join everything remaining in RawWords as the payload
		// We use the Engine's RawWords to look ahead
		if index+1 < len(e.RawWords) {
			payload = strings.Join(e.RawWords[index+1:], " ")
		}
		// Execute and return immediately (consumes rest of input)
		return true, t.cmd.Action(e, payload)
	}

	// Execute the standard command once
	if err := t.cmd.Action(e, ""); err != nil {
		return false, err
	}

	// Store this as the previous command for potential repetition
	e.LastCmd = t.cmd
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
	// We only repeat if we have a valid previous command in memory
	if e.LastCmd != nil {
		// The command already ran once when we encountered it.
		// We run it (value - 1) more times.
		if t.value > 1 {
			for k := 0; k < t.value-1; k++ {
				if err := e.LastCmd.Action(e, ""); err != nil {
					return false, err
				}
			}
		}
		// CRITICAL: Wash away the previous action.
		// As per requirements: "left 10 10" -> The second 10 should be skipped.
		e.LastCmd = nil
	}
	// If LastCmd is nil, we simply ignore this number.
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
