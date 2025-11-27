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
	ClientPort    = "3000"
	ServerPort    = "8000"
	ThrottleMs    = 200
	MouseDistance = 50
)

// --- GLOBAL STATE ---
var (
	mu              sync.Mutex
	lastSuccessTime time.Time

	// CHANGED: Modifiers are just strings we track, we don't press them immediately
	pendingModifiers []string

	// STATE FOR ACCUMULATION
	lastVerb          string
	numberAccumulator string
	executedCount     int

	// Whitelist: Now uses Phonetic Alphabet instead of raw letters
	validCommands = map[string]bool{
		// Controls
		"left": true, "right": true, "up": true, "down": true,
		"west": true, "east": true, "north": true, "south": true,
		"click": true, "exit": true,
		"control": true, "command": true, "option": true, "alt": true, "shift": true,
		
		// Phonetic Alphabet
		"alpha": true, "bravo": true, "charlie": true, "delta": true,
		"echo": true, "foxtrot": true, "golf": true, "hotel": true,
		"india": true, "juliet": true, "kilo": true, "lima": true,
		"mike": true, "november": true, "oscar": true, "papa": true,
		"quebec": true, "romeo": true, "sierra": true, "tango": true,
		"uniform": true, "victor": true, "whiskey": true, "xray": true,
		"yankee": true, "zulu": true,
	}

	// Maps phonetic words to the key to be pressed
	phoneticMap = map[string]string{
		"alpha": "a", "bravo": "b", "charlie": "c", "delta": "d",
		"echo": "e", "foxtrot": "f", "golf": "g", "hotel": "h",
		"india": "i", "juliet": "j", "kilo": "k", "lima": "l",
		"mike": "m", "november": "n", "oscar": "o", "papa": "p",
		"quebec": "q", "romeo": "r", "sierra": "s", "tango": "t",
		"uniform": "u", "victor": "v", "whiskey": "w", "xray": "x",
		"yankee": "y", "zulu": "z",
	}

	digitMap = map[string]string{
		"zero": "0", "one": "1", "won": "1", "two": "2", "to": "2", "too": "2",
		"three": "3", "four": "4", "for": "4", "five": "5",
		"six": "6", "seven": "7", "eight": "8", "nine": "9", "ten": "10",
	}
)

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

		success, errType := processAndExecute(req.Command)

		if success {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"executed"}`))
		} else {
			if errType == "throttle" {
				http.Error(w, "Throttled", http.StatusTooManyRequests)
			} else {
				http.Error(w, "Invalid Command", http.StatusBadRequest)
			}
		}
	})
	return app.Serve(ServerPort)
}

func processAndExecute(rawInput string) (bool, string) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(lastSuccessTime) < time.Duration(ThrottleMs)*time.Millisecond {
		return false, "throttle"
	}

	cleanedInput := cleanInput(rawInput)
	if cleanedInput == "" {
		return false, "empty"
	}

	digitStr, isNumber := getDigitString(cleanedInput)

	if isNumber {
		if lastVerb == "" {
			return false, "no_context"
		}
		if len(numberAccumulator) >= 2 {
			return true, "limit_reached"
		}

		numberAccumulator += digitStr
		targetTotal, _ := strconv.Atoi(numberAccumulator)
		delta := targetTotal - executedCount

		if delta > 0 {
			executeAction(lastVerb, delta)
			executedCount += delta
		}

	} else {
		words := strings.Fields(cleanedInput)
		command := words[len(words)-1]

		if !validCommands[command] {
			return false, "invalid"
		}

		lastVerb = command
		numberAccumulator = ""
		executedCount = 0

		executeAction(command, 1)
		executedCount = 1
	}

	lastSuccessTime = time.Now()
	return true, ""
}

func executeAction(verb string, count int) {
	fmt.Printf("[Execute] %s x %d | Mods: %v\n", verb, count, pendingModifiers)

	// 1. IF MODIFIER: Just queue it and return.
	if isModifier(verb) {
		queueModifier(verb)
		return
	}

	// 2. CONVERT MODIFIERS TO INTERFACE for robotgo
	modInterface := make([]interface{}, len(pendingModifiers))
	for i, v := range pendingModifiers {
		modInterface[i] = v
	}

	// 3. PERFORM ACTION LOOP
	for i := 0; i < count; i++ {
		switch verb {
		case "left", "right", "up", "down":
			handleMouse(verb, MouseDistance)
		case "west", "east", "north", "south":
			handleKeyboardDirection(verb, modInterface)
		case "click":
			safeModifierClick(modInterface)
		case "exit":
			os.Exit(0)
		default:
			// Check Phonetic Map
			if key, ok := phoneticMap[verb]; ok {
				robotgo.KeyTap(key, modInterface...)
			} else if len(verb) == 1 {
				// Fallback for single letters if they slip through
				robotgo.KeyTap(verb, modInterface...)
			}
		}

		if count > 1 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	// 4. CLEANUP
	pendingModifiers = []string{}
}

// --- MODIFIER LOGIC (Stateless/Atomic) ---

func isModifier(word string) bool {
	switch word {
	case "control", "command", "option", "alt", "shift":
		return true
	}
	return false
}

func queueModifier(word string) {
	key := word
	if runtime.GOOS == "darwin" {
		switch word {
		case "command":
			key = "cmd"
		case "option":
			key = "lalt" // RobotGo prefers lalt/ralt on Mac
		case "control":
			key = "lctrl"
		}
	} else {
		if word == "command" {
			key = "control"
		}
	}

	// Deduplication
	for _, m := range pendingModifiers {
		if m == key {
			return
		}
	}

	pendingModifiers = append(pendingModifiers, key)
	fmt.Printf("(Modifier) Queued: %s\n", key)
}

func safeModifierClick(mods []interface{}) {
	for _, m := range mods {
		if s, ok := m.(string); ok {
			robotgo.KeyToggle(s, "down")
		}
	}
	robotgo.Click("left")
	for _, m := range mods {
		if s, ok := m.(string); ok {
			robotgo.KeyToggle(s, "up")
		}
	}
}

// --- PARSING HELPERS ---

func cleanInput(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	return strings.TrimSpace(s)
}

func getDigitString(input string) (string, bool) {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "", false
	}
	lastWord := words[len(words)-1]

	if val, ok := digitMap[lastWord]; ok {
		return val, true
	}
	if _, err := strconv.Atoi(lastWord); err == nil {
		return lastWord, true
	}
	return "", false
}

// --- ACTION HELPERS ---

func handleMouse(direction string, val int) {
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
}

func handleKeyboardDirection(cardinal string, mods []interface{}) {
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
		robotgo.KeyTap(key, mods...)
	}
}