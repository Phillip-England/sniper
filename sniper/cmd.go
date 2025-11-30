package sniper

// Cmd represents a voice command within the system.
type Cmd interface {

	// Mode tells the engine if this is Sequential or Isolated
	Mode() ExecutionMode

	// Action contains the actual business logic to perform.
	// It accepts the Engine so the command can access state (Keyboard, Tokens, etc.).
	// It also accepts a captured phrase string if the command is Isolated.
	Action(e *Engine, phrase string) error
}

// ExecutionMode defines how the engine handles the command queue
type ExecutionMode int

const (
	// ModeSequential: The command joins the queue and waits its turn.
	ModeSequential ExecutionMode = iota

	// ModeIsolated: The command pauses or clears the queue and runs alone.
	// (e.g., "Stop Listening" or "System Shutdown")
	ModeIsolated
)

// --- Raw Text Handling ---

// RawToken represents a word that was not matched in the command registry.
// It is stored in the Token list so that Isolated commands (like "camel case")
// can access the raw text data contextually.
type RawToken struct {
	Value string
}

func (RawToken) Mode() ExecutionMode { return ModeSequential }

// Action for RawToken is a no-op. It is skipped during standard execution
// but remains available in the e.Tokens slice for other commands to inspect.
func (RawToken) Action(e *Engine, phrase string) error {
	return nil
}

// --- Navigation Commands ---

// Left represents a command to move left.
type Left struct{}

func (Left) Mode() ExecutionMode { return ModeSequential }
func (Left) Action(e *Engine, phrase string) error {
	// Call the mouse method defined in mouse.go
	e.Mouse.MoveLeft()
	return nil
}

// Right represents a command to move right.
type Right struct{}

func (Right) Mode() ExecutionMode { return ModeSequential }
func (Right) Action(e *Engine, phrase string) error {
	// Call the mouse method defined in mouse.go
	e.Mouse.MoveRight()
	return nil
}

// Up represents a command to move up.
type Up struct{}

func (Up) Mode() ExecutionMode { return ModeSequential }
func (Up) Action(e *Engine, phrase string) error {
	// Call the mouse method defined in mouse.go
	e.Mouse.MoveUp()
	return nil
}

// Down represents a command to move down.
type Down struct{}

func (Down) Mode() ExecutionMode { return ModeSequential }
func (Down) Action(e *Engine, phrase string) error {
	// Call the mouse method defined in mouse.go
	e.Mouse.MoveDown()
	return nil
}