package sniper

import "strings"

// Cmd represents a voice command within the system.
type Cmd interface {
	// Name returns the unique string identifier for the command (e.g., "shift", "a").
	Name() string

	// CalledBy returns the slice of strings that trigger the command.
	CalledBy() []string

	// Effects returns a list of middleware to run for this command.
	Effects() []EffectFunc

	// Action contains the actual business logic to perform.
	Action(e *Engine, phrase string) error
}

// ----------------------------------------------------------------------------
// MODIFIERS
// ----------------------------------------------------------------------------

type Shift struct{}

func (Shift) Name() string          { return "shift" }
func (Shift) CalledBy() []string    { return []string{"shift"} }
func (Shift) Effects() []EffectFunc { return nil }
func (c Shift) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Shift()
		return nil
	}, c.Effects()...)
}

type Control struct{}

func (Control) Name() string          { return "control" }
func (Control) CalledBy() []string    { return []string{"control"} }
func (Control) Effects() []EffectFunc { return nil }
func (c Control) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control()
		return nil
	}, c.Effects()...)
}

type Alt struct{}

func (Alt) Name() string          { return "alt" }
func (Alt) CalledBy() []string    { return []string{"alt"} }
func (Alt) Effects() []EffectFunc { return nil }
func (c Alt) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Alt()
		return nil
	}, c.Effects()...)
}

type Command struct{}

func (Command) Name() string          { return "command" }
func (Command) CalledBy() []string    { return []string{"command"} }
func (Command) Effects() []EffectFunc { return nil }
func (c Command) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Command()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// NAVIGATION (ARROWS mapped to Cardinals)
// ----------------------------------------------------------------------------

type North struct{} // Up

func (North) Name() string          { return "north" }
func (North) CalledBy() []string    { return []string{"north"} }
func (North) Effects() []EffectFunc { return nil }
func (c North) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Up()
		return nil
	}, c.Effects()...)
}

type South struct{} // Down

func (South) Name() string          { return "south" }
func (South) CalledBy() []string    { return []string{"south"} }
func (South) Effects() []EffectFunc { return nil }
func (c South) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Down()
		return nil
	}, c.Effects()...)
}

type East struct{} // Right

func (East) Name() string          { return "east" }
func (East) CalledBy() []string    { return []string{"east"} }
func (East) Effects() []EffectFunc { return nil }
func (c East) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Right()
		return nil
	}, c.Effects()...)
}

type West struct{} // Left

func (West) Name() string          { return "west" }
func (West) CalledBy() []string    { return []string{"west"} }
func (West) Effects() []EffectFunc { return nil }
func (c West) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Left()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// EDITING & SPECIAL KEYS
// ----------------------------------------------------------------------------

type Enter struct{}

func (Enter) Name() string          { return "enter" }
func (Enter) CalledBy() []string    { return []string{"enter"} }
func (Enter) Effects() []EffectFunc { return nil }
func (c Enter) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Enter()
		return nil
	}, c.Effects()...)
}

type Tab struct{}

func (Tab) Name() string          { return "tab" }
func (Tab) CalledBy() []string    { return []string{"tab"} }
func (Tab) Effects() []EffectFunc { return nil }
func (c Tab) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Tab()
		return nil
	}, c.Effects()...)
}

type Space struct{}

func (Space) Name() string          { return "space" }
func (Space) CalledBy() []string    { return []string{"space", "next"} }
func (Space) Effects() []EffectFunc { return nil }
func (c Space) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Space()
		return nil
	}, c.Effects()...)
}

type Back struct{} // Backspace

func (Back) Name() string          { return "back" }
func (Back) CalledBy() []string    { return []string{"back"} }
func (Back) Effects() []EffectFunc { return nil }
func (c Back) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Backspace()
		return nil
	}, c.Effects()...)
}

type Delete struct{}

func (Delete) Name() string          { return "delete" }
func (Delete) CalledBy() []string    { return []string{"delete"} }
func (Delete) Effects() []EffectFunc { return nil }
func (c Delete) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Delete()
		return nil
	}, c.Effects()...)
}

type Escape struct{}

func (Escape) Name() string          { return "escape" }
func (Escape) CalledBy() []string    { return []string{"escape"} }
func (Escape) Effects() []EffectFunc { return nil }
func (c Escape) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Escape()
		return nil
	}, c.Effects()...)
}

type Home struct{}

func (Home) Name() string          { return "home" }
func (Home) CalledBy() []string    { return []string{"home"} }
func (Home) Effects() []EffectFunc { return nil }
func (c Home) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Home()
		return nil
	}, c.Effects()...)
}

type End struct{}

func (End) Name() string          { return "end" }
func (End) CalledBy() []string    { return []string{"end"} }
func (End) Effects() []EffectFunc { return nil }
func (c End) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.End()
		return nil
	}, c.Effects()...)
}

type PageUp struct{}

func (PageUp) Name() string          { return "page_up" }
func (PageUp) CalledBy() []string    { return []string{"ascend"} }
func (PageUp) Effects() []EffectFunc { return nil }
func (c PageUp) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.PageUp()
		return nil
	}, c.Effects()...)
}

type PageDown struct{}

func (PageDown) Name() string          { return "page_down" }
func (PageDown) CalledBy() []string    { return []string{"descend"} }
func (PageDown) Effects() []EffectFunc { return nil }
func (c PageDown) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.PageDown()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// SYMBOLS (Single word names)
// ----------------------------------------------------------------------------

type Dot struct{} // .

func (Dot) Name() string          { return "dot" }
func (Dot) CalledBy() []string    { return []string{"dot", "."} }
func (Dot) Effects() []EffectFunc { return nil }
func (c Dot) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Period()
		return nil
	}, c.Effects()...)
}

type Comma struct{} // ,

func (Comma) Name() string          { return "comma" }
func (Comma) CalledBy() []string    { return []string{"comma"} }
func (Comma) Effects() []EffectFunc { return nil }
func (c Comma) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Comma()
		return nil
	}, c.Effects()...)
}

type Slash struct{} // /

func (Slash) Name() string          { return "slash" }
func (Slash) CalledBy() []string    { return []string{"slash", "/"} }
func (Slash) Effects() []EffectFunc { return nil }
func (c Slash) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Slash()
		return nil
	}, c.Effects()...)
}

type Backslash struct{} // \

func (Backslash) Name() string          { return "backslash" }
func (Backslash) CalledBy() []string    { return []string{"backslash"} }
func (Backslash) Effects() []EffectFunc { return nil }
func (c Backslash) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Backslash()
		return nil
	}, c.Effects()...)
}

type Semi struct{} // ;

func (Semi) Name() string          { return "semi" }
func (Semi) CalledBy() []string    { return []string{"semi"} }
func (Semi) Effects() []EffectFunc { return nil }
func (c Semi) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Semicolon()
		return nil
	}, c.Effects()...)
}

type Quote struct{} // '

func (Quote) Name() string          { return "quote" }
func (Quote) CalledBy() []string    { return []string{"quote"} }
func (Quote) Effects() []EffectFunc { return nil }
func (c Quote) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Quote()
		return nil
	}, c.Effects()...)
}

type Bracket struct{} // [

func (Bracket) Name() string          { return "bracket" }
func (Bracket) CalledBy() []string    { return []string{"bracket"} }
func (Bracket) Effects() []EffectFunc { return nil }
func (c Bracket) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.BracketLeft()
		return nil
	}, c.Effects()...)
}

type Closing struct{} // ]

func (Closing) Name() string          { return "closing" }
func (Closing) CalledBy() []string    { return []string{"closing"} }
func (Closing) Effects() []EffectFunc { return nil }
func (c Closing) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.BracketRight()
		return nil
	}, c.Effects()...)
}

type Dash struct{} // -

func (Dash) Name() string          { return "dash" }
func (Dash) CalledBy() []string    { return []string{"dash", "-"} }
func (Dash) Effects() []EffectFunc { return nil }
func (c Dash) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Minus()
		return nil
	}, c.Effects()...)
}

type Equals struct{} // =

func (Equals) Name() string          { return "equals" }
func (Equals) CalledBy() []string    { return []string{"equals", "="} }
func (Equals) Effects() []EffectFunc { return nil }
func (c Equals) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Equal()
		return nil
	}, c.Effects()...)
}

type Tick struct{} // `

func (Tick) Name() string          { return "tick" }
func (Tick) CalledBy() []string    { return []string{"tick"} }
func (Tick) Effects() []EffectFunc { return nil }
func (c Tick) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Backtick()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// ALPHABET (NATO)
// ----------------------------------------------------------------------------

type A struct{}

func (A) Name() string          { return "a" }
func (A) CalledBy() []string    { return []string{"alpha", "a"} }
func (A) Effects() []EffectFunc { return nil }
func (c A) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.A()
		return nil
	}, c.Effects()...)
}

type B struct{}

func (B) Name() string          { return "b" }
func (B) CalledBy() []string    { return []string{"bravo", "b"} }
func (B) Effects() []EffectFunc { return nil }
func (c B) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.B()
		return nil
	}, c.Effects()...)
}

type C struct{}

func (C) Name() string          { return "c" }
func (C) CalledBy() []string    { return []string{"charlie", "c"} }
func (C) Effects() []EffectFunc { return nil }
func (c C) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.C()
		return nil
	}, c.Effects()...)
}

type D struct{}

func (D) Name() string          { return "d" }
func (D) CalledBy() []string    { return []string{"delta", "d"} }
func (D) Effects() []EffectFunc { return nil }
func (c D) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.D()
		return nil
	}, c.Effects()...)
}

type E struct{}

func (E) Name() string          { return "e" }
func (E) CalledBy() []string    { return []string{"echo", "e"} }
func (E) Effects() []EffectFunc { return nil }
func (c E) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.E()
		return nil
	}, c.Effects()...)
}

type F struct{}

func (F) Name() string          { return "f" }
func (F) CalledBy() []string    { return []string{"foxtrot", "f"} }
func (F) Effects() []EffectFunc { return nil }
func (c F) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F()
		return nil
	}, c.Effects()...)
}

type G struct{}

func (G) Name() string          { return "g" }
func (G) CalledBy() []string    { return []string{"golf", "g"} }
func (G) Effects() []EffectFunc { return nil }
func (c G) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.G()
		return nil
	}, c.Effects()...)
}

type H struct{}

func (H) Name() string          { return "h" }
func (H) CalledBy() []string    { return []string{"hotel", "h"} }
func (H) Effects() []EffectFunc { return nil }
func (c H) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.H()
		return nil
	}, c.Effects()...)
}

type I struct{}

func (I) Name() string          { return "i" }
func (I) CalledBy() []string    { return []string{"india", "i"} }
func (I) Effects() []EffectFunc { return nil }
func (c I) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.I()
		return nil
	}, c.Effects()...)
}

type J struct{}

func (J) Name() string          { return "j" }
func (J) CalledBy() []string    { return []string{"juliet", "j"} }
func (J) Effects() []EffectFunc { return nil }
func (c J) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.J()
		return nil
	}, c.Effects()...)
}

type K struct{}

func (K) Name() string          { return "k" }
func (K) CalledBy() []string    { return []string{"kilo", "k"} }
func (K) Effects() []EffectFunc { return nil }
func (c K) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.K()
		return nil
	}, c.Effects()...)
}

type L struct{}

func (L) Name() string          { return "l" }
func (L) CalledBy() []string    { return []string{"lima", "l"} }
func (L) Effects() []EffectFunc { return nil }
func (c L) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.L()
		return nil
	}, c.Effects()...)
}

type M struct{}

func (M) Name() string          { return "m" }
func (M) CalledBy() []string    { return []string{"mike", "m"} }
func (M) Effects() []EffectFunc { return nil }
func (c M) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.M()
		return nil
	}, c.Effects()...)
}

type N struct{}

func (N) Name() string          { return "n" }
func (N) CalledBy() []string    { return []string{"november", "n", "in"} }
func (N) Effects() []EffectFunc { return nil }
func (c N) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.N()
		return nil
	}, c.Effects()...)
}

type O struct{}

func (O) Name() string          { return "o" }
func (O) CalledBy() []string    { return []string{"oscar", "o"} }
func (O) Effects() []EffectFunc { return nil }
func (c O) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.O()
		return nil
	}, c.Effects()...)
}

type P struct{}

func (P) Name() string          { return "p" }
func (P) CalledBy() []string    { return []string{"papa", "p"} }
func (P) Effects() []EffectFunc { return nil }
func (c P) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.P()
		return nil
	}, c.Effects()...)
}

type Q struct{}

func (Q) Name() string          { return "q" }
func (Q) CalledBy() []string    { return []string{"quebec", "q"} }
func (Q) Effects() []EffectFunc { return nil }
func (c Q) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Q()
		return nil
	}, c.Effects()...)
}

type R struct{}

func (R) Name() string          { return "r" }
func (R) CalledBy() []string    { return []string{"romeo", "r"} }
func (R) Effects() []EffectFunc { return nil }
func (c R) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.R()
		return nil
	}, c.Effects()...)
}

type S struct{}

func (S) Name() string          { return "s" }
func (S) CalledBy() []string    { return []string{"sierra", "s"} }
func (S) Effects() []EffectFunc { return nil }
func (c S) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.S()
		return nil
	}, c.Effects()...)
}

type T struct{}

func (T) Name() string          { return "t" }
func (T) CalledBy() []string    { return []string{"tango", "t"} }
func (T) Effects() []EffectFunc { return nil }
func (c T) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.T()
		return nil
	}, c.Effects()...)
}

type U struct{}

func (U) Name() string          { return "u" }
func (U) CalledBy() []string    { return []string{"uniform", "u"} }
func (U) Effects() []EffectFunc { return nil }
func (c U) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.U()
		return nil
	}, c.Effects()...)
}

type V struct{}

func (V) Name() string          { return "v" }
func (V) CalledBy() []string    { return []string{"victor", "v"} }
func (V) Effects() []EffectFunc { return nil }
func (c V) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.V()
		return nil
	}, c.Effects()...)
}

type W struct{}

func (W) Name() string          { return "w" }
func (W) CalledBy() []string    { return []string{"whiskey", "w"} }
func (W) Effects() []EffectFunc { return nil }
func (c W) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.W()
		return nil
	}, c.Effects()...)
}

type X struct{}

func (X) Name() string          { return "x" }
func (X) CalledBy() []string    { return []string{"xray", "x"} }
func (X) Effects() []EffectFunc { return nil }
func (c X) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.X()
		return nil
	}, c.Effects()...)
}

type Y struct{}

func (Y) Name() string          { return "y" }
func (Y) CalledBy() []string    { return []string{"yankee", "y"} }
func (Y) Effects() []EffectFunc { return nil }
func (c Y) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Y()
		return nil
	}, c.Effects()...)
}

type Z struct{}

func (Z) Name() string          { return "z" }
func (Z) CalledBy() []string    { return []string{"zulu", "z"} }
func (Z) Effects() []EffectFunc { return nil }
func (c Z) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Z()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// NUMBERS
// ----------------------------------------------------------------------------

type Zero struct{}

func (Zero) Name() string          { return "zero" }
func (Zero) CalledBy() []string    { return []string{"zero"} }
func (Zero) Effects() []EffectFunc { return nil }
func (c Zero) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num0()
		return nil
	}, c.Effects()...)
}

type One struct{}

func (One) Name() string          { return "one" }
func (One) CalledBy() []string    { return []string{"one"} }
func (One) Effects() []EffectFunc { return nil }
func (c One) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num1()
		return nil
	}, c.Effects()...)
}

type Two struct{}

func (Two) Name() string          { return "two" }
func (Two) CalledBy() []string    { return []string{"two"} }
func (Two) Effects() []EffectFunc { return nil }
func (c Two) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num2()
		return nil
	}, c.Effects()...)
}

type Three struct{}

func (Three) Name() string          { return "three" }
func (Three) CalledBy() []string    { return []string{"three"} }
func (Three) Effects() []EffectFunc { return nil }
func (c Three) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num3()
		return nil
	}, c.Effects()...)
}

type Four struct{}

func (Four) Name() string          { return "four" }
func (Four) CalledBy() []string    { return []string{"four"} }
func (Four) Effects() []EffectFunc { return nil }
func (c Four) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num4()
		return nil
	}, c.Effects()...)
}

type Five struct{}

func (Five) Name() string          { return "five" }
func (Five) CalledBy() []string    { return []string{"five"} }
func (Five) Effects() []EffectFunc { return nil }
func (c Five) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num5()
		return nil
	}, c.Effects()...)
}

type Six struct{}

func (Six) Name() string          { return "six" }
func (Six) CalledBy() []string    { return []string{"six"} }
func (Six) Effects() []EffectFunc { return nil }
func (c Six) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num6()
		return nil
	}, c.Effects()...)
}

type Seven struct{}

func (Seven) Name() string          { return "seven" }
func (Seven) CalledBy() []string    { return []string{"seven"} }
func (Seven) Effects() []EffectFunc { return nil }
func (c Seven) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num7()
		return nil
	}, c.Effects()...)
}

type Eight struct{}

func (Eight) Name() string          { return "eight" }
func (Eight) CalledBy() []string    { return []string{"eight"} }
func (Eight) Effects() []EffectFunc { return nil }
func (c Eight) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num8()
		return nil
	}, c.Effects()...)
}

type Nine struct{}

func (Nine) Name() string          { return "nine" }
func (Nine) CalledBy() []string    { return []string{"nine"} }
func (Nine) Effects() []EffectFunc { return nil }
func (c Nine) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Num9()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// FUNCTION KEYS
// ----------------------------------------------------------------------------

type FOne struct{}

func (FOne) Name() string          { return "f1" }
func (FOne) CalledBy() []string    { return []string{"f1"} }
func (FOne) Effects() []EffectFunc { return nil }
func (c FOne) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F1()
		return nil
	}, c.Effects()...)
}

type FTwo struct{}

func (FTwo) Name() string          { return "f2" }
func (FTwo) CalledBy() []string    { return []string{"f2"} }
func (FTwo) Effects() []EffectFunc { return nil }
func (c FTwo) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F2()
		return nil
	}, c.Effects()...)
}

type FThree struct{}

func (FThree) Name() string          { return "f3" }
func (FThree) CalledBy() []string    { return []string{"f3"} }
func (FThree) Effects() []EffectFunc { return nil }
func (c FThree) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F3()
		return nil
	}, c.Effects()...)
}

type FFour struct{}

func (FFour) Name() string          { return "f4" }
func (FFour) CalledBy() []string    { return []string{"f4"} }
func (FFour) Effects() []EffectFunc { return nil }
func (c FFour) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F4()
		return nil
	}, c.Effects()...)
}

type FFive struct{}

func (FFive) Name() string          { return "f5" }
func (FFive) CalledBy() []string    { return []string{"f5"} }
func (FFive) Effects() []EffectFunc { return nil }
func (c FFive) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F5()
		return nil
	}, c.Effects()...)
}

type FSix struct{}

func (FSix) Name() string          { return "f6" }
func (FSix) CalledBy() []string    { return []string{"f6"} }
func (FSix) Effects() []EffectFunc { return nil }
func (c FSix) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F6()
		return nil
	}, c.Effects()...)
}

type FSeven struct{}

func (FSeven) Name() string          { return "f7" }
func (FSeven) CalledBy() []string    { return []string{"f7"} }
func (FSeven) Effects() []EffectFunc { return nil }
func (c FSeven) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F7()
		return nil
	}, c.Effects()...)
}

type FEight struct{}

func (FEight) Name() string          { return "f8" }
func (FEight) CalledBy() []string    { return []string{"f8"} }
func (FEight) Effects() []EffectFunc { return nil }
func (c FEight) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F8()
		return nil
	}, c.Effects()...)
}

type FNine struct{}

func (FNine) Name() string          { return "f9" }
func (FNine) CalledBy() []string    { return []string{"f9"} }
func (FNine) Effects() []EffectFunc { return nil }
func (c FNine) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F9()
		return nil
	}, c.Effects()...)
}

type FTen struct{}

func (FTen) Name() string          { return "f10" }
func (FTen) CalledBy() []string    { return []string{"f10"} }
func (FTen) Effects() []EffectFunc { return nil }
func (c FTen) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F10()
		return nil
	}, c.Effects()...)
}

type FEleven struct{}

func (FEleven) Name() string          { return "f11" }
func (FEleven) CalledBy() []string    { return []string{"f11"} }
func (FEleven) Effects() []EffectFunc { return nil }
func (c FEleven) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F11()
		return nil
	}, c.Effects()...)
}

type FTwelve struct{}

func (FTwelve) Name() string          { return "f12" }
func (FTwelve) CalledBy() []string    { return []string{"f12"} }
func (FTwelve) Effects() []EffectFunc { return nil }
func (c FTwelve) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.F12()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// MOUSE (Basic)
// ----------------------------------------------------------------------------

type Click struct{}

func (c Click) Name() string        { return "click" }
func (c Click) CalledBy() []string  { return []string{"click"} }
func (Click) Effects() []EffectFunc { return []EffectFunc{WaitAfter(50)} }
func (c Click) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.Mouse.Click()
		return nil
	}, c.Effects()...)
}

// Left represents a command to move the mouse left.
type Left struct{}

func (Left) Name() string          { return "mouse_left" }
func (Left) CalledBy() []string    { return []string{"left"} }
func (Left) Effects() []EffectFunc { return nil }
func (Left) Action(e *Engine, phrase string) error {
	return EffectChain(e, func() error {
		e.Mouse.MoveLeft()
		return nil
	}, nil...) // nil checks are safe in spread
}

// Right represents a command to move the mouse right.
type Right struct{}

func (Right) Name() string          { return "mouse_right" }
func (Right) CalledBy() []string    { return []string{"right", "write"} }
func (Right) Effects() []EffectFunc { return nil }
func (Right) Action(e *Engine, phrase string) error {
	return EffectChain(e, func() error {
		e.Mouse.MoveRight()
		return nil
	}, nil...)
}

// Up represents a command to move the mouse up.
type Up struct{}

func (Up) Name() string          { return "mouse_up" }
func (Up) CalledBy() []string    { return []string{"up"} }
func (Up) Effects() []EffectFunc { return nil }
func (Up) Action(e *Engine, phrase string) error {
	return EffectChain(e, func() error {
		e.Mouse.MoveUp()
		return nil
	}, nil...)
}

// Down represents a command to move the mouse down.
type Down struct{}

func (Down) Name() string          { return "mouse_down" }
func (Down) CalledBy() []string    { return []string{"down"} }
func (Down) Effects() []EffectFunc { return nil }
func (Down) Action(e *Engine, phrase string) error {
	return EffectChain(e, func() error {
		e.Mouse.MoveDown()
		return nil
	}, nil...)
}

// ----------------------------------------------------------------------------
// TEXT FORMATTING & SPEECH
// ----------------------------------------------------------------------------

type RawType struct{}

func (RawType) Name() string          { return "raw_type" }
func (RawType) CalledBy() []string    { return []string{"type"} }
func (RawType) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c RawType) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// 1. Get the raw text following the "type" command
		text := e.State.RemainingRawWords

		// 2. Smash the input together (remove all spaces)
		// e.g., "type a b c" -> "abc"
		text = strings.ReplaceAll(text, " ", "")

		// 3. Type the resulting string literal
		e.StickyKeyboard.TypeStr(text)

		return nil
	}, c.Effects()...)
}

// CamelCase converts the subsequent phrase into camelCase (e.g., "myVariableName").
type CamelCase struct{}

func (CamelCase) Name() string          { return "camel_case" }
func (CamelCase) CalledBy() []string    { return []string{"camel"} }
func (CamelCase) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c CamelCase) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Pass the remaining spoken words to the keyboard's Camel handler
		e.StickyKeyboard.CamelCase(e.State.RemainingRawWords)
		return nil
	}, c.Effects()...)
}

// PascalCase converts the subsequent phrase into PascalCase (e.g., "MyVariableName").
type PascalCase struct{}

func (PascalCase) Name() string          { return "pascal_case" }
func (PascalCase) CalledBy() []string    { return []string{"pascal"} }
func (PascalCase) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c PascalCase) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Pass the remaining spoken words to the keyboard's Pascal handler
		e.StickyKeyboard.PascalCase(e.State.RemainingRawWords)
		return nil
	}, c.Effects()...)
}

// SnakeCase converts the subsequent phrase into snake_case (e.g., "my_variable_name").
type SnakeCase struct{}

func (SnakeCase) Name() string          { return "snake_case" }
func (SnakeCase) CalledBy() []string    { return []string{"snake"} }
func (SnakeCase) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c SnakeCase) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Pass the remaining spoken words to the keyboard's Snake handler
		e.StickyKeyboard.SnakeCase(e.State.RemainingRawWords)
		return nil
	}, c.Effects()...)
}

// Say types out the subsequent phrase formatted as a sentence.
type Say struct{}

func (Say) Name() string          { return "say" }
func (Say) CalledBy() []string    { return []string{"say"} }
func (Say) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c Say) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Pass the remaining spoken words to the keyboard's Sentence handler
		e.StickyKeyboard.Sentence(e.State.RemainingRawWords)
		return nil
	}, c.Effects()...)
}

type Number struct{}

func (Number) Name() string          { return "number" }
func (Number) CalledBy() []string    { return []string{"number"} }
func (Number) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c Number) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// 1. Check if there is a next token to look at
		if len(e.State.RemainingTokens) > 0 {
			nextToken := e.State.RemainingTokens[0]

			// 2. Check if the next token is a number
			if nextToken.Type() == TokenTypeNumber {
				// 3. Manually type out the number literal
				e.StickyKeyboard.TypeStr(nextToken.Literal())
			}
		}

		// If it wasn't a number, or there were no tokens left,
		// we essentially do nothing (skip).
		return nil
	}, c.Effects()...)
}

// Word types the single immediate next word and ignores the rest.
// e.g. "word git commit" -> types "git" (ignores "commit")
type Word struct{}

func (Word) Name() string          { return "word" }
func (Word) CalledBy() []string    { return []string{"word"} }
func (Word) Effects() []EffectFunc { return []EffectFunc{KillAfter()} }
func (c Word) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// 1. Get the text following the "word" command
		// e.g. "git status"
		text := e.State.RemainingRawWords

		// 2. Split the text into individual words
		words := strings.Fields(text)

		// 3. If there is at least one word, type only the first one
		if len(words) > 0 {
			e.StickyKeyboard.TypeStr(words[0])
		}

		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// SHORTCUTS (Combos)
// ----------------------------------------------------------------------------

// Copy performs Control+C.
type Copy struct{}

func (Copy) Name() string          { return "copy" }
func (Copy) CalledBy() []string    { return []string{"copy"} }
func (Copy) Effects() []EffectFunc { return nil }
func (c Copy) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.C()       // Press C
		return nil
	}, c.Effects()...)
}

// Select performs Control+A (Select All).
type Select struct{}

func (Select) Name() string          { return "select" }
func (Select) CalledBy() []string    { return []string{"select", "select all"} }
func (Select) Effects() []EffectFunc { return nil }
func (c Select) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.A()       // Press A
		return nil
	}, c.Effects()...)
}

// Paste performs Control+V.
type Paste struct{}

func (Paste) Name() string          { return "paste" }
func (Paste) CalledBy() []string    { return []string{"paste"} }
func (Paste) Effects() []EffectFunc { return nil }
func (c Paste) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.V()       // Press V
		return nil
	}, c.Effects()...)
}

// Telescope performs Control+P.
type Telescope struct{}

func (Telescope) Name() string          { return "telescope" }
func (Telescope) CalledBy() []string    { return []string{"telescope"} }
func (Telescope) Effects() []EffectFunc { return nil }
func (c Telescope) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.P()       // Press P
		return nil
	}, c.Effects()...)
}

type Find struct{}

func (Find) Name() string          { return "find" }
func (Find) CalledBy() []string    { return []string{"find"} }
func (Find) Effects() []EffectFunc { return []EffectFunc{ClickBefore()} }
func (c Find) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.F()       // Press P
		return nil
	}, c.Effects()...)
}

type DeleteWord struct{}

func (DeleteWord) Name() string          { return "delete_word" }
func (DeleteWord) CalledBy() []string    { return []string{"oops"} }
func (DeleteWord) Effects() []EffectFunc { return []EffectFunc{} }
func (c DeleteWord) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control()   // Hold Control
		e.StickyKeyboard.Backspace() // Press P
		return nil
	}, c.Effects()...)
}

// Grab clicks the mouse (to focus), Selects All, and then Copies.
type Save struct{}

func (Save) Name() string       { return "yank" }
func (Save) CalledBy() []string { return []string{"save", "safe"} }

// Uses the new ClickBefore effect
func (Save) Effects() []EffectFunc { return []EffectFunc{} }
func (c Save) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Logic: Ctrl (Hold) -> A (Select All) -> C (Copy) -> Ctrl (Release)
		e.StickyKeyboard.Control()
		e.StickyKeyboard.S()
		return nil
	}, c.Effects()...)
}

// Undo performs Control+Z.
type Undo struct{}

func (Undo) Name() string          { return "undo" }
func (Undo) CalledBy() []string    { return []string{"undo", "reverse"} }
func (Undo) Effects() []EffectFunc { return nil }
func (c Undo) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		e.StickyKeyboard.Control() // Hold Control
		e.StickyKeyboard.Z()       // Press Z
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// ADVANCED ACTIONS (Grab & Shove)
// ----------------------------------------------------------------------------

// Grab clicks the mouse (to focus), Selects All, and then Copies.
type Grab struct{}

func (Grab) Name() string       { return "grab" }
func (Grab) CalledBy() []string { return []string{"grab"} }

// Uses the new ClickBefore effect
func (Grab) Effects() []EffectFunc { return []EffectFunc{ClickBefore(), ClickAfter()} }
func (c Grab) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Logic: Ctrl (Hold) -> A (Select All) -> C (Copy) -> Ctrl (Release)
		e.StickyKeyboard.Control()
		e.StickyKeyboard.A()
		e.StickyKeyboard.Control()
		e.StickyKeyboard.C()
		return nil
	}, c.Effects()...)
}

// Grab clicks the mouse (to focus), Selects All, and then Copies.
type Yank struct{}

func (Yank) Name() string       { return "yank" }
func (Yank) CalledBy() []string { return []string{"yank"} }

// Uses the new ClickBefore effect
func (Yank) Effects() []EffectFunc { return []EffectFunc{ClickBefore()} }
func (c Yank) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Logic: Ctrl (Hold) -> A (Select All) -> C (Copy) -> Ctrl (Release)
		e.StickyKeyboard.Control()
		e.StickyKeyboard.A()
		e.StickyKeyboard.Control()
		e.StickyKeyboard.X()
		return nil
	}, c.Effects()...)
}

// Shove clicks the mouse (to focus) and then Pastes.
type Shove struct{}

func (Shove) Name() string       { return "shove" }
func (Shove) CalledBy() []string { return []string{"shove"} }

// Uses the new ClickBefore effect
func (Shove) Effects() []EffectFunc { return []EffectFunc{ClickBefore()} }
func (c Shove) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// Logic: Ctrl (Hold) -> V (Paste) -> Ctrl (Release)
		e.StickyKeyboard.Control()
		e.StickyKeyboard.V()
		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// HISTORY COMMANDS
// ----------------------------------------------------------------------------

type Repeat struct{}

func (Repeat) Name() string          { return "repeat" }
func (Repeat) CalledBy() []string    { return []string{"repeat", "again"} }
func (Repeat) Effects() []EffectFunc { return nil }
func (c Repeat) Action(e *Engine, p string) error {
	return EffectChain(e, func() error {
		// 1. Check if we have history
		if e.LastState == nil {
			return nil
		}

		// 2. We need to construct a fresh state based on LastState.
		// We cannot reuse LastState directly because 'Advance' consumes it
		// (removes RemainingTokens). If we consume it, we can't repeat twice.
		replayState := &EngineState{
			Tokens:       e.LastState.Tokens,
			TokenIndices: e.LastState.TokenIndices,
			RawWords:     e.LastState.RawWords,
			// Fresh tracking slices:
			HandledTokens:   make([]Token, 0, len(e.LastState.Tokens)),
			RemainingTokens: make([]Token, len(e.LastState.Tokens)),
		}
		// Copy tokens into remaining
		copy(replayState.RemainingTokens, e.LastState.Tokens)
		replayState.RemainingRawWords = strings.Join(e.LastState.RawWords, " ")

		// 3. Swap the Engine State
		currentState := e.State // Backup current state ("repeat")
		e.State = replayState   // Swap in the replay

		// 4. Execute (This will run the logic of the previous commands)
		if err := e.Execute(); err != nil {
			return err
		}

		// 5. Restore State (Optional, but good practice to leave engine clean)
		e.State = currentState

		return nil
	}, c.Effects()...)
}

// ----------------------------------------------------------------------------
// COMMAND REGISTRY
// ----------------------------------------------------------------------------

// Registry contains a slice of all available commands to be used elsewhere.
var Registry = []Cmd{
	// Modifiers
	Shift{}, Control{}, Alt{}, Command{},

	// Navigation
	North{}, South{}, East{}, West{},

	// Editing
	Enter{}, Tab{}, Space{}, Back{}, Delete{}, Escape{},
	Home{}, End{}, PageUp{}, PageDown{},

	// Symbols
	Dot{}, Comma{}, Slash{}, Backslash{}, Semi{}, Quote{},
	Bracket{}, Closing{}, Dash{}, Equals{}, Tick{},

	// Alphabet
	A{}, B{}, C{}, D{}, E{}, F{},
	G{}, H{}, I{}, J{}, K{}, L{},
	M{}, N{}, O{}, P{}, Q{}, R{},
	S{}, T{}, U{}, V{}, W{}, X{},
	Y{}, Z{},

	// Numbers
	Number{},
	Zero{}, One{}, Two{}, Three{}, Four{},
	Five{}, Six{}, Seven{}, Eight{}, Nine{},

	// Function Keys
	FOne{}, FTwo{}, FThree{}, FFour{}, FFive{}, FSix{},
	FSeven{}, FEight{}, FNine{}, FTen{}, FEleven{}, FTwelve{},

	// Mouse
	Click{}, Left{}, Right{}, Up{}, Down{},

	// Formatting
	CamelCase{}, PascalCase{}, SnakeCase{}, Say{}, RawType{}, Word{},

	// SHORTCUTS (New Combos)
	Copy{}, Select{}, Paste{}, Telescope{}, Undo{}, Save{},

	// ADVANCED ACTIONS (New Click+Combo)
	Grab{}, Shove{}, Find{}, DeleteWord{}, Yank{},

	// HISTORY
	Repeat{},
}
