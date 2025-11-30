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

// ----------------------------------------------------------------------------
// MODIFIERS
// ----------------------------------------------------------------------------

type Shift struct{}

func (Shift) Mode() ExecutionMode               { return ModeSequential }
func (c Shift) Action(e *Engine, p string) error { e.StickyKeyboard.Shift(); return nil }

type Control struct{}

func (Control) Mode() ExecutionMode               { return ModeSequential }
func (c Control) Action(e *Engine, p string) error { e.StickyKeyboard.Control(); return nil }

type Alt struct{}

func (Alt) Mode() ExecutionMode               { return ModeSequential }
func (c Alt) Action(e *Engine, p string) error { e.StickyKeyboard.Alt(); return nil }

type Command struct{}

func (Command) Mode() ExecutionMode               { return ModeSequential }
func (c Command) Action(e *Engine, p string) error { e.StickyKeyboard.Command(); return nil }

// ----------------------------------------------------------------------------
// NAVIGATION (ARROWS mapped to Cardinals)
// ----------------------------------------------------------------------------

type North struct{} // Up

func (North) Mode() ExecutionMode               { return ModeSequential }
func (c North) Action(e *Engine, p string) error { e.StickyKeyboard.Up(); return nil }

type South struct{} // Down

func (South) Mode() ExecutionMode               { return ModeSequential }
func (c South) Action(e *Engine, p string) error { e.StickyKeyboard.Down(); return nil }

type East struct{} // Right

func (East) Mode() ExecutionMode               { return ModeSequential }
func (c East) Action(e *Engine, p string) error { e.StickyKeyboard.Right(); return nil }

type West struct{} // Left

func (West) Mode() ExecutionMode               { return ModeSequential }
func (c West) Action(e *Engine, p string) error { e.StickyKeyboard.Left(); return nil }

// ----------------------------------------------------------------------------
// EDITING & SPECIAL KEYS
// ----------------------------------------------------------------------------

type Enter struct{}

func (Enter) Mode() ExecutionMode               { return ModeSequential }
func (c Enter) Action(e *Engine, p string) error { e.StickyKeyboard.Enter(); return nil }

type Tab struct{}

func (Tab) Mode() ExecutionMode               { return ModeSequential }
func (c Tab) Action(e *Engine, p string) error { e.StickyKeyboard.Tab(); return nil }

type Space struct{}

func (Space) Mode() ExecutionMode               { return ModeSequential }
func (c Space) Action(e *Engine, p string) error { e.StickyKeyboard.Space(); return nil }

type Back struct{} // Backspace

func (Back) Mode() ExecutionMode               { return ModeSequential }
func (c Back) Action(e *Engine, p string) error { e.StickyKeyboard.Backspace(); return nil }

type Delete struct{}

func (Delete) Mode() ExecutionMode               { return ModeSequential }
func (c Delete) Action(e *Engine, p string) error { e.StickyKeyboard.Delete(); return nil }

type Escape struct{}

func (Escape) Mode() ExecutionMode               { return ModeSequential }
func (c Escape) Action(e *Engine, p string) error { e.StickyKeyboard.Escape(); return nil }

type Home struct{}

func (Home) Mode() ExecutionMode               { return ModeSequential }
func (c Home) Action(e *Engine, p string) error { e.StickyKeyboard.Home(); return nil }

type End struct{}

func (End) Mode() ExecutionMode               { return ModeSequential }
func (c End) Action(e *Engine, p string) error { e.StickyKeyboard.End(); return nil }

type PageUp struct{}

func (PageUp) Mode() ExecutionMode               { return ModeSequential }
func (c PageUp) Action(e *Engine, p string) error { e.StickyKeyboard.PageUp(); return nil }

type PageDown struct{}

func (PageDown) Mode() ExecutionMode               { return ModeSequential }
func (c PageDown) Action(e *Engine, p string) error { e.StickyKeyboard.PageDown(); return nil }

// ----------------------------------------------------------------------------
// SYMBOLS (Single word names)
// ----------------------------------------------------------------------------

type Dot struct{} // .

func (Dot) Mode() ExecutionMode               { return ModeSequential }
func (c Dot) Action(e *Engine, p string) error { e.StickyKeyboard.Period(); return nil }

type Comma struct{} // ,

func (Comma) Mode() ExecutionMode               { return ModeSequential }
func (c Comma) Action(e *Engine, p string) error { e.StickyKeyboard.Comma(); return nil }

type Slash struct{} // /

func (Slash) Mode() ExecutionMode               { return ModeSequential }
func (c Slash) Action(e *Engine, p string) error { e.StickyKeyboard.Slash(); return nil }

type Backslash struct{} // \

func (Backslash) Mode() ExecutionMode               { return ModeSequential }
func (c Backslash) Action(e *Engine, p string) error { e.StickyKeyboard.Backslash(); return nil }

type Semi struct{} // ;

func (Semi) Mode() ExecutionMode               { return ModeSequential }
func (c Semi) Action(e *Engine, p string) error { e.StickyKeyboard.Semicolon(); return nil }

type Quote struct{} // '

func (Quote) Mode() ExecutionMode               { return ModeSequential }
func (c Quote) Action(e *Engine, p string) error { e.StickyKeyboard.Quote(); return nil }

type Bracket struct{} // [

func (Bracket) Mode() ExecutionMode               { return ModeSequential }
func (c Bracket) Action(e *Engine, p string) error { e.StickyKeyboard.BracketLeft(); return nil }

type Closing struct{} // ]

func (Closing) Mode() ExecutionMode               { return ModeSequential }
func (c Closing) Action(e *Engine, p string) error { e.StickyKeyboard.BracketRight(); return nil }

type Dash struct{} // -

func (Dash) Mode() ExecutionMode               { return ModeSequential }
func (c Dash) Action(e *Engine, p string) error { e.StickyKeyboard.Minus(); return nil }

type Equals struct{} // =

func (Equals) Mode() ExecutionMode               { return ModeSequential }
func (c Equals) Action(e *Engine, p string) error { e.StickyKeyboard.Equal(); return nil }

type Tick struct{} // `

func (Tick) Mode() ExecutionMode               { return ModeSequential }
func (c Tick) Action(e *Engine, p string) error { e.StickyKeyboard.Backtick(); return nil }

// ----------------------------------------------------------------------------
// ALPHABET (NATO)
// ----------------------------------------------------------------------------

type Alpha struct{}

func (Alpha) Mode() ExecutionMode               { return ModeSequential }
func (c Alpha) Action(e *Engine, p string) error { e.StickyKeyboard.A(); return nil }

type Bravo struct{}

func (Bravo) Mode() ExecutionMode               { return ModeSequential }
func (c Bravo) Action(e *Engine, p string) error { e.StickyKeyboard.B(); return nil }

type Charlie struct{}

func (Charlie) Mode() ExecutionMode               { return ModeSequential }
func (c Charlie) Action(e *Engine, p string) error { e.StickyKeyboard.C(); return nil }

type Delta struct{}

func (Delta) Mode() ExecutionMode               { return ModeSequential }
func (c Delta) Action(e *Engine, p string) error { e.StickyKeyboard.D(); return nil }

type Echo struct{}

func (Echo) Mode() ExecutionMode               { return ModeSequential }
func (c Echo) Action(e *Engine, p string) error { e.StickyKeyboard.E(); return nil }

type Foxtrot struct{}

func (Foxtrot) Mode() ExecutionMode               { return ModeSequential }
func (c Foxtrot) Action(e *Engine, p string) error { e.StickyKeyboard.F(); return nil }

type Golf struct{}

func (Golf) Mode() ExecutionMode               { return ModeSequential }
func (c Golf) Action(e *Engine, p string) error { e.StickyKeyboard.G(); return nil }

type Hotel struct{}

func (Hotel) Mode() ExecutionMode               { return ModeSequential }
func (c Hotel) Action(e *Engine, p string) error { e.StickyKeyboard.H(); return nil }

type India struct{}

func (India) Mode() ExecutionMode               { return ModeSequential }
func (c India) Action(e *Engine, p string) error { e.StickyKeyboard.I(); return nil }

type Juliet struct{}

func (Juliet) Mode() ExecutionMode               { return ModeSequential }
func (c Juliet) Action(e *Engine, p string) error { e.StickyKeyboard.J(); return nil }

type Kilo struct{}

func (Kilo) Mode() ExecutionMode               { return ModeSequential }
func (c Kilo) Action(e *Engine, p string) error { e.StickyKeyboard.K(); return nil }

type Lima struct{}

func (Lima) Mode() ExecutionMode               { return ModeSequential }
func (c Lima) Action(e *Engine, p string) error { e.StickyKeyboard.L(); return nil }

type Mike struct{}

func (Mike) Mode() ExecutionMode               { return ModeSequential }
func (c Mike) Action(e *Engine, p string) error { e.StickyKeyboard.M(); return nil }

type November struct{}

func (November) Mode() ExecutionMode               { return ModeSequential }
func (c November) Action(e *Engine, p string) error { e.StickyKeyboard.N(); return nil }

type Oscar struct{}

func (Oscar) Mode() ExecutionMode               { return ModeSequential }
func (c Oscar) Action(e *Engine, p string) error { e.StickyKeyboard.O(); return nil }

type Papa struct{}

func (Papa) Mode() ExecutionMode               { return ModeSequential }
func (c Papa) Action(e *Engine, p string) error { e.StickyKeyboard.P(); return nil }

type Quebec struct{}

func (Quebec) Mode() ExecutionMode               { return ModeSequential }
func (c Quebec) Action(e *Engine, p string) error { e.StickyKeyboard.Q(); return nil }

type Romeo struct{}

func (Romeo) Mode() ExecutionMode               { return ModeSequential }
func (c Romeo) Action(e *Engine, p string) error { e.StickyKeyboard.R(); return nil }

type Sierra struct{}

func (Sierra) Mode() ExecutionMode               { return ModeSequential }
func (c Sierra) Action(e *Engine, p string) error { e.StickyKeyboard.S(); return nil }

type Tango struct{}

func (Tango) Mode() ExecutionMode               { return ModeSequential }
func (c Tango) Action(e *Engine, p string) error { e.StickyKeyboard.T(); return nil }

type Uniform struct{}

func (Uniform) Mode() ExecutionMode               { return ModeSequential }
func (c Uniform) Action(e *Engine, p string) error { e.StickyKeyboard.U(); return nil }

type Victor struct{}

func (Victor) Mode() ExecutionMode               { return ModeSequential }
func (c Victor) Action(e *Engine, p string) error { e.StickyKeyboard.V(); return nil }

type Whiskey struct{}

func (Whiskey) Mode() ExecutionMode               { return ModeSequential }
func (c Whiskey) Action(e *Engine, p string) error { e.StickyKeyboard.W(); return nil }

type Xray struct{}

func (Xray) Mode() ExecutionMode               { return ModeSequential }
func (c Xray) Action(e *Engine, p string) error { e.StickyKeyboard.X(); return nil }

type Yankee struct{}

func (Yankee) Mode() ExecutionMode               { return ModeSequential }
func (c Yankee) Action(e *Engine, p string) error { e.StickyKeyboard.Y(); return nil }

type Zulu struct{}

func (Zulu) Mode() ExecutionMode               { return ModeSequential }
func (c Zulu) Action(e *Engine, p string) error { e.StickyKeyboard.Z(); return nil }

// ----------------------------------------------------------------------------
// NUMBERS
// ----------------------------------------------------------------------------

type Zero struct{}

func (Zero) Mode() ExecutionMode               { return ModeSequential }
func (c Zero) Action(e *Engine, p string) error { e.StickyKeyboard.Num0(); return nil }

type One struct{}

func (One) Mode() ExecutionMode               { return ModeSequential }
func (c One) Action(e *Engine, p string) error { e.StickyKeyboard.Num1(); return nil }

type Two struct{}

func (Two) Mode() ExecutionMode               { return ModeSequential }
func (c Two) Action(e *Engine, p string) error { e.StickyKeyboard.Num2(); return nil }

type Three struct{}

func (Three) Mode() ExecutionMode               { return ModeSequential }
func (c Three) Action(e *Engine, p string) error { e.StickyKeyboard.Num3(); return nil }

type Four struct{}

func (Four) Mode() ExecutionMode               { return ModeSequential }
func (c Four) Action(e *Engine, p string) error { e.StickyKeyboard.Num4(); return nil }

type Five struct{}

func (Five) Mode() ExecutionMode               { return ModeSequential }
func (c Five) Action(e *Engine, p string) error { e.StickyKeyboard.Num5(); return nil }

type Six struct{}

func (Six) Mode() ExecutionMode               { return ModeSequential }
func (c Six) Action(e *Engine, p string) error { e.StickyKeyboard.Num6(); return nil }

type Seven struct{}

func (Seven) Mode() ExecutionMode               { return ModeSequential }
func (c Seven) Action(e *Engine, p string) error { e.StickyKeyboard.Num7(); return nil }

type Eight struct{}

func (Eight) Mode() ExecutionMode               { return ModeSequential }
func (c Eight) Action(e *Engine, p string) error { e.StickyKeyboard.Num8(); return nil }

type Nine struct{}

func (Nine) Mode() ExecutionMode               { return ModeSequential }
func (c Nine) Action(e *Engine, p string) error { e.StickyKeyboard.Num9(); return nil }

// ----------------------------------------------------------------------------
// FUNCTION KEYS
// ----------------------------------------------------------------------------

type FOne struct{}

func (FOne) Mode() ExecutionMode               { return ModeSequential }
func (c FOne) Action(e *Engine, p string) error { e.StickyKeyboard.F1(); return nil }

type FTwo struct{}

func (FTwo) Mode() ExecutionMode               { return ModeSequential }
func (c FTwo) Action(e *Engine, p string) error { e.StickyKeyboard.F2(); return nil }

type FThree struct{}

func (FThree) Mode() ExecutionMode               { return ModeSequential }
func (c FThree) Action(e *Engine, p string) error { e.StickyKeyboard.F3(); return nil }

type FFour struct{}

func (FFour) Mode() ExecutionMode               { return ModeSequential }
func (c FFour) Action(e *Engine, p string) error { e.StickyKeyboard.F4(); return nil }

type FFive struct{}

func (FFive) Mode() ExecutionMode               { return ModeSequential }
func (c FFive) Action(e *Engine, p string) error { e.StickyKeyboard.F5(); return nil }

type FSix struct{}

func (FSix) Mode() ExecutionMode               { return ModeSequential }
func (c FSix) Action(e *Engine, p string) error { e.StickyKeyboard.F6(); return nil }

type FSeven struct{}

func (FSeven) Mode() ExecutionMode               { return ModeSequential }
func (c FSeven) Action(e *Engine, p string) error { e.StickyKeyboard.F7(); return nil }

type FEight struct{}

func (FEight) Mode() ExecutionMode               { return ModeSequential }
func (c FEight) Action(e *Engine, p string) error { e.StickyKeyboard.F8(); return nil }

type FNine struct{}

func (FNine) Mode() ExecutionMode               { return ModeSequential }
func (c FNine) Action(e *Engine, p string) error { e.StickyKeyboard.F9(); return nil }

type FTen struct{}

func (FTen) Mode() ExecutionMode               { return ModeSequential }
func (c FTen) Action(e *Engine, p string) error { e.StickyKeyboard.F10(); return nil }

type FEleven struct{}

func (FEleven) Mode() ExecutionMode               { return ModeSequential }
func (c FEleven) Action(e *Engine, p string) error { e.StickyKeyboard.F11(); return nil }

type FTwelve struct{}

func (FTwelve) Mode() ExecutionMode               { return ModeSequential }
func (c FTwelve) Action(e *Engine, p string) error { e.StickyKeyboard.F12(); return nil }

// ----------------------------------------------------------------------------
// MOUSE (Basic)
// ----------------------------------------------------------------------------

type Click struct{}

func (c Click) Name() string                     { return "click" }
func (Click) Mode() ExecutionMode                { return ModeSequential }
func (c Click) Action(e *Engine, p string) error { e.Mouse.Click(); return nil }

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