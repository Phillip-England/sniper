package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Phillip-England/vii"
	"github.com/go-vgo/robotgo"
)


const (
	CmdTypeAction   string = "ACTION"
	CmdTypeModifier string = "MODIFIER"
)

type Cmd interface {
	Type() string
	Name() string
	Action() error
}

// --- CONFIGURATION ---

const (
	ClientPort    = "3000"
	ServerPort    = "8000"
	// ThrottleMs removed
	MouseDistance = 50
	HistoryFile   = "history.json"
)

// --- INTERFACES & TYPES ---

// ActionExecutor abstracts the physical side effects (RobotGo)
type ActionExecutor interface {
	MoveMouse(x, y int)
	GetMousePos() (int, int)
	TapKey(key string, modifiers []interface{})
	ToggleKey(key string, direction string)
	Click(btn string)
	TypeStr(str string)
}

// RobotGoExecutor is the concrete implementation of ActionExecutor
type RobotGoExecutor struct{}

func (r *RobotGoExecutor) MoveMouse(x, y int)               { robotgo.Move(x, y) }
func (r *RobotGoExecutor) GetMousePos() (int, int)          { return robotgo.Location() }
func (r *RobotGoExecutor) TapKey(k string, m []interface{}) { robotgo.KeyTap(k, m...) }
func (r *RobotGoExecutor) ToggleKey(k string, d string)     { robotgo.KeyToggle(k, d) }
func (r *RobotGoExecutor) Click(btn string)                 { robotgo.Click(btn) }
func (r *RobotGoExecutor) TypeStr(str string)               { robotgo.TypeStr(str) }

// --- DOMAIN DATA ---

var (
	// Phonetic Alphabet Mapping
	phoneticMap = map[string]string{
		"alpha": "a", "bravo": "b", "charlie": "c", "delta": "d",
		"echo": "e", "foxtrot": "f", "golf": "g", "hotel": "h",
		"india": "i", "juliet": "j", "kilo": "k", "lima": "l",
		"mike": "m", "november": "n", "oscar": "o", "papa": "p",
		"quebec": "q", "romeo": "r", "sierra": "s", "tango": "t",
		"uniform": "u", "victor": "v", "whiskey": "w", "xray": "x",
		"yankee": "y", "zulu": "z",
	}

	// Spanish Number Mapping (Direct Typing)
	spanishMap = map[string]string{
		"cero": "0", "uno": "1", "dos": "2", "tres": "3",
		"cuatro": "4", "quatro": "4", "cinco": "5", "seis": "6",
		"siete": "7", "ocho": "8", "nueve": "9", "diez": "10",
	}

	// Symbol / Special Character Mapping
	symbolMap = map[string]string{
		// Row 1 (Shift + Numbers)
		"tilde": "~", "wave": "~",
		"tick": "`",
		"bang": "!",
		"at": "@",
		"hash": "#",
		"dollar": "$",
		"percent": "%",
		"caret": "^", "peak": "^",
		"ampersand": "&", "and": "&",
		"star": "*",
		"open": "(",
		"close": ")",
		"underscore": "_", "floor": "_",
		"plus": "+",
		"minus": "-",

		// Row 2/3 (Brackets, Slashes, etc)
		"equal": "=",
		"brace": "{", "curly": "{",
		"lock": "}", "rcurly": "}", // "lock" preserved for backward compat
		"bracket": "[",
		"kit": "]", "rbracket": "]", // "kit" preserved for backward compat
		"pipe": "|", "bar": "|",
		"backslash": "\\", // "back" removed to avoid conflict with backspace/direction
		"slash": "/",

		// Punctuation
		"colon": ":",
		"semicolon": ";",
		"quote": "\"",
		"single": "'",
		"less": "<",
		"greater": ">", "great": ">",
		"comma": ",", "tail": ",",
		"dot": ".",
		"question": "?",
	}

	// Digit Mapping (For Accumulator Logic)
	digitMap = map[string]string{
		"zero": "0", "one": "1", "won": "1", "two": "2", "to": "2", "too": "2",
		"three": "3", "four": "4", "for": "4", "five": "5",
		"six": "6", "seven": "7", "eight": "8", "nine": "9", "ten": "10",
	}

	// Valid Commands Whitelist
	validCommands = map[string]bool{
		// Directions / Mouse
		"left": true, "right": true, "up": true, "down": true,
		"west": true, "east": true, "north": true, "south": true,
		"click": true, "exit": true,
		// Modifiers
		"control": true, "command": true, "option": true, "alt": true, "shift": true,
		"release": true,
		// Typing Controls
		"backspace": true, "space": true, "enter": true, "tab": true, "escape": true,
		"number": true, "back": true, // Added "back" explicitly
	}
)

func init() {
	// Merge maps into valid commands
	for k := range phoneticMap {
		validCommands[k] = true
	}
	for k := range symbolMap {
		validCommands[k] = true
	}
	for k := range spanishMap {
		validCommands[k] = true
	}
}

// --- STATE MANAGEMENT ---

type SessionState struct {
	mu                sync.Mutex
	// lastSuccessTime removed
	pendingModifiers  []string
	lastVerb          string
	numberAccumulator string
	executedCount     int
}

func NewSessionState() *SessionState {
	return &SessionState{
		pendingModifiers: []string{},
	}
}

// CheckThrottle method removed entirely

// --- CORE ENGINE ---

type SniperEngine struct {
	state    *SessionState
	history  *HistoryManager
	executor ActionExecutor
}

func NewSniperEngine(hist *HistoryManager, exec ActionExecutor) *SniperEngine {
	return &SniperEngine{
		state:    NewSessionState(),
		history:  hist,
		executor: exec,
	}
}

// ProcessInput acts as the controller that splits multi-word commands
// and processes them sequentially.
func (e *SniperEngine) ProcessInput(rawInput string) (bool, string) {
	// Throttle check removed

	cleanedInput := e.cleanInput(rawInput)
	if cleanedInput == "" {
		return false, "empty"
	}

	// 1. History Lookup (Expand the macro if it exists)
	// e.g., if "eagle" -> "alpha bravo", we want to process "alpha bravo"
	if historicCmd, found := e.history.FindCommand(cleanedInput); found {
		fmt.Printf("[History] Trigger '%s' -> Executing '%s'\n", cleanedInput, historicCmd)
		cleanedInput = historicCmd
	}

	// 2. Split into words
	words := strings.Fields(cleanedInput)
	if len(words) == 0 {
		return false, "empty"
	}

	// 3. Iterate and Process
	e.state.mu.Lock()
	defer e.state.mu.Unlock()

	overallSuccess := false
	lastReason := ""

	for _, word := range words {
		success, reason := e.processOneWord(word)
		if success {
			overallSuccess = true
		} else {
			lastReason = reason
		}
	}

	// 4. Update History
	// We push the FULL expanded command line to history if at least one part succeeded
	if overallSuccess {
		e.history.Push(cleanedInput)
		return true, ""
	}

	return false, lastReason
}

// processOneWord handles a single token (e.g., "alpha" or "five" or "down")
func (e *SniperEngine) processOneWord(word string) (bool, string) {
	// Identify Type: Digit or Command
	digitStr, isNumber := e.getDigitString(word)

	if isNumber {
		return e.handleDigitSequence(digitStr)
	}

	return e.handleCommand(word)
}

func (e *SniperEngine) handleDigitSequence(digitStr string) (bool, string) {
	if e.state.lastVerb == "" {
		return false, "no_context"
	}

	// MODE: Typing Numbers (e.g., "number five")
	if e.state.lastVerb == "number" {
		e.executor.TypeStr(digitStr)
		return true, ""
	}

	// MODE: Repetition (e.g., "down five")
	if len(e.state.numberAccumulator) >= 2 {
		return true, "limit_reached"
	}

	e.state.numberAccumulator += digitStr
	targetTotal, _ := strconv.Atoi(e.state.numberAccumulator)
	delta := targetTotal - e.state.executedCount

	if delta > 0 {
		e.executeActionLocked(e.state.lastVerb, delta)
		e.state.executedCount += delta
	}
	return true, ""
}

func (e *SniperEngine) handleCommand(command string) (bool, string) {
	if !validCommands[command] {
		return false, "invalid"
	}

	// Reset Accumulators
	e.state.lastVerb = command
	e.state.numberAccumulator = ""
	e.state.executedCount = 1

	e.executeActionLocked(command, 1)

	return true, ""
}

func (e *SniperEngine) executeActionLocked(verb string, count int) {
	fmt.Printf("[Execute] %s x %d | Pending Mods: %v\n", verb, count, e.state.pendingModifiers)

	// 1. Modifier Handling
	if e.isModifier(verb) {
		e.queueModifier(verb)
		return
	}

	// 2. Prepare Modifiers
	currentMods := make([]interface{}, len(e.state.pendingModifiers))
	for i, v := range e.state.pendingModifiers {
		currentMods[i] = v
	}
	cleanupMods := make([]string, len(e.state.pendingModifiers))
	copy(cleanupMods, e.state.pendingModifiers)

	// 3. Execution Loop
	for i := 0; i < count; i++ {
		switch verb {
		case "number":
			// No-op: "number" is purely a context verb for the next digit input
		case "space":
			e.executor.TapKey("space", currentMods)
		case "back":
			e.executor.TapKey("backspace", currentMods)
		case "enter":
			e.executor.TapKey("enter", currentMods)
		case "tab":
			e.executor.TapKey("tab", currentMods)
		case "escape":
			e.executor.TapKey("escape", currentMods)
		case "release":
			// No-op, cleanup handles release
		case "left", "right", "up", "down":
			e.handleMouse(verb, MouseDistance)
		case "west", "east", "north", "south":
			e.handleDirectionalKeys(verb, currentMods)
		case "click":
			e.safeModifierClick(currentMods)
		case "exit":
			os.Exit(0)
		default:
			// Check Maps
			if symbol, ok := symbolMap[verb]; ok {
				e.executor.TypeStr(symbol)
			} else if digit, ok := spanishMap[verb]; ok {
				e.executor.TypeStr(digit)
			} else if key, ok := phoneticMap[verb]; ok {
				e.executor.TapKey(key, currentMods)
			} else if len(verb) == 1 {
				e.executor.TapKey(verb, currentMods)
			}
		}

		if count > 1 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	// 4. Cleanup Modifiers
	if len(cleanupMods) > 0 {
		for _, m := range cleanupMods {
			e.executor.ToggleKey(m, "up")
		}
	}

	// 5. Clear State
	e.state.pendingModifiers = []string{}
}

// --- ACTION HELPERS ---

func (e *SniperEngine) handleMouse(direction string, val int) {
	x, y := e.executor.GetMousePos()
	switch direction {
	case "left":
		e.executor.MoveMouse(x-val, y)
	case "right":
		e.executor.MoveMouse(x+val, y)
	case "up":
		e.executor.MoveMouse(x, y-val)
	case "down":
		e.executor.MoveMouse(x, y+val)
	}
}

func (e *SniperEngine) handleDirectionalKeys(cardinal string, mods []interface{}) {
	key := ""
	switch cardinal {
	case "west":
		key = "left"
	case "east":
		key = "right"
	case "north":
		key = "up"
	case "south":
		key = "down"
	}
	if key != "" {
		e.executor.TapKey(key, mods)
	}
}

func (e *SniperEngine) safeModifierClick(mods []interface{}) {
	for _, m := range mods {
		if s, ok := m.(string); ok {
			e.executor.ToggleKey(s, "down")
		}
	}
	e.executor.Click("left")
	for _, m := range mods {
		if s, ok := m.(string); ok {
			e.executor.ToggleKey(s, "up")
		}
	}
}

func (e *SniperEngine) isModifier(word string) bool {
	switch word {
	case "control", "command", "option", "alt", "shift":
		return true
	}
	return false
}

func (e *SniperEngine) queueModifier(word string) {
	key := word
	if runtime.GOOS == "darwin" {
		switch word {
		case "command":
			key = "cmd"
		case "option":
			key = "lalt"
		case "control":
			key = "lctrl"
		}
	} else {
		if word == "command" {
			key = "control"
		}
	}

	for _, m := range e.state.pendingModifiers {
		if m == key {
			return
		}
	}
	e.state.pendingModifiers = append(e.state.pendingModifiers, key)
	fmt.Printf("(Modifier) Queued: %s\n", key)
}

func (e *SniperEngine) cleanInput(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	return strings.TrimSpace(s)
}

func (e *SniperEngine) getDigitString(word string) (string, bool) {
	// Previously split input, now takes a single word
	if val, ok := digitMap[word]; ok {
		return val, true
	}
	if _, err := strconv.Atoi(word); err == nil {
		return word, true
	}
	return "", false
}

// --- HISTORY MANAGER ---

var HistoryTriggers = []string{
	"acorn", "beacon", "cactus", "dial", "eagle", "falcon", "garden", "harbor", "iceberg", "jungle",
	"lantern", "magnet", "nectar", "oasis", "paddle", "quilt", "radar", "saddle", "tablet", "ultra",
	"valley", "wagon", "yacht", "zebra", "amber", "bronze", "cedar", "denim", "ember", "fabric",
	"gadget", "habit", "icon", "jacket", "kabob", "laser", "marble", "nacho", "orbit", "packet",
	"quartz", "razor", "safari", "tactic", "vacuum", "wafer", "yarn", "zenith", "apron", "barrel",
	"canvas", "dagger", "easel", "fable", "gasket", "haven", "idiom", "joker", "kennel", "ladder",
	"mantle", "napkin", "object", "palace", "quiver", "rabbit", "saint", "talent", "uncle", "vapor",
	"waffle", "yellow", "zone", "arrow", "basket", "candle", "danger", "earth", "fantasy", "galaxy",
	"hammer", "image", "jazz", "kettle", "lemon", "magic", "nature", "ocean", "panda", "queen",
	"radius", "salad", "target", "unit", "velvet", "wallet", "yogurt", "zoom",
}

type HistoryEntry struct {
	Trigger string `json:"trigger"`
	Command string `json:"command"`
}

type HistoryManager struct {
	mu      sync.Mutex
	Entries []HistoryEntry
}

func NewHistoryManager() *HistoryManager {
	h := &HistoryManager{}
	h.Init()
	return h
}

func (h *HistoryManager) Init() {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.ReadFile(HistoryFile)
	if err == nil {
		json.Unmarshal(file, &h.Entries)
	}

	if len(h.Entries) != len(HistoryTriggers) {
		newEntries := make([]HistoryEntry, len(HistoryTriggers))
		for i, word := range HistoryTriggers {
			cmd := ""
			if i < len(h.Entries) && h.Entries[i].Trigger == word {
				cmd = h.Entries[i].Command
			}
			newEntries[i] = HistoryEntry{Trigger: word, Command: cmd}
		}
		h.Entries = newEntries
		h.save()
	}
}

func (h *HistoryManager) Push(newCommand string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.Entries) == 0 {
		return
	}
	if h.Entries[0].Command == newCommand {
		return
	}

	cmds := make([]string, len(h.Entries))
	for i := range h.Entries {
		cmds[i] = h.Entries[i].Command
	}

	copy(cmds[1:], cmds[0:len(cmds)-1])
	cmds[0] = newCommand

	for i := range h.Entries {
		h.Entries[i].Command = cmds[i]
	}

	h.save()
}

func (h *HistoryManager) FindCommand(trigger string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, entry := range h.Entries {
		if entry.Trigger == trigger && entry.Command != "" {
			return entry.Command, true
		}
	}
	return "", false
}

func (h *HistoryManager) save() {
	data, _ := json.MarshalIndent(h.Entries, "", "  ")
	os.WriteFile(HistoryFile, data, 0644)
}

// --- MAIN APPLICATION ---

func main() {
	history := NewHistoryManager()
	executor := &RobotGoExecutor{}
	engine := NewSniperEngine(history, executor)

	errChan := make(chan error, 2)
	go func() {
		fmt.Printf("Client running on port %s\n", ClientPort)
		if err := runClientSide(); err != nil {
			errChan <- err
		}
	}()
	go func() {
		fmt.Printf("Server running on port %s\n", ServerPort)
		if err := runServerSide(engine); err != nil {
			errChan <- err
		}
	}()
	log.Fatal(<-errChan)
}

func runClientSide() error {
	app := vii.NewApp()
	app.Use(vii.MwLogger, vii.MwTimeout(10), vii.MwCORS)
	app.Static("./static")
	app.Favicon()
	if err := app.Templates("./templates", nil); err != nil {
		return err
	}
	app.At("GET /", func(w http.ResponseWriter, r *http.Request) {
		vii.ExecuteTemplate(w, r, "index.html", nil)
	})
	app.At("GET /mouse", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{"Locations": map[string]interface{}{}}
		vii.ExecuteTemplate(w, r, "mouse.html", data)
	})
	app.At("GET /signs", func(w http.ResponseWriter, r *http.Request) {
		vii.ExecuteTemplate(w, r, "signs.html", nil)
	})
	return app.Serve(ClientPort)
}

func runServerSide(engine *SniperEngine) error {
	app := vii.NewApp()
	app.Use(vii.MwLogger, vii.MwTimeout(10), vii.MwCORS)

	app.At("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is healthy"))
	})

	app.At("POST /api/data", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Command string `json:"command"`
		}
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return
		}

		success, reason := engine.ProcessInput(req.Command)

		if success {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"executed"}`))
		} else {
			// Throttle check removed here
			http.Error(w, "Invalid Command: "+reason, http.StatusBadRequest)
		}
	})
	return app.Serve(ServerPort)
}