package sniper

import (
	"strconv"
	"strings"
	"time"
)

type ExecutonMode string

const (
	ModeRapid  ExecutonMode = "RAPID"
	ModePhrase ExecutonMode = "PHRASE"
)

// EngineState holds the transient state for a single parse/execute cycle.
type EngineState struct {
	ExecutionMode     ExecutonMode
	Tokens            []Token
	RemainingTokens   []Token
	HandledTokens     []Token
	RemainingRawWords string
	TokenIndices      []int
	RawWords          []string
	LastCmd           Cmd
	FirstCmdIsValid   bool
	ConsumedArgs      []string // Stores words like "banana" consumed by commands
	SkipCount         int      // How many tokens to skip in the main loop
}

// Advance updates the tracking slices and strings for the current execution step.
func (s *EngineState) Advance(i int, token Token) {
	// 1. Update RemainingRawWords
	if i+1 < len(s.RawWords) {
		s.RemainingRawWords = strings.Join(s.RawWords[i+1:], " ")
	} else {
		s.RemainingRawWords = ""
	}

	// 2. Add to Handled list
	s.HandledTokens = append(s.HandledTokens, token)

	// 3. Remove from Remaining list (pop from front)
	if len(s.RemainingTokens) > 0 {
		s.RemainingTokens = s.RemainingTokens[1:]
	}
}

type Engine struct {
	StickyKeyboard *StickyKeyboard
	registry       map[string]Cmd
	Mouse          *Mouse
	Memory         *MouseMemory // New: Persistence layer
	Delay          time.Duration

	State     *EngineState
	LastState *EngineState

	IsOperating bool
	RawInput    string
}

func NewEngine() *Engine {
	e := &Engine{
		StickyKeyboard: NewStickyKeyboard(),
		registry:       make(map[string]Cmd),
		Mouse:          NewMouse(),
		Memory:         NewMouseMemory(), // Initialize Memory
		Delay:          time.Microsecond * 800,
		State:          nil,
		LastState:      nil,
		IsOperating:    true,
	}

	e.registerCommands()
	return e
}

func (e *Engine) registerCommands() {
	for _, cmd := range Registry {
		for _, trigger := range cmd.CalledBy() {
			key := strings.ToLower(trigger)
			e.registry[key] = cmd
		}
	}
}

func (e *Engine) Parse(input string, mode string) {
	if e.State != nil && !strings.Contains(strings.ToLower(e.RawInput), "repeat") {
		e.LastState = e.State
	}

	e.RawInput = input

	var executionMode ExecutonMode
	if mode == "rapid" {
		executionMode = ModeRapid
	}
	if mode == "phrase" {
		executionMode = ModePhrase
	}
	s := &EngineState{
		LastCmd:         nil,
		FirstCmdIsValid: false,
		ConsumedArgs:    make([]string, 0),
		SkipCount:       0,
		ExecutionMode:   executionMode,
	}

	input = strings.ToLower(input)
	rawInput := strings.Fields(input)

	s.Tokens = make([]Token, 0, len(rawInput))
	s.TokenIndices = make([]int, 0, len(rawInput))
	s.RawWords = make([]string, 0, len(rawInput))

	for i, word := range rawInput {
		// Pass e.Memory to TokenFactory so we can recognize saved spots
		token := TokenFactory(word, e.registry, e.Memory)
		s.Tokens = append(s.Tokens, token)
		s.RawWords = append(s.RawWords, token.Literal())
		s.TokenIndices = append(s.TokenIndices, i)

		if i == 0 && token.Type() == TokenTypeCmd {
			s.FirstCmdIsValid = true
		}
	}

	s.HandledTokens = make([]Token, 0, len(s.Tokens))
	s.RemainingTokens = make([]Token, len(s.Tokens))
	copy(s.RemainingTokens, s.Tokens)
	s.RemainingRawWords = strings.Join(s.RawWords, " ")

	e.State = s
}

func (e *Engine) Execute() error {
	if e.State == nil {
		return nil
	}

	if e.State.ExecutionMode == ModePhrase {
		err := e.handlePhraseMode()
		if err != nil {
			return err
		}
		e.IsOperating = true
		return nil
	}

	if e.State.ExecutionMode == ModeRapid {
		// handle rapid execution
		lastTok := e.State.Tokens[len(e.State.Tokens)-1]

		// handling regular commands
		if lastTok.Type() == 1 {
			shouldStop, err := lastTok.Handle(e, 0)
			if err != nil {
				return err
			}
			if shouldStop {
				e.IsOperating = false
			}
		}

		// handling numbers
		if lastTok.Type() == 2 {
			amt, err := strconv.Atoi(lastTok.Literal())
			if err != nil {
				return err
			}
			prevTok := e.LastState.Tokens[len(e.LastState.Tokens)-1]
			amt = amt - 1
			for {
				if amt <= 0 {
					break
				}
				shouldStop, err := prevTok.Handle(e, 0)
				if err != nil {
					return err
				}
				if shouldStop {
					e.IsOperating = false
				}
				amt -= 1
			}

		}

		// handling raw value
		if lastTok.Type() == 0 {
			// skip for now..
		}

	}

	return nil
}

func (e *Engine) handlePhraseMode() error {
	for i, token := range e.State.Tokens {
		if !e.IsOperating {
			break
		}

		// 1. Check if we need to skip this token (because it was consumed as an argument)
		if e.State.SkipCount > 0 {
			e.State.SkipCount--
			// We still need to advance internal state tracking for accuracy
			e.State.Advance(i, token)
			continue
		}

		e.State.Advance(i, token)

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

func (e *Engine) UpdateInternalState(i int, token Token) {
	e.State.Advance(i, token)
}
