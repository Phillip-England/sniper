package sniper

import (
	"strings"
	"time"
)

// EngineState holds the transient state for a single parse/execute cycle.
type EngineState struct {
	Tokens            []Token
	RemainingTokens   []Token
	HandledTokens     []Token
	RemainingRawWords string
	TokenIndices      []int
	RawWords          []string
	LastCmd           Cmd
	FirstCmdIsValid   bool
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
	Delay          time.Duration

	// State holds the transient data for the current command sequence.
	State *EngineState

	// LastState holds the state of the PREVIOUS successful execution.
	// Used for the "repeat" command.
	LastState *EngineState

	IsOperating bool
	RawInput    string
}

func NewEngine() *Engine {
	e := &Engine{
		StickyKeyboard: NewStickyKeyboard(),
		registry:       make(map[string]Cmd),
		Mouse:          NewMouse(),
		Delay:          time.Microsecond * 800,
		State:          nil,
		LastState:      nil,
		IsOperating:    true, // Fixed: Default to true
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

func (e *Engine) Parse(input string) {
	// 0. HISTORY MANAGEMENT
	// Before we wipe e.State for the new input, we decide if we should
	// save the current e.State into e.LastState.
	// We only save if:
	// 1. We have a state to save.
	// 2. The input that generated that state was NOT "repeat".
	//    (If we repeat "repeat", we want to execute the thing before that, not "repeat" itself).
	if e.State != nil && !strings.Contains(strings.ToLower(e.RawInput), "repeat") {
		e.LastState = e.State
	}

	e.RawInput = input

	// Initialize a fresh state structure
	s := &EngineState{
		LastCmd:         nil,
		FirstCmdIsValid: false,
	}

	input = strings.ToLower(input)
	rawInput := strings.Fields(input)

	s.Tokens = make([]Token, 0, len(rawInput))
	s.TokenIndices = make([]int, 0, len(rawInput))
	s.RawWords = make([]string, 0, len(rawInput))

	for i, word := range rawInput {
		token := TokenFactory(word, e.registry)
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

	for i, token := range e.State.Tokens {
		if !e.IsOperating {
			break
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

// UpdateInternalState is kept for backward compatibility if used elsewhere,
// but Execute() now uses State.Advance directly.
func (e *Engine) UpdateInternalState(i int, token Token) {
	e.State.Advance(i, token)
}
