package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Phillip-England/vii"
	"github.com/go-vgo/robotgo"
)

// --- CONFIGURATION ---

const (
	ClientPort = "3000"
	ServerPort = "8000"
	// ThrottleMs is the gatekeeper window. If the same command comes in
	// within this window, it is ignored to prevent echo/rapid-fire.
	ThrottleMs = 400
)

// Global state to remember the last valid action verb (e.g., "left", "west")
var lastVerb string

// Global state to store keys currently held down (e.g. "control", "alt")
var activeModifiers []string

// Global state for throttling/gatekeeping
var lastProcessedCmd string
var lastProcessedTime time.Time

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

		// Pass to gatekeeper/handler
		executed := handleCommand(req.Command)

		// Return status so client knows whether to play the "Click" sound
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"executed": executed,
		})
	})
	return app.Serve(ServerPort)
}

// handleCommand returns true if the command was executed, false if throttled/ignored
func handleCommand(rawCommand string) bool {
	// --- PREPROCESSOR ---
	finalCommand := preprocessCommand(rawCommand)
	cmd := strings.ToLower(finalCommand)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}
	verb := parts[0]

	// --- GATEKEEPER (THROTTLE LOGIC) ---
	// If the EXACT same command (verb) is received too quickly, ignore it.
	// We check if it is NOT numeric because we want numbers to pass through rapidly if needed.
	if !isNumeric(verb) {
		isRepeat := (verb == lastProcessedCmd)
		timeDelta := time.Since(lastProcessedTime)

		if isRepeat && timeDelta < time.Duration(ThrottleMs)*time.Millisecond {
			fmt.Printf("[Throttle] Ignored duplicate: %s (Delta: %v)\n", verb, timeDelta)
			return false
		}

		// Update Throttle State
		lastProcessedCmd = verb
		lastProcessedTime = time.Now()
	}

	fmt.Printf("Raw Input: %s | Processed: %s\n", rawCommand, finalCommand)

	// --- 1. MODIFIER KEY LOGIC ---
	if isModifier(verb) {
		handleModifier(verb)
		return true
	}

	// --- LOGIC: HANDLE NUMBER-ONLY INPUT ---
	if isNumeric(verb) && len(parts) == 1 {
		if lastVerb == "" {
			fmt.Println("Received number but no previous command context.")
			return false
		}

		totalRequested := parseFuzzyNumber(verb)
		
		// --- THE FIX: Decrement by 1 ---
		// User said "Right" (1 move executed). 
		// User says "Seven" (wants 7 total moves).
		// We execute 6 more.
		remainingCount := totalRequested - 1 

		if remainingCount <= 0 {
			// If user says "One", total is 1. We already did 1. Do nothing.
			lastVerb = ""
			releaseAllModifiers() 
			return true
		}

		fmt.Printf("Repeating '%s' %d more times (Total Requested: %d)\n", lastVerb, remainingCount, totalRequested)

		switch lastVerb {
		case "left", "right", "up", "down":
			// Note: We use 100 as the multiplier here. 
			// If your default move is 50, you might want to change this to 50.
			distance := remainingCount * 100
			handleMouse(fmt.Sprintf("%s %d", lastVerb, distance))

		case "west", "east", "north", "south":
			handleKeyboard(fmt.Sprintf("%s %d", lastVerb, remainingCount))

		case "click":
			handleClick(fmt.Sprintf("%s %d", lastVerb, remainingCount))

		default:
			if len(lastVerb) == 1 {
				for i := 0; i < remainingCount; i++ {
					robotgo.KeyTap(lastVerb)
					time.Sleep(5 * time.Millisecond)
				}
				fmt.Printf("(Keyboard) Typed '%s' %d times\n", lastVerb, remainingCount)
			}
		}

		// Reset context so we don't chain numbers endlessly
		lastVerb = ""
		releaseAllModifiers()
		return true
	}

	// --- NORMAL EXECUTION ---
	switch verb {
	case "left", "right", "up", "down":
		lastVerb = verb
		// Updated default to 100 to match the numeric multiplier above
		// This makes "Right" + "Seven" consistent mathematically.
		handleMouse(cmd) 
		releaseAllModifiers()
	case "west", "east", "north", "south":
		lastVerb = verb
		handleKeyboard(cmd)
		releaseAllModifiers()
	case "click":
		lastVerb = verb
		handleClick(cmd)
		releaseAllModifiers()

	// --- ALPHABET HANDLING ---
	case "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", 
	     "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z":
		lastVerb = verb
		robotgo.KeyTap(verb)
		releaseAllModifiers()

	default:
		fmt.Printf("Unknown Command Ignored: %s\n", finalCommand)
		return false
	}

	return true
}

// --- Modifier Helpers ---

func isModifier(word string) bool {
	switch word {
	case "control", "command", "option", "alt", "shift":
		return true
	}
	return false
}

func handleModifier(word string) {
	key := word
	if key == "option" {
		key = "alt"
	} else if key == "command" {
		key = "cmd"
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
	for i := len(activeModifiers) - 1; i >= 0; i-- {
		key := activeModifiers[i]
		robotgo.KeyToggle(key, "up")
		time.Sleep(20 * time.Millisecond)
		fmt.Printf("(Modifier) Released: %s\n", key)
	}
	activeModifiers = []string{}
}

// --- Input Helpers ---

func handleMouse(command string) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return
	}
	direction := parts[0]
	// Changed default to 100 to match the numeric multiplier logic
	val := 100 
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

func handleKeyboard(command string) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return
	}
	cardinalDirection := parts[0]
	count := 1 
	if len(parts) > 1 {
		if v := parseFuzzyNumber(parts[1]); v > 0 {
			count = v
		}
	}

	key := ""
	switch cardinalDirection {
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
		for i := 0; i < count; i++ {
			robotgo.KeyTap(key)
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Printf("(Keyboard) Pressed %s (mapped from %s) %d times\n", key, cardinalDirection, count)
	}
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
	}
	fmt.Printf("(Mouse) Clicked left %d times\n", count)
}

// isNumeric checks if a string is a valid integer representation
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
			}
			words[i] = val + punctuation
		}
	}
	return strings.Join(words, " ")
}