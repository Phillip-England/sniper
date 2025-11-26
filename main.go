package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Phillip-England/vii"
	"github.com/go-vgo/robotgo"
)

const (
	ClientPort   = "3000"
	ServerPort   = "8000"
	OllamaURL    = "http://localhost:11434/api/generate"
	OllamaModel  = "llama3.2"
	ScriptsFile  = "scripts.json"
	HistoryFile  = "history.json"
	AlphabetFile = "alphabet.json"
	PhrasesFile  = "phrases.json"
)

var lastBatchCommand string

type MousePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type KeyAction struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers,omitempty"`
}

type AlphabetConfig struct {
	Default []KeyAction `json:"default"`
	Darwin  []KeyAction `json:"darwin,omitempty"`
	Linux   []KeyAction `json:"linux,omitempty"`
	Windows []KeyAction `json:"windows,omitempty"`
}

type HistoryItem struct {
	ID        int    `json:"id"`
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
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

	// History - Return Raw JSON
	app.At("GET /history", func(w http.ResponseWriter, r *http.Request) {
		history := loadHistory()
		for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
			history[i], history[j] = history[j], history[i]
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(history)
	})

	// Scripts - Return Raw JSON
	app.At("GET /scripts", func(w http.ResponseWriter, r *http.Request) {
		scripts := loadScripts()
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(scripts)
	})

	// Alphabet - Return Raw JSON
	app.At("GET /alphabet", func(w http.ResponseWriter, r *http.Request) {
		alphabet := loadAlphabet()
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(alphabet)
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
	fmt.Printf("Raw Input: %s\n", rawCommand)

	// Log to History first
	addToHistory(rawCommand)

	// --- BATCHING LOGIC ---
	// Always check for the " then " delimiter to handle command chaining.
	// We use lower case checking to be case-insensitive for the delimiter.
	lowerInput := strings.ToLower(rawCommand)
	if strings.Contains(lowerInput, " then ") {
		// Update lastBatchCommand so 'script save' works with implicit batches
		lastBatchCommand = rawCommand

		parts := strings.Split(lowerInput, " then ")
		for _, part := range parts {
			cleanCmd := strings.TrimSpace(part)
			// Ignore empty commands or commands that are just whitespace
			if cleanCmd != "" {
				handleCommand(cleanCmd)
				// Reset modifiers and wait between batch steps
				resetModifiers()
				time.Sleep(300 * time.Millisecond)
			}
		}
		// Return immediately so we don't process the full string as a single command
		return
	}
	// ----------------------

	cmd := strings.ToLower(rawCommand)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}
	verb := parts[0]
	args := parts[1:]

	switch verb {
	case "history":
		if len(args) < 1 {
			return
		}
		if args[0] == "undo" {
			hist := loadHistory()
			if len(hist) > 0 {
				removeCount := 1
				if len(hist) > 1 {
					removeCount = 2
				}
				hist = hist[:len(hist)-removeCount]
				updatedData, err := json.MarshalIndent(hist, "", "  ")
				if err == nil {
					os.WriteFile(HistoryFile, updatedData, 0644)
					fmt.Printf("HISTORY LOG: Removed last %d items.\n", removeCount)
				}
			}
			return
		}

		if args[0] == "capture" {
			if len(args) < 3 {
				fmt.Println("Usage: history capture <count> <name>")
				return
			}
			count := parseFuzzyNumber(args[1])
			if count <= 0 {
				fmt.Println("Count must be greater than 0")
				return
			}
			name := strings.Join(args[2:], " ")
			hist := loadHistory()
			availableLen := len(hist) - 1
			if availableLen < count {
				fmt.Printf("Not enough history. You requested %d, but only have %d commands.\n", count, availableLen)
				return
			}
			endIndex := availableLen
			startIndex := availableLen - count
			targetItems := hist[startIndex:endIndex]
			var collectedCommands []string
			for _, item := range targetItems {
				collectedCommands = append(collectedCommands, item.Command)
			}
			if len(collectedCommands) > 0 {
				batchCmd := strings.Join(collectedCommands, " then ")
				saveScript(name, batchCmd)
				fmt.Printf("Captured last %d commands into script '%s'.\n", count, name)
			}
			return
		}

		if args[0] == "start" {
			if len(args) < 5 {
				fmt.Println("Usage: history start <start_id> and <end_id> <name>")
				return
			}
			startID := parseFuzzyNumber(args[1])
			endID := parseFuzzyNumber(args[3])
			name := strings.Join(args[4:], " ")
			if startID == 0 || endID == 0 {
				fmt.Println("Invalid history IDs provided.")
				return
			}
			if startID > endID {
				fmt.Println("Start ID cannot be greater than End ID.")
				return
			}
			hist := loadHistory()
			var collectedCommands []string
			for i := startID; i <= endID; i++ {
				found := false
				for _, item := range hist {
					if item.ID == i {
						if !strings.HasPrefix(item.Command, "history") {
							collectedCommands = append(collectedCommands, item.Command)
						}
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("Warning: History ID %d not found, skipping.\n", i)
				}
			}
			if len(collectedCommands) > 0 {
				batchCmd := strings.Join(collectedCommands, " then ")
				saveScript(name, batchCmd)
				fmt.Printf("Saved batch script '%s' with %d commands (IDs %d-%d).\n", name, len(collectedCommands), startID, endID)
			} else {
				fmt.Println("No valid commands found in that range.")
			}
			return
		}

		if args[0] == "save" {
			if len(args) < 3 {
				fmt.Println("Usage: history save <id> <name>")
				return
			}
			id, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("Invalid ID")
				return
			}
			name := strings.Join(args[2:], " ")
			hist := loadHistory()
			foundCmd := ""
			for _, item := range hist {
				if item.ID == id {
					foundCmd = item.Command
					break
				}
			}
			if foundCmd != "" {
				saveScript(name, foundCmd)
				fmt.Printf("Saved history #%d as script '%s'\n", id, name)
			} else {
				fmt.Printf("History ID #%d not found.\n", id)
			}
			return
		}

		id, err := strconv.Atoi(args[0])
		if err == nil {
			hist := loadHistory()
			for _, item := range hist {
				if item.ID == id {
					fmt.Printf("Invoking history #%d: %s\n", id, item.Command)
					handleCommand(item.Command)
					return
				}
			}
			fmt.Printf("History ID #%d not found.\n", id)
		}

	case "wait", "weight":
		if len(args) < 1 {
			return
		}
		ms := parseFuzzyNumber(args[0])
		if ms > 0 {
			fmt.Printf("Sleeping for %dms\n", ms)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

	case "repeat":
		if len(args) < 2 {
			return
		}
		count := parseFuzzyNumber(args[0])
		if count < 1 {
			count = 1
		}
		subCommand := strings.Join(args[1:], " ")
		for i := 0; i < count; i++ {
			handleCommand(subCommand)
			time.Sleep(50 * time.Millisecond)
		}

	case "batch":
		fullBatch := strings.Join(args, " ")
		lastBatchCommand = fullBatch
		subCommands := strings.Split(fullBatch, " then ")
		for _, subCmd := range subCommands {
			cleanCmd := strings.TrimSpace(subCmd)
			if cleanCmd != "" {
				handleCommand(cleanCmd)
				resetModifiers()
				time.Sleep(300 * time.Millisecond)
			}
		}

	case "script":
		if len(args) == 0 {
			return
		}
		if args[0] == "delete" {
			if len(args) < 2 {
				fmt.Println("Usage: script delete <name>")
				return
			}
			name := strings.Join(args[1:], " ")
			deleteScript(name)
			fmt.Printf("Deleted script '%s'\n", name)
			return
		}
		if lastBatchCommand == "" {
			fmt.Println("No previous batch command to save.")
			return
		}
		name := strings.Join(args, " ")
		saveScript(name, lastBatchCommand)
		fmt.Printf("Saved script '%s'\n", name)

	case "phrase":
		if len(args) == 0 {
			return
		}
		if args[0] == "delete" {
			if len(args) < 2 {
				fmt.Println("Usage: phrase delete <name>")
				return
			}
			name := strings.Join(args[1:], " ")
			deletePhrase(name)
			fmt.Printf("Deleted phrase '%s'\n", name)
			return
		}

	case "play":
		name := strings.Join(args, " ")
		scripts := loadScripts()
		if storedCmd, ok := scripts[name]; ok {
			fmt.Printf("Playing script '%s'...\n", name)
			handleCommand("batch " + storedCmd)
		} else {
			fmt.Printf("Script '%s' not found.\n", name)
		}

	case "terminal":
		phrase := strings.Join(args, " ")
		go handleTerminalPhrase(phrase)

	case "camel":
		pasteText(toCamelCase(strings.Join(args, " ")))
	case "pascal":
		pasteText(toPascalCase(strings.Join(args, " ")))
	case "snake":
		pasteText(toSnakeCase(strings.Join(args, " ")))
	case "lower":
		pasteText(strings.ToLower(strings.Join(args, " ")))
	case "upper":
		pasteText(strings.ToUpper(strings.Join(args, " ")))

	case "teleport":
		teleportMouse(strings.Join(args, " "))
	case "attack":
		attackMouse(strings.Join(args, " "))
	case "remember":
		rememberMousePosition(strings.Join(args, " "))
	case "forget":
		forgetMousePosition(strings.Join(args, " "))

	case "mouse":
		handleMouse(strings.Join(args, " "))

	case "north", "south", "east", "west":
		handleArrowKeys(verb, args)

	case "left", "right", "write", "up", "down":
		handleMouse(cmd)

	case "scroll":
		handleScroll(cmd)

	case "click":
		robotgo.Click("left", false)
	case "double", "dclick":
		robotgo.Click("left", false)
		time.Sleep(time.Millisecond * 100)
		robotgo.Click("left", false)
	case "clack", "rclick":
		robotgo.Click("right", false)
	case "triple", "tclick":
		robotgo.Click("left", false)
		time.Sleep(time.Millisecond * 100)
		robotgo.Click("left", false)
		time.Sleep(time.Millisecond * 100)
		robotgo.Click("left", false)

	case "say":
		rawText := strings.Join(args, " ")
		if len(rawText) > 0 {
			r := []rune(rawText)
			r[0] = unicode.ToUpper(r[0])
			pasteText(string(r) + ". ")
		}
	case "question":
		rawText := strings.Join(args, " ")
		if len(rawText) > 0 {
			r := []rune(rawText)
			r[0] = unicode.ToUpper(r[0])
			pasteText(string(r) + "? ")
		}
	case "exclaim":
		rawText := strings.Join(args, " ")
		if len(rawText) > 0 {
			r := []rune(rawText)
			r[0] = unicode.ToUpper(r[0])
			pasteText(string(r) + "! ")
		}

	case "log":
		fmt.Println("SYSTEM LOG:", strings.Join(args, " "))

	// --- UNIFIED MODIFIER BRANCH ---
	case "control", "ctrl", "command", "cmd", "shift", "alt", "option", "function":
		// Case 1: Single Key Press (e.g. "command")
		if len(parts) == 1 {
			if key := getModifierKey(verb); key != "" {
				fmt.Printf("Single Key Tap: %s\n", key)
				robotgo.KeyTap(key)
			}
			return
		}

		// Case 2: Modifier Combinations (e.g. "command s")
		targetToken := parts[len(parts)-1]
		modifierTokens := parts[:len(parts)-1]

		var activeModifiers []interface{}

		// Accumulate all modifiers in the command string using unified helper
		for _, token := range modifierTokens {
			if key := getModifierKey(token); key != "" {
				activeModifiers = append(activeModifiers, key)
			}
		}

		// Resolve Target Key (Check Alphabet -> Check Number -> Raw)
		alphabet := loadAlphabet()
		var keyToPress string

		if config, ok := alphabet[targetToken]; ok {
			var actions []KeyAction
			switch runtime.GOOS {
			case "darwin":
				actions = config.Darwin
			case "linux":
				actions = config.Linux
			case "windows":
				actions = config.Windows
			}
			if len(actions) == 0 {
				actions = config.Default
			}

			if len(actions) > 0 {
				keyToPress = actions[0].Key
				// Append inherent modifiers from alphabet config (e.g. colon = shift + ;)
				if len(actions[0].Modifiers) > 0 {
					for _, m := range actions[0].Modifiers {
						activeModifiers = append(activeModifiers, m)
					}
				}
			}
		} else {
			// Fuzzy Number Check
			if num := parseFuzzyNumber(targetToken); num > 0 {
				keyToPress = strconv.Itoa(num)
			} else if num == 0 && (targetToken == "zero" || targetToken == "rim") {
				keyToPress = "0"
			} else {
				// Raw fallback
				keyToPress = targetToken
			}
		}

		fmt.Printf("Modifier Action -> Key: %s | Mods: %v\n", keyToPress, activeModifiers)
		robotgo.KeyTap(keyToPress, activeModifiers...)
		resetModifiers()
		time.Sleep(20 * time.Millisecond)

	// --- DIRECT ALPHABET/KEY MAPPING ---
	case "alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa", "quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey", "xray", "yankee", "zulu",
		"square", "cube", "curve", "loop", "arch", "rim", "less", "great", "plus", "equal", "hyphen", "score", "star", "cent", "hat", "bang", "at", "hash", "cash", "amp", "pipe", "window", "slash", "colon", "semicolon", "quote", "tick", "comma", "dot", "query", "tilde", "grave",
		"next", "gap", "enter", "return", "tab", "escape", "back", "scratch":

		alphabet := loadAlphabet()
		if config, ok := alphabet[verb]; ok {
			var actions []KeyAction
			switch runtime.GOOS {
			case "darwin":
				actions = config.Darwin
			case "linux":
				actions = config.Linux
			case "windows":
				actions = config.Windows
			}
			if len(actions) == 0 {
				actions = config.Default
			}

			if len(actions) > 0 {
				// Loop through all actions (though usually just one for alphabet keys)
				for _, action := range actions {
					modifiers := make([]interface{}, len(action.Modifiers))
					for i, v := range action.Modifiers {
						modifiers[i] = v
					}
					robotgo.KeyTap(action.Key, modifiers...)
					if len(modifiers) > 0 {
						resetModifiers()
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}

	case "spell":
		// --- SPELL BRANCH (Iterates over args) ---
		alphabet := loadAlphabet()
		capitalizeNext := false

		// We iterate over 'args' so we don't type the word "spell"
		for _, token := range args {
			// Check if we need to set the capitalize flag for the next word
			if strings.ToLower(token) == "capital" {
				capitalizeNext = true
				continue
			}

			lowerToken := strings.ToLower(token)
			if config, ok := alphabet[lowerToken]; ok {
				var actions []KeyAction
				switch runtime.GOOS {
				case "darwin":
					actions = config.Darwin
				case "linux":
					actions = config.Linux
				case "windows":
					actions = config.Windows
				}
				if len(actions) == 0 {
					actions = config.Default
				}
				for _, action := range actions {
					modifiers := make([]interface{}, len(action.Modifiers))
					for i, v := range action.Modifiers {
						modifiers[i] = v
					}

					// If the previous word was "capital", add Shift to modifiers
					if capitalizeNext {
						modifiers = append(modifiers, "shift")
					}

					robotgo.KeyTap(action.Key, modifiers...)

					// Reset inside loop for safety between characters
					if len(modifiers) > 0 {
						resetModifiers()
					}
					time.Sleep(20 * time.Millisecond)
				}
			} else {
				// Type raw word if not in alphabet
				textToType := token
				if capitalizeNext {
					r := []rune(textToType)
					if len(r) > 0 {
						r[0] = unicode.ToUpper(r[0])
						textToType = string(r)
					}
				}
				robotgo.TypeStr(textToType)
			}

			// Reset the capitalization flag after the word is typed
			capitalizeNext = false
			// Space out words slightly
			time.Sleep(20 * time.Millisecond)
		}

	default:
		// Default now does nothing (or logs unknown command) to prevent accidental typing
		fmt.Printf("Unknown Command Ignored: %s\n", rawCommand)
	}
}

// --- History Helpers ---

func loadHistory() []HistoryItem {
	history := []HistoryItem{}
	fileBytes, err := os.ReadFile(HistoryFile)
	if err != nil {
		return history
	}
	json.Unmarshal(fileBytes, &history)
	return history
}

func addToHistory(command string) {
	history := loadHistory()

	nextID := 1
	if len(history) > 0 {
		nextID = history[len(history)-1].ID + 1
	}

	newItem := HistoryItem{
		ID:        nextID,
		Command:   command,
		Timestamp: time.Now().Format("15:04:05"),
	}
	history = append(history, newItem)

	if len(history) > 1000 {
		history = history[len(history)-1000:]
	}

	updatedData, err := json.MarshalIndent(history, "", "  ")
	if err == nil {
		os.WriteFile(HistoryFile, updatedData, 0644)
	}
}

// --- Scripting Helpers ---

func loadScripts() map[string]string {
	scripts := make(map[string]string)
	fileBytes, err := os.ReadFile(ScriptsFile)
	if err != nil {
		return scripts
	}
	json.Unmarshal(fileBytes, &scripts)
	return scripts
}

func saveScript(name string, command string) {
	scripts := loadScripts()
	scripts[name] = command
	updatedData, err := json.MarshalIndent(scripts, "", "  ")
	if err == nil {
		os.WriteFile(ScriptsFile, updatedData, 0644)
	}
}

func deleteScript(name string) {
	scripts := loadScripts()
	if _, ok := scripts[name]; ok {
		delete(scripts, name)
		updatedData, err := json.MarshalIndent(scripts, "", "  ")
		if err == nil {
			os.WriteFile(ScriptsFile, updatedData, 0644)
		}
	}
}

// --- Phrases Helpers ---

func loadPhrases() map[string]string {
	phrases := make(map[string]string)
	fileBytes, err := os.ReadFile(PhrasesFile)
	if err != nil {
		return phrases
	}
	json.Unmarshal(fileBytes, &phrases)
	return phrases
}

func deletePhrase(name string) {
	phrases := loadPhrases()
	if _, ok := phrases[name]; ok {
		delete(phrases, name)
		updatedData, err := json.MarshalIndent(phrases, "", "  ")
		if err == nil {
			os.WriteFile(PhrasesFile, updatedData, 0644)
		}
	}
}

// --- Alphabet Helpers (Renamed from Shortcuts) ---

func loadAlphabet() map[string]AlphabetConfig {
	alphabet := make(map[string]AlphabetConfig)
	fileBytes, err := os.ReadFile(AlphabetFile)
	if err != nil {
		return alphabet
	}
	json.Unmarshal(fileBytes, &alphabet)
	return alphabet
}

// --- Ollama Helpers ---

func promptOllama(systemContext, userPrompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  OllamaModel,
		Prompt: userPrompt,
		System: systemContext,
		Stream: false,
	}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(OllamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	var ollamaResp OllamaResponse
	json.Unmarshal(bodyBytes, &ollamaResp)
	return ollamaResp.Response, nil
}

func handleTerminalPhrase(phrase string) {
	fmt.Printf("Generating terminal command for: %s\n", phrase)
	sys := "You are a Linux command generator. Output ONLY the raw command executable."
	response, err := promptOllama(sys, phrase)
	if err == nil {
		clean := strings.TrimSpace(strings.ReplaceAll(response, "`", ""))
		pasteText(clean)
	}
}

// --- Text Helpers ---

func toSnakeCase(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}
func toPascalCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			r := []rune(strings.ToLower(w))
			r[0] = unicode.ToUpper(r[0])
			words[i] = string(r)
		}
	}
	return strings.Join(words, "")
}
func toCamelCase(s string) string {
	s = toPascalCase(s)
	if len(s) > 0 {
		r := []rune(s)
		r[0] = unicode.ToLower(r[0])
		return string(r)
	}
	return s
}

// --- Input Helpers ---

func pasteText(text string) {
	resetModifiers()
	cmdKey := "control"
	if runtime.GOOS == "darwin" {
		cmdKey = "cmd"
	}
	robotgo.KeyToggle(cmdKey, "up")
	orig, _ := robotgo.ReadAll()
	robotgo.WriteAll(text)
	robotgo.KeyTap("v", cmdKey)
	robotgo.KeyToggle(cmdKey, "up")
	time.Sleep(200 * time.Millisecond)
	robotgo.WriteAll(orig)
}

func teleportMouse(phrase string) {
	fileBytes, _ := os.ReadFile("mouse_config.json")
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if pos, ok := positions[phrase]; ok {
		robotgo.Move(pos.X, pos.Y)
		fmt.Printf("Teleported to %s\n", phrase)
	}
}

func attackMouse(phrase string) {
	fileBytes, _ := os.ReadFile("mouse_config.json")
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if pos, ok := positions[phrase]; ok {
		robotgo.Move(pos.X, pos.Y)
		time.Sleep(50 * time.Millisecond)
		robotgo.Click("left", false)
		fmt.Printf("Attacked %s\n", phrase)
	}
}

func rememberMousePosition(phrase string) {
	x, y := robotgo.Location()
	fileBytes, _ := os.ReadFile("mouse_config.json")
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	positions[phrase] = MousePoint{X: x, Y: y}
	data, _ := json.MarshalIndent(positions, "", "  ")
	os.WriteFile("mouse_config.json", data, 0644)
	fmt.Printf("Remembered %s\n", phrase)
}

func forgetMousePosition(phrase string) {
	if phrase == "all" {
		os.WriteFile("mouse_config.json", []byte("{}"), 0644)
		return
	}
	fileBytes, _ := os.ReadFile("mouse_config.json")
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if _, ok := positions[phrase]; ok {
		delete(positions, phrase)
		data, _ := json.MarshalIndent(positions, "", "  ")
		os.WriteFile("mouse_config.json", data, 0644)
	}
}

func handleArrowKeys(direction string, args []string) {
	key := ""
	switch direction {
	case "north":
		key = "up"
	case "south":
		key = "down"
	case "east":
		key = "right"
	case "west":
		key = "left"
	}
	count := 1
	if len(args) > 0 {
		if c := parseFuzzyNumber(args[0]); c > 0 {
			count = c
		}
	}
	for i := 0; i < count; i++ {
		robotgo.KeyTap(key)
		time.Sleep(100 * time.Millisecond)
	}
}

func handleMouse(command string) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return
	}
	direction := parts[0]
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
	case "write":
		robotgo.Move(x+val, y)
	case "right":
		robotgo.Move(x+val, y)
	case "up":
		robotgo.Move(x, y-val)
	case "down":
		robotgo.Move(x, y+val)
	}
}

func handleScroll(command string) {
	parts := strings.Fields(command)
	if len(parts) < 3 {
		fmt.Println("Scroll command requires direction and amount")
		return
	}
	direction := parts[1]
	totalAmount := parseFuzzyNumber(parts[2])
	if totalAmount <= 0 {
		return
	}

	stepSize := 20
	if totalAmount < stepSize {
		stepSize = totalAmount
	}
	steps := totalAmount / stepSize
	remainder := totalAmount % stepSize

	scrollFunc := func(amt int) {
		switch direction {
		case "up":
			robotgo.Scroll(0, -amt)
		case "down":
			robotgo.Scroll(0, amt)
		case "left":
			robotgo.Scroll(-amt, 0)
		case "right":
			robotgo.Scroll(amt, 0)
		}
	}

	for i := 0; i < steps; i++ {
		scrollFunc(stepSize)
		time.Sleep(15 * time.Millisecond)
	}
	if remainder > 0 {
		scrollFunc(remainder)
	}
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

func resetModifiers() {
	robotgo.KeyToggle("shift", "up")
	if runtime.GOOS == "darwin" {
		robotgo.KeyToggle("command", "up")
	} else {
		robotgo.KeyToggle("control", "up")
	}
	robotgo.KeyToggle("alt", "up")
	time.Sleep(time.Millisecond * 20)
}

// getModifierKey returns the robotgo key identifier for a given spoken verb.
// This unifies key mapping logic for both single presses and combinations.
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
		// Maps "command" to "control" on non-Mac systems for cross-platform shortcut compatibility (e.g. Cmd+S -> Ctrl+S)
		return "control"
	case "shift":
		return "shift"
	case "alt":
		return "alt"
	case "option":
		return "lalt"
	case "function":
		return ""
	default:
		return ""
	}
}