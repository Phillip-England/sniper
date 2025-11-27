package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Phillip-England/vii"
	"github.com/go-vgo/robotgo"
)

// --- CONFIGURATION ---

const (
	ClientPort = "3000"
	ServerPort = "8000"
)

// Global state to remember the last valid action verb (e.g., "left", "west")
var lastVerb string

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
		handleCommand(req.Command)
		w.Write([]byte("Command processed"))
	})
	return app.Serve(ServerPort)
}

func handleCommand(rawCommand string) {
	// --- PREPROCESSOR ---
	// Convert spelled out numbers (one, two) to digits (1, 2)
	finalCommand := preprocessCommand(rawCommand)

	fmt.Printf("Raw Input: %s | Processed: %s\n", rawCommand, finalCommand)

	cmd := strings.ToLower(finalCommand)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	verb := parts[0]

	// --- LOGIC: HANDLE NUMBER-ONLY INPUT ---
	// If input is just "2", "100", "one", etc.
	if isNumeric(verb) && len(parts) == 1 {
		if lastVerb == "" {
			fmt.Println("Received number but no previous command context.")
			return
		}

		totalRequested := parseFuzzyNumber(verb)

		// Subtract 1 because the action was already performed once in the previous request
		remainingCount := totalRequested - 1

		if remainingCount <= 0 {
			fmt.Printf("Total requested %d. Already executed 1. No remaining actions needed.\n", totalRequested)
			// Even if we don't move, we reset the verb so we don't get stuck in a loop
			lastVerb = ""
			return
		}

		fmt.Printf("Repeating '%s' %d more times (Total: %d | Previous: 1)\n", lastVerb, remainingCount, totalRequested)

		// Execute based on what the LAST verb was
		switch lastVerb {
		// Mouse Logic: Multiply default distance (100) by remaining count
		case "left", "right", "up", "down":
			distance := remainingCount * 100
			// Construct a command string that looks like "left 200"
			executionCmd := fmt.Sprintf("%s %d", lastVerb, distance)
			handleMouse(executionCmd)

		// Keyboard Logic: Just pass the remaining count as key presses
		case "west", "east", "north", "south":
			// Construct a command string that looks like "west 2"
			executionCmd := fmt.Sprintf("%s %d", lastVerb, remainingCount)
			handleKeyboard(executionCmd)

		// Click Logic: Pass the remaining count as click repetitions
		case "click":
			executionCmd := fmt.Sprintf("%s %d", lastVerb, remainingCount)
			handleClick(executionCmd)
		}

		// RESET VERB HERE
		// Once the repetition is handled, we clear the memory.
		lastVerb = ""
		return
	}

	// --- NORMAL EXECUTION ---
	// If it's not a number, we execute normally and save the verb for later
	switch verb {
	case "left", "right", "up", "down":
		lastVerb = verb // Save state
		handleMouse(cmd)
	case "west", "east", "north", "south":
		lastVerb = verb // Save state
		handleKeyboard(cmd)
	case "click":
		lastVerb = verb // Save state
		handleClick(cmd)
	default:
		fmt.Printf("Unknown Command Ignored: %s\n", finalCommand)
	}
}

// --- Input Helpers ---

func handleMouse(command string) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return
	}
	direction := parts[0]
	val := 50 // Default pixel distance
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

	count := 1 // Default key presses
	if len(parts) > 1 {
		if v := parseFuzzyNumber(parts[1]); v > 0 {
			count = v
		}
	}

	// Map cardinal directions to keyboard keys
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
		}
		fmt.Printf("(Keyboard) Pressed %s (mapped from %s) %d times\n", key, cardinalDirection, count)
	}
}

func handleClick(command string) {
	parts := strings.Fields(command)
	count := 1 // Default to 1 click
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

// preprocessCommand iterates through every word in the command and converts
// spelled-out numbers into digit literals (e.g., "twenty" -> "20").
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
		// Clean punctuation for matching (e.g. "five.")
		cleanWord := strings.ToLower(strings.TrimRight(word, ",.!?"))
		if val, ok := wordToNum[cleanWord]; ok {
			// Restore punctuation if it existed
			punctuation := ""
			if len(word) > len(cleanWord) {
				punctuation = word[len(cleanWord):]
			}
			words[i] = val + punctuation
		}
	}
	return strings.Join(words, " ")
}