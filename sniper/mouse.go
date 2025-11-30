package sniper

import (
	"math"
	"time"

	"github.com/go-vgo/robotgo"
)

// Mouse represents the state of the mouse cursor.
type Mouse struct {
	X    int
	Y    int
	Jump int // Determines how far the mouse moves on directional commands
}

// NewMouse initializes a new Mouse struct with the current screen position
// and a default Jump value.
func NewMouse() *Mouse {
	x, y := robotgo.Location()
	return &Mouse{
		X:    x,
		Y:    y,
		Jump: 50, // Default jump distance in pixels
	}
}

// SyncPosition updates the internal X and Y coordinates to match the actual system mouse position.
func (m *Mouse) SyncPosition() {
	x, y := robotgo.Location()
	m.X = x
	m.Y = y
}

// SetJump allows you to update the distance the mouse moves.
func (m *Mouse) SetJump(pixels int) {
	m.Jump = pixels
}

// --- Movement Methods (Using m.Jump) ---

// MoveLeft moves the mouse left by the current Jump amount.
func (m *Mouse) MoveLeft() {
	m.SyncPosition()
	m.X -= m.Jump
	robotgo.Move(m.X, m.Y)
}

// MoveRight moves the mouse right by the current Jump amount.
func (m *Mouse) MoveRight() {
	m.SyncPosition()
	m.X += m.Jump
	robotgo.Move(m.X, m.Y)
}

// MoveUp moves the mouse up by the current Jump amount.
func (m *Mouse) MoveUp() {
	m.SyncPosition()
	m.Y -= m.Jump
	robotgo.Move(m.X, m.Y)
}

// MoveDown moves the mouse down by the current Jump amount.
func (m *Mouse) MoveDown() {
	m.SyncPosition()
	m.Y += m.Jump
	robotgo.Move(m.X, m.Y)
}

// --- Click Methods ---

// Click performs a single left click.
func (m *Mouse) Click() {
	robotgo.Click("left")
}

// DoubleClick performs two left clicks with a small delay.
func (m *Mouse) DoubleClick() {
	robotgo.Click("left")
	time.Sleep(100 * time.Millisecond)
	robotgo.Click("left")
}

// TripleClick performs three left clicks.
func (m *Mouse) TripleClick() {
	robotgo.Click("left")
	time.Sleep(100 * time.Millisecond)
	robotgo.Click("left")
	time.Sleep(100 * time.Millisecond)
	robotgo.Click("left")
}

// --- Scrolling Methods ---

// ScrollDown scrolls the screen down.
func (m *Mouse) ScrollDown(amount int) {
	chunkSize := 10
	delay := 20 * time.Millisecond
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=0, y=-1 (Usually down on standard OS configs)
		robotgo.Scroll(0, -1)
		time.Sleep(delay)
	}
}

// ScrollUp scrolls the screen up.
func (m *Mouse) ScrollUp(amount int) {
	chunkSize := 10
	delay := 20 * time.Millisecond
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=0, y=1 (Usually up)
		robotgo.Scroll(0, 1)
		time.Sleep(delay)
	}
}

// ScrollLeft scrolls the screen left.
func (m *Mouse) ScrollLeft(amount int) {
	chunkSize := 10
	delay := 20 * time.Millisecond
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=1, y=0 (Positive X is usually left in robotgo depending on OS)
		// If this scrolls right instead, switch to -1
		robotgo.Scroll(1, 0)
		time.Sleep(delay)
	}
}

// ScrollRight scrolls the screen right.
func (m *Mouse) ScrollRight(amount int) {
	chunkSize := 10
	delay := 20 * time.Millisecond
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=-1, y=0 (Negative X is usually right in robotgo depending on OS)
		// If this scrolls left instead, switch to 1
		robotgo.Scroll(-1, 0)
		time.Sleep(delay)
	}
}