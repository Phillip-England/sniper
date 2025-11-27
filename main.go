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

// --- CONFIGURATION ---
const (
	ClientPort                = "3000"
	ServerPort                = "8000"
	ThrottleMs                = 400
	MOUSE_MOVE_DISTANCE       = 100
	MOUSE_REPETITION_DISTANCE = 50
	HistoryFile               = "history.json"
	SequenceTimeout           = 3000
)

// Global state
var (
	mu                sync.Mutex
	lastVerb          string
	activeModifiers   []string
	lastProcessedCmd  string
	lastProcessedTime time.Time
	// Sequencing state
	lastNumericInput string
	lastTotalCount   int
)

// HistoryItem represents a single command record with an ID
type HistoryItem struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

// 100 distinct words not in the standard command set
var idWords = []string{
	"acorn", "bacon", "cactus", "dairy", "eagle", "fable", "giant", "haven", "igloo", "joker",
	"koala", "lemon", "melon", "neon", "orbit", "pizza", "radar", "solar", "tiger", "ultra",
	"vivid", "wagon", "xenon", "yogurt", "zebra", "amber", "bravo", "cedar", "delta", "epoch",
	"falcon", "gamma", "hazel", "inlet", "jump", "karma", "lunar", "mango", "noble", "ocean",
	"piano", "quartz", "radio", "sute", "tempo", "urban", "virus", "whale", "xerox", "youth",
	"zesty", "angel", "bingo", "candy", "diner", "elfin", "flame", "ghost", "happy", "irony",
	"jelly", "kitty", "laser", "magic", "ninja", "olive", "panda", "quest", "robin", "sugar",
	"token", "unity", "viper", "water", "xylophone", "yacht", "zinc", "atlas", "blaze", "comet",
	"dream", "ember", "frost", "glory", "hero", "image", "jewel", "knife", "logic", "metal",
	"nurse", "onion", "pilot", "quiet", "river", "snake", "table", "uncle", "video", "wolf",
}

func main() {
	errChan := make(chan error, 2)
	go func() {
		fmt.Printf("Client running on port %s\n", ClientPort)
		if err := runClientSide(); err != nil {
			errChan <- err
		}
	}()
	go func() {
		fmt.Printf("Server running on port %s\n", ServerPort)
		if err := runServerSide(); err != nil {
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
	return app.Serve(ClientPort)
}

func runServerSide() error {
	app := vii.NewApp()
	app.Use(vii.MwLogger, vii.MwTimeout(10), vii.MwCORS)
	app.At("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is healthy"))
	})
	app.At("POST /api/data", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		executed := handleCommand(req.Command)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"executed": executed,
		})
	})
	return app.Serve(ServerPort)
}

func handleCommand(rawCommand string) bool {
	mu.Lock()
	defer mu.Unlock()
	finalCommand := preprocessCommand(rawCommand)
	cmd := strings.ToLower(finalCommand)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}
	verb := parts[0]

	// --- HISTORY COMMAND ---
	if verb == "history" {
		printHistory()
		return true
	}

	// --- GATEKEEPER ---
	if !isNumeric(verb) {
		isRepeat := (verb == lastProcessedCmd)
		timeDelta := time.Since(lastProcessedTime)
		if isRepeat && timeDelta < time.Duration(ThrottleMs)*time.Millisecond {
			fmt.Printf("[Throttle] Ignored duplicate: %s\n", verb)
			return false
		}
		lastProcessedCmd = verb
		lastProcessedTime = time.Now()
	}
	fmt.Printf("Raw: %s | Processed: %s\n", rawCommand, finalCommand)

	// --- 1. MODIFIER LOGIC ---
	if isModifier(verb) {
		handleModifier(verb)
		updateHistory(finalCommand)
		return true
	}

	// --- 2. NUMBER LOGIC (Refinement & Concatenation) ---
	if isNumeric(verb) && len(parts) == 1 {
		if lastVerb == "" {
			return false
		}

		now := time.Now()
		timeSinceLast := now.Sub(lastProcessedTime)

		// Determine if we are continuing a sequence
		isRecent := timeSinceLast < time.Duration(SequenceTimeout)*time.Millisecond
		isConcatenation := isRecent && lastNumericInput != ""

		var targetTotal int
		var newNumStr string

		if isConcatenation {
			// Case A: Concatenate (e.g., "2" -> "2") => "22"
			newNumStr = lastNumericInput + verb
			targetTotal, _ = strconv.Atoi(newNumStr)
		} else if isRecent {
			// Case B: Refinement (e.g., "Left" -> "2") => Total 2
			newNumStr = verb
			targetTotal = parseFuzzyNumber(verb)
		} else {
			// Case C: Fresh Number Command (Timeout expired)
			newNumStr = verb
			targetTotal = parseFuzzyNumber(verb)
			lastTotalCount = 0
		}

		delta := targetTotal - lastTotalCount

		if delta > 0 {
			fmt.Printf("[Sequence] Verb='%s' PrevTotal=%d NewTotal=%d Delta=%d (Concat=%v)\n",
				lastVerb, lastTotalCount, targetTotal, delta, isConcatenation)

			executeAction(lastVerb, delta, true)

			lastTotalCount = targetTotal
			lastNumericInput = newNumStr
			lastProcessedTime = now // Update time to keep the sequence alive
			updateHistory(finalCommand)
		} else {
			fmt.Printf("[Sequence] Ignored non-positive delta: %d\n", delta)
		}

		return true
	}

	// --- 3. STANDARD EXECUTION ---
	success := executeAction(cmd, 1, false)
	if success {
		// Update context
		if isAlphabet(verb) || isDirection(verb) || verb == "click" {
			lastVerb = verb
			// Reset sequence trackers for new verb
			lastTotalCount = 1
			lastNumericInput = ""
		}
		// CRITICAL: Delay before releasing modifiers
		time.Sleep(50 * time.Millisecond)
		releaseAllModifiers()
		updateHistory(finalCommand)
	}
	return success
}

// =========================================================================
// --- CORE ACTION DISPATCH AND EXECUTION (INLINED LOGIC) ---
// =========================================================================

func executeAction(command string, count int, isNumericRepetition bool) bool {
	parts := strings.Fields(command)
	verb := parts[0]
	args := parts[1:]

	// Prepare modifiers for KeyTap
	mods := make([]interface{}, len(activeModifiers))
	for i, m := range activeModifiers {
		mods[i] = m
	}

	switch verb {
	// --- MOUSE MOVEMENT ACTIONS (LEFT, RIGHT, UP, DOWN) ---
	case "left":
		fallthrough
	case "right":
		fallthrough
	case "up":
		fallthrough
	case "down":
		// Mouse movements are not usually affected by key modifiers.
		if isNumericRepetition {
			// **UPDATED** to use MOUSE_REPETITION_DISTANCE
			distance := count * MOUSE_REPETITION_DISTANCE
			handleMouse(fmt.Sprintf("%s %d", verb, distance))
		} else {
			handleMouse(strings.Join(append([]string{verb}, args...), " "))
		}
		return true

	// --- KEYBOARD DIRECTION ACTIONS (WEST, EAST, NORTH, SOUTH) ---
	case "west":
		fallthrough
	case "east":
		fallthrough
	case "north":
		fallthrough
	case "south":
		key := ""
		switch verb {
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
			// Apply repetition and modifiers
			for i := 0; i < count; i++ {
				robotgo.KeyTap(key, mods...)
				time.Sleep(5 * time.Millisecond)
			}
			fmt.Printf("(Keyboard) Pressed %s %d times with mods: %v\n", key, count, activeModifiers)
			return true
		}
		return false

	// --- CLICK ACTION ---
	case "click":
		cmdToSend := verb
		if len(args) > 0 {
			cmdToSend = strings.Join(append([]string{verb}, args...), " ")
		}
		if isNumericRepetition {
			handleClick(fmt.Sprintf("%s %d", verb, count))
		} else {
			handleClick(cmdToSend)
		}
		return true

	// --- DEFAULT / ALPHABET ACTION ---
	default:
		if isAlphabet(verb) {
			// Apply repetition and modifiers
			for i := 0; i < count; i++ {
				robotgo.KeyTap(verb, mods...)
				time.Sleep(5 * time.Millisecond)
			}
			fmt.Printf("(Keyboard) Pressed %s %d times with mods: %v\n", verb, count, activeModifiers)
			return true
		}
		return false
	}
}

// =========================================================================
// --- HELPER FUNCTIONS ---
// =========================================================================

// --- History Helpers ---

func loadHistory() []HistoryItem {
	file, err := os.ReadFile(HistoryFile)
	if err != nil {
		return []HistoryItem{}
	}
	var history []HistoryItem
	json.Unmarshal(file, &history)
	return history
}

func saveHistory(history []HistoryItem) {
	data, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(HistoryFile, data, 0644)
}

func updateHistory(cmd string) {
	history := loadHistory()

	// Determine next ID index based on the last item in history
	nextIndex := 0
	if len(history) > 0 {
		lastID := history[len(history)-1].ID
		// Find index of lastID in idWords
		for i, word := range idWords {
			if word == lastID {
				nextIndex = i + 1
				break
			}
		}
	}

	// Loop back to 0 if we reached the end of the word list
	if nextIndex >= len(idWords) {
		nextIndex = 0
	}

	id := idWords[nextIndex]

	newItem := HistoryItem{
		ID:      id,
		Command: cmd,
	}

	history = append(history, newItem)

	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	saveHistory(history)
}

func printHistory() {
	history := loadHistory()
	count := 5
	if len(history) < count {
		count = len(history)
	}
	fmt.Println("--- Recent History ---")
	// Show newest first
	for i := 0; i < count; i++ {
		idx := len(history) - 1 - i
		item := history[idx]
		fmt.Printf("%d. [%s] %s\n", i+1, item.ID, item.Command)
	}
}

// --- Key Mapping Helpers ---

// getModifierKey returns the robotgo key identifier for a given spoken verb,
// ensuring cross-platform compatibility.
func getModifierKey(verb string) string {
	switch verb {
	case "control", "ctrl":
		if runtime.GOOS == "darwin" {
			return "lctrl"
		}
		return "control"
	case "command", "cmd":
		if runtime.GOOS == "darwin" {
			return "cmd"
		}
		return "control"
	case "shift":
		return "shift"
	case "alt":
		return "alt"
	case "option":
		return "alt"
	case "function":
		return ""
	default:
		return ""
	}
}

// --- Modifier Handlers ---

func isModifier(word string) bool {
	switch word {
	case "control", "ctrl", "command", "cmd", "option", "alt", "shift":
		return true
	}
	return false
}

func handleModifier(spokenWord string) {
	key := getModifierKey(spokenWord)

	if key == "" {
		return
	}

	for _, m := range activeModifiers {
		if m == key {
			return
		}
	}

	robotgo.KeyToggle(key, "down")
	time.Sleep(20 * time.Millisecond)
	activeModifiers = append(activeModifiers, key)
	fmt.Printf("(Modifier) Holding down: %s\n", key)
}

func releaseAllModifiers() {
	if len(activeModifiers) == 0 {
		return
	}
	// LIFO Release
	for i := len(activeModifiers) - 1; i >= 0; i-- {
		key := activeModifiers[i]
		robotgo.KeyToggle(key, "up")
		time.Sleep(20 * time.Millisecond)
		fmt.Printf("(Modifier) Released: %s\n", key)
	}
	activeModifiers = []string{}
}

// --- General Helpers ---

func isAlphabet(s string) bool {
	if len(s) != 1 {
		return false
	}
	char := s[0]
	return char >= 'a' && char <= 'z'
}

func isDirection(s string) bool {
	switch s {
	case "left", "right", "up", "down", "west", "east", "north", "south":
		return true
	}
	return false
}

func handleMouse(command string) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return
	}
	direction := parts[0]
	// **UPDATED** to use MOUSE_MOVE_DISTANCE as the default value
	val := MOUSE_MOVE_DISTANCE
	if len(parts) > 1 {
		if v := parseFuzzyNumber(parts[1]); v > 0 {
			val = v
		}
	}
	x, y := robotgo.Location()
	switch direction {
	case "left":
		robotgo.Move(x-val, y)
	case "right":
		robotgo.Move(x+val, y)
	case "up":
		robotgo.Move(x, y-val)
	case "down":
		robotgo.Move(x, y+val)
	}
	fmt.Printf("(Mouse) Moved %s by %d px\n", direction, val)
}

func handleClick(command string) {
	parts := strings.Fields(command)
	count := 1
	if len(parts) > 1 {
		if v := parseFuzzyNumber(parts[1]); v > 0 {
			count = v
		}
	}
	for i := 0; i < count; i++ {
		robotgo.Click("left")
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("(Mouse) Clicked left %d times\n", count)
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func parseFuzzyNumber(s string) int {
	wordToNum := map[string]int{
		"one": 1, "won": 1, "two": 2, "to": 2, "too": 2,
		"three": 3, "four": 4, "for": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	if val, ok := wordToNum[strings.ToLower(s)]; ok {
		return val
	}
	clean := strings.ReplaceAll(s, "$", "")
	clean = strings.ReplaceAll(clean, ",", "")
	clean = strings.ReplaceAll(clean, "ms", "")
	clean = strings.ReplaceAll(clean, "s", "")
	val, _ := strconv.Atoi(clean)
	return val
}

func preprocessCommand(input string) string {
	wordToNum := map[string]string{
		"zero": "0", "one": "1", "won": "1", "two": "2", "to": "2", "too": "2",
		"three": "3", "four": "4", "for": "4", "five": "5",
		"six": "6", "seven": "7", "eight": "8", "nine": "9", "ten": "10",
		"eleven": "11", "twelve": "12", "thirteen": "13", "fourteen": "14",
		"fifteen": "15", "sixteen": "16", "seventeen": "17", "eighteen": "18", "nineteen": "19",
		"twenty": "20", "thirty": "30", "forty": "40", "fifty": "50",
		"sixty": "60", "seventy": "70", "eighty": "80", "ninety": "90",
		"hundred": "100",
	}
	words := strings.Fields(input)
	for i, word := range words {
		cleanWord := strings.ToLower(strings.TrimRight(word, ",.!?"))
		if val, ok := wordToNum[cleanWord]; ok {
			punctuation := ""
			if len(word) > len(cleanWord) {
				punctuation = word[len(cleanWord):]
				words[i] = val + punctuation
			}
		}
	}
	return strings.Join(words, " ")
}