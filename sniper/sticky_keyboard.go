package sniper

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-vgo/robotgo"
)

// StickyKeyboard represents a keyboard that remembers modifier keys
// until a non-modifier key is pressed.
type StickyKeyboard struct {
	// pendingModifiers holds keys like "shift", "command" waiting for the next keystroke
	pendingModifiers []string

	// mu protects the pendingModifiers slice for thread safety
	mu sync.Mutex

	// PostReleaseDelay is the time to sleep after keys are released
	// to ensure the OS registers the state change.
	PostReleaseDelay time.Duration
}

// NewStickyKeyboard initializes the keyboard structure.
func NewStickyKeyboard() *StickyKeyboard {
	return &StickyKeyboard{
		pendingModifiers: make([]string, 0),
		PostReleaseDelay: 5 * time.Millisecond, // Adjustable delay
	}
}

// ----------------------------------------------------------------------------
// INTERNAL LOGIC
// ----------------------------------------------------------------------------

// queueModifier adds a modifier to the memory. It acts as the "Hold" phase.
// It detects OS differences (Command vs Control) automatically.
func (k *StickyKeyboard) queueModifier(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Normalize modifiers based on OS
	normalizedKey := key
	if runtime.GOOS == "darwin" {
		switch key {
		case "command":
			normalizedKey = "cmd"
		case "option":
			normalizedKey = "lalt" // left alt usually maps to option
		case "control":
			normalizedKey = "lctrl"
		}
	} else {
		// Windows/Linux mapping
		if key == "command" {
			normalizedKey = "control" // standard mapping for windows users using mac terms
		}
	}

	// Prevent duplicates (e.g., calling Shift twice shouldn't add it twice)
	for _, m := range k.pendingModifiers {
		if m == normalizedKey {
			return
		}
	}

	k.pendingModifiers = append(k.pendingModifiers, normalizedKey)
	fmt.Printf("[Keyboard] Modifier Queued: %s\n", normalizedKey)
}

// executeTap performs the actual robotgo action.
// It applies all pending modifiers, taps the target key, and then
// explicitly ensures all modifiers are released.
func (k *StickyKeyboard) executeTap(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Convert string slice to interface slice for robotgo
	args := make([]interface{}, len(k.pendingModifiers))
	for i, v := range k.pendingModifiers {
		args[i] = v
	}

	if len(args) > 0 {
		fmt.Printf("[Keyboard] Tapping '%s' with modifiers: %v\n", key, args)
	} else {
		fmt.Printf("[Keyboard] Tapping '%s'\n", key)
	}

	// RobotGo KeyTap holds the modifiers (args) and taps the key.
	robotgo.KeyTap(key, args...)

	// EXPLICIT SAFETY RELEASE
	// Even though KeyTap attempts to release modifiers, we explicitly
	// force a KeyUp on every pending modifier to prevent "stuck key" states.
	for _, mod := range k.pendingModifiers {
		robotgo.KeyUp(mod)
	}

	// Clear memory immediately after execution
	k.pendingModifiers = []string{}

	// CRITICAL: The small delay requested to ensure OS registers the release
	time.Sleep(k.PostReleaseDelay)
}

// ----------------------------------------------------------------------------
// MODIFIER METHODS
// Calling these does NOT press the key immediately. It adds them to memory.
// ----------------------------------------------------------------------------

func (k *StickyKeyboard) Shift()   { k.queueModifier("shift") }
func (k *StickyKeyboard) Command() { k.queueModifier("command") }
func (k *StickyKeyboard) Control() { k.queueModifier("ctrl") }
func (k *StickyKeyboard) Alt()     { k.queueModifier("alt") }
func (k *StickyKeyboard) Option()  { k.queueModifier("option") } // Alias for Alt/Mac specific

// ----------------------------------------------------------------------------
// STANDARD KEY METHODS
// Calling these consumes any pending modifiers.
// ----------------------------------------------------------------------------

// --- Letters ---
func (k *StickyKeyboard) A() { k.executeTap("a") }
func (k *StickyKeyboard) B() { k.executeTap("b") }
func (k *StickyKeyboard) C() { k.executeTap("c") }
func (k *StickyKeyboard) D() { k.executeTap("d") }
func (k *StickyKeyboard) E() { k.executeTap("e") }
func (k *StickyKeyboard) F() { k.executeTap("f") }
func (k *StickyKeyboard) G() { k.executeTap("g") }
func (k *StickyKeyboard) H() { k.executeTap("h") }
func (k *StickyKeyboard) I() { k.executeTap("i") }
func (k *StickyKeyboard) J() { k.executeTap("j") }
func (k *StickyKeyboard) K() { k.executeTap("k") }
func (k *StickyKeyboard) L() { k.executeTap("l") }
func (k *StickyKeyboard) M() { k.executeTap("m") }
func (k *StickyKeyboard) N() { k.executeTap("n") }
func (k *StickyKeyboard) O() { k.executeTap("o") }
func (k *StickyKeyboard) P() { k.executeTap("p") }
func (k *StickyKeyboard) Q() { k.executeTap("q") }
func (k *StickyKeyboard) R() { k.executeTap("r") }
func (k *StickyKeyboard) S() { k.executeTap("s") }
func (k *StickyKeyboard) T() { k.executeTap("t") }
func (k *StickyKeyboard) U() { k.executeTap("u") }
func (k *StickyKeyboard) V() { k.executeTap("v") }
func (k *StickyKeyboard) W() { k.executeTap("w") }
func (k *StickyKeyboard) X() { k.executeTap("x") }
func (k *StickyKeyboard) Y() { k.executeTap("y") }
func (k *StickyKeyboard) Z() { k.executeTap("z") }

// --- Numbers (Single Digits) ---
func (k *StickyKeyboard) Num0() { k.executeTap("0") }
func (k *StickyKeyboard) Num1() { k.executeTap("1") }
func (k *StickyKeyboard) Num2() { k.executeTap("2") }
func (k *StickyKeyboard) Num3() { k.executeTap("3") }
func (k *StickyKeyboard) Num4() { k.executeTap("4") }
func (k *StickyKeyboard) Num5() { k.executeTap("5") }
func (k *StickyKeyboard) Num6() { k.executeTap("6") }
func (k *StickyKeyboard) Num7() { k.executeTap("7") }
func (k *StickyKeyboard) Num8() { k.executeTap("8") }
func (k *StickyKeyboard) Num9() { k.executeTap("9") }

// --- Expanded Number Support ---

// TypeInt types out the string representation of an integer.
// For example, TypeInt(100) taps '1', then '0', then '0'.
// Note: Pending modifiers (like Shift) will only apply to the FIRST digit typed
// because executeTap clears modifiers after the first press.
func (k *StickyKeyboard) TypeInt(n int) {
	str := strconv.Itoa(n)
	for _, char := range str {
		k.executeTap(string(char))
	}
}

// TypeStr types out any string character by character.
func (k *StickyKeyboard) TypeStr(s string) {
	for _, char := range s {
		k.executeTap(string(char))
	}
}

// --- Casing Methods ---

// CamelCase converts "hello world" to "helloWorld"
func (k *StickyKeyboard) CamelCase(phrase string) {
	words := strings.Fields(phrase)
	for i, w := range words {
		// Ensure word is lower case first
		runes := []rune(strings.ToLower(w))
		if len(runes) == 0 {
			continue
		}
		// If it's not the first word, capitalize the first letter
		if i > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	// Join with empty string
	k.TypeStr(strings.Join(words, ""))
}

// PascalCase converts "hello world" to "HelloWorld"
func (k *StickyKeyboard) PascalCase(phrase string) {
	words := strings.Fields(phrase)
	for i, w := range words {
		runes := []rune(strings.ToLower(w))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	k.TypeStr(strings.Join(words, ""))
}

// SnakeCase converts "Hello World" to "hello_world"
func (k *StickyKeyboard) SnakeCase(phrase string) {
	words := strings.Fields(phrase)
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	k.TypeStr(strings.Join(words, "_"))
}

// Sentence types the words like a standard sentence (e.g., "Hello world")
func (k *StickyKeyboard) Sentence(phrase string) error {
	if len(phrase) == 0 {
		return nil
	}

	phrase += ". "
	// Capitalize the very first letter of the sentence
	runes := []rune(phrase)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return k.Type(string(runes))
}

// Type uses robotgo to type the final formatted string string.
func (k *StickyKeyboard) Type(text string) error {
	// Robotgo does not return an error, so we wrap it.
	robotgo.TypeStr(text)
	return nil
}

// --- Function Keys ---
func (k *StickyKeyboard) F1()  { k.executeTap("f1") }
func (k *StickyKeyboard) F2()  { k.executeTap("f2") }
func (k *StickyKeyboard) F3()  { k.executeTap("f3") }
func (k *StickyKeyboard) F4()  { k.executeTap("f4") }
func (k *StickyKeyboard) F5()  { k.executeTap("f5") }
func (k *StickyKeyboard) F6()  { k.executeTap("f6") }
func (k *StickyKeyboard) F7()  { k.executeTap("f7") }
func (k *StickyKeyboard) F8()  { k.executeTap("f8") }
func (k *StickyKeyboard) F9()  { k.executeTap("f9") }
func (k *StickyKeyboard) F10() { k.executeTap("f10") }
func (k *StickyKeyboard) F11() { k.executeTap("f11") }
func (k *StickyKeyboard) F12() { k.executeTap("f12") }

// --- Navigation & Edit ---
func (k *StickyKeyboard) Enter()     { k.executeTap("enter") }
func (k *StickyKeyboard) Tab()       { k.executeTap("tab") }
func (k *StickyKeyboard) Space()     { k.executeTap("space") }
func (k *StickyKeyboard) Backspace() { k.executeTap("backspace") }
func (k *StickyKeyboard) Delete()    { k.executeTap("delete") }
func (k *StickyKeyboard) Escape()    { k.executeTap("escape") }
func (k *StickyKeyboard) Left()      { k.executeTap("left") }
func (k *StickyKeyboard) Right()     { k.executeTap("right") }
func (k *StickyKeyboard) Up()        { k.executeTap("up") }
func (k *StickyKeyboard) Down()      { k.executeTap("down") }
func (k *StickyKeyboard) Home()      { k.executeTap("home") }
func (k *StickyKeyboard) End()       { k.executeTap("end") }
func (k *StickyKeyboard) PageUp()    { k.executeTap("pageup") }
func (k *StickyKeyboard) PageDown()  { k.executeTap("pagedown") }

// --- Punctuation ---
func (k *StickyKeyboard) Period()       { k.executeTap(".") }
func (k *StickyKeyboard) Comma()        { k.executeTap(",") }
func (k *StickyKeyboard) Slash()        { k.executeTap("/") }
func (k *StickyKeyboard) Backslash()    { k.executeTap("\\") }
func (k *StickyKeyboard) Semicolon()    { k.executeTap(";") }
func (k *StickyKeyboard) Quote()        { k.executeTap("'") }
func (k *StickyKeyboard) BracketLeft()  { k.executeTap("[") }
func (k *StickyKeyboard) BracketRight() { k.executeTap("]") }
func (k *StickyKeyboard) Minus()        { k.executeTap("-") }
func (k *StickyKeyboard) Equal()        { k.executeTap("=") }
func (k *StickyKeyboard) Backtick()     { k.executeTap("`") }
