package sniper

import (
	"math"
	"time"

	"github.com/go-vgo/robotgo"
)

// Mouse represents the state of the mouse cursor.
type Mouse struct {
	X     int
	Y     int
	Jump  int // Determines how far the mouse moves on directional commands
	Delay time.Duration
}

// NewMouse initializes a new Mouse struct with the current screen position
// and a default Jump value.
func NewMouse() *Mouse {
	x, y := robotgo.Location()
	return &Mouse{
		X:     x,
		Y:     y,
		Jump:  1, // Default jump distance in pixels
		Delay: time.Microsecond * 1,
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

// --- Movement Methods (Using m.Jump with Bounds Checking) ---

// MoveLeft moves the mouse left by the current Jump amount, stopping at the screen edge (0).
func (m *Mouse) MoveLeft() {
	m.SyncPosition()

	targetX := m.X - m.Jump

	// Boundary check: Left edge is 0
	if targetX < 0 {
		targetX = 0
	}

	m.X = targetX
	robotgo.Move(m.X, m.Y)
	time.Sleep(m.Delay)
}

// MoveRight moves the mouse right by the current Jump amount, stopping at the screen width.
func (m *Mouse) MoveRight() {
	m.SyncPosition()

	// Get screen width for boundary check
	screenWidth, _ := robotgo.GetScreenSize()
	targetX := m.X + m.Jump

	// Boundary check: Right edge is screenWidth - 1 (0-indexed)
	if targetX >= screenWidth {
		targetX = screenWidth - 1
	}

	m.X = targetX
	robotgo.Move(m.X, m.Y)
	time.Sleep(m.Delay)
}

// MoveUp moves the mouse up by the current Jump amount, stopping at the top edge (0).
func (m *Mouse) MoveUp() {
	m.SyncPosition()

	targetY := m.Y - m.Jump

	// Boundary check: Top edge is 0
	if targetY < 0 {
		targetY = 0
	}

	m.Y = targetY
	robotgo.Move(m.X, m.Y)
	time.Sleep(m.Delay)
}

// MoveDown moves the mouse down by the current Jump amount, stopping at the screen height.
func (m *Mouse) MoveDown() {
	m.SyncPosition()

	// Get screen height for boundary check
	_, screenHeight := robotgo.GetScreenSize()
	targetY := m.Y + m.Jump

	// Boundary check: Bottom edge is screenHeight - 1 (0-indexed)
	if targetY >= screenHeight {
		targetY = screenHeight - 1
	}

	m.Y = targetY
	robotgo.Move(m.X, m.Y)
	time.Sleep(m.Delay)
}

// --- Click Methods ---

// Click performs a single left click.
func (m *Mouse) Click() {
	robotgo.Click("left")
}

// DoubleClick performs two left clicks with a small delay.
func (m *Mouse) DoubleClick() {
	robotgo.Click("left")
	time.Sleep(m.Delay)
	robotgo.Click("left")
}

// TripleClick performs three left clicks.
func (m *Mouse) TripleClick() {
	robotgo.Click("left")
	time.Sleep(m.Delay)
	robotgo.Click("left")
	time.Sleep(m.Delay)
	robotgo.Click("left")
}

// --- Scrolling Methods ---

// ScrollDown scrolls the screen down.
func (m *Mouse) ScrollDown(amount int) {
	chunkSize := 10
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=0, y=-1 (Usually down on standard OS configs)
		robotgo.Scroll(0, -1)
		time.Sleep(m.Delay)
	}
}

// ScrollUp scrolls the screen up.
func (m *Mouse) ScrollUp(amount int) {
	chunkSize := 10
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=0, y=1 (Usually up)
		robotgo.Scroll(0, 1)
		time.Sleep(m.Delay)
	}
}

// ScrollLeft scrolls the screen left.
func (m *Mouse) ScrollLeft(amount int) {
	chunkSize := 10
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=1, y=0 (Positive X is usually left in robotgo depending on OS)
		// If this scrolls right instead, switch to -1
		robotgo.Scroll(1, 0)
		time.Sleep(m.Delay)
	}
}

// ScrollRight scrolls the screen right.
func (m *Mouse) ScrollRight(amount int) {
	chunkSize := 10
	steps := int(math.Ceil(float64(amount) / float64(chunkSize)))

	for i := 0; i < steps; i++ {
		// x=-1, y=0 (Negative X is usually right in robotgo depending on OS)
		// If this scrolls left instead, switch to 1
		robotgo.Scroll(-1, 0)
		time.Sleep(m.Delay)
	}
}
