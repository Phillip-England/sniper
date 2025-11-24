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
	ClientPort  = "3000"
	ServerPort  = "8000"
	OllamaURL   = "http://localhost:11434/api/generate"
	OllamaModel = "llama3.2" // Ensure you have run: ollama pull llama3.2
)

type MousePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MousePageData struct {
	Locations map[string]MousePoint
	Commands  []string
}

type CharPageData struct {
	Symbols map[string]SymbolConfig
}

// KeyAction defines the specific key press and modifiers
type KeyAction struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers,omitempty"`
}

// SymbolConfig now supports arrays of actions to handle sequences (macros)
type SymbolConfig struct {
	Default []KeyAction `json:"default"`
	Darwin  []KeyAction `json:"darwin,omitempty"`
	Linux   []KeyAction `json:"linux,omitempty"`
	Windows []KeyAction `json:"windows,omitempty"`
}

// Ollama structs for API communication
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
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
	app.At("GET /mouse", func(w http.ResponseWriter, r *http.Request) {
		positions := make(map[string]MousePoint)
		fileBytes, err := os.ReadFile("mouse_config.json")
		if err == nil {
			json.Unmarshal(fileBytes, &positions)
		}
		// Simplified command list for the UI
		staticCmds := []string{
			"teleport [name]", "attack [name]",
			"remember [name]", "forget [name]",
			"click", "rclick", "tclick",
			"left [dist]", "right [dist]", "up [dist]", "down [dist]",
			"scroll up [dist]", "scroll down [dist]",
			"type [text/code]", "sentence [text]",
			"terminal [instruction]",
		}
		data := MousePageData{
			Locations: positions,
			Commands:  staticCmds,
		}
		vii.ExecuteTemplate(w, r, "mouse.html", data)
	})
	app.At("GET /signs", func(w http.ResponseWriter, r *http.Request) {
		data := CharPageData{
			Symbols: loadSymbolMap(),
		}
		vii.ExecuteTemplate(w, r, "signs.html", data)
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

	cmd := strings.ToLower(rawCommand)
	// Normalizations
	cmd = strings.ReplaceAll(cmd, "double click", "dclick")
	cmd = strings.ReplaceAll(cmd, "right click", "rclick")
	cmd = strings.ReplaceAll(cmd, "triple click", "tclick")
	// Note: We don't need to replace "select all" etc anymore because "type select"
	// or "type copyall" handles it via symbols.json

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}
	verb := parts[0]
	args := parts[1:]

	switch verb {
	case "terminal":
		phrase := strings.Join(args, " ")
		go handleTerminalPhrase(phrase)
	case "teleport":
		phrase := strings.Join(args, " ")
		teleportMouse(phrase)
	case "attack":
		phrase := strings.Join(args, " ")
		attackMouse(phrase)
	case "remember":
		phrase := strings.Join(args, " ")
		rememberMousePosition(phrase)
	case "forget":
		phrase := strings.Join(args, " ")
		forgetMousePosition(phrase)
	case "left", "right", "up", "down":
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
	case "sentence":
		rawText := strings.Join(args, " ")
		if len(rawText) > 0 {
			r := []rune(rawText)
			r[0] = unicode.ToUpper(r[0])
			formattedText := string(r)
			formattedText = formattedText + ". "
			pasteText(formattedText)
		}
	case "log":
		fmt.Println("SYSTEM LOG:", strings.Join(args, " "))
	default:
		symbols := loadSymbolMap()
		for _, token := range parts {
			lowerToken := strings.ToLower(token)
			if config, ok := symbols[lowerToken]; ok {
				var actions []KeyAction
				switch runtime.GOOS {
				case "darwin":
					if len(config.Darwin) > 0 {
						actions = config.Darwin
					} else {
						actions = config.Default
					}
				case "linux":
					if len(config.Linux) > 0 {
						actions = config.Linux
					} else {
						actions = config.Default
					}
				case "windows":
					if len(config.Windows) > 0 {
						actions = config.Windows
					} else {
						actions = config.Default
					}
				default:
					actions = config.Default
				}
				for _, action := range actions {
					modifiers := make([]interface{}, len(action.Modifiers))
					for i, v := range action.Modifiers {
						modifiers[i] = v
					}
					if len(actions) > 1 {
						time.Sleep(50 * time.Millisecond)
					}
					robotgo.KeyTap(action.Key, modifiers...)
					for _, mod := range action.Modifiers {
						robotgo.KeyToggle(mod, "up")
					}
				}
			} else {
				robotgo.TypeStr(token)
			}
		}
	}
}

// promptOllama sends a prompt to the running Ollama instance and returns the response text.
func promptOllama(prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  OllamaModel,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(OllamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status: %s", resp.Status)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}

// handleTerminalPhrase specifically asks Ollama to generate a Linux command
func handleTerminalPhrase(phrase string) {
	fmt.Printf("Generating terminal command for: %s\n", phrase)

	systemInstruction := "You are a Linux command generator. Output ONLY the raw command executable for the following request. Do not include markdown formatting, backticks, or explanations. Do not include the preceding $ sign."
	fullPrompt := fmt.Sprintf("%s\nRequest: %s", systemInstruction, phrase)

	response, err := promptOllama(fullPrompt)
	if err != nil {
		fmt.Printf("Error generating command: %v\n", err)
		return
	}

	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.Trim(cleanResponse, "`")
	cleanResponse = strings.TrimSpace(cleanResponse)

	fmt.Printf("Ollama suggested: %s\n", cleanResponse)
	pasteText(cleanResponse)
}

func loadSymbolMap() map[string]SymbolConfig {
	symbols := make(map[string]SymbolConfig)
	fileBytes, err := os.ReadFile("symbols.json")
	if err != nil {
		fmt.Println("Error reading symbols.json:", err)
		return symbols
	}
	if err := json.Unmarshal(fileBytes, &symbols); err != nil {
		fmt.Println("Error parsing symbols.json:", err)
	}
	return symbols
}

func pasteText(text string) {
	cmdKey := "control"
	if runtime.GOOS == "darwin" {
		cmdKey = "cmd"
	}
	robotgo.KeyToggle(cmdKey, "up")
	originalClipboard, _ := robotgo.ReadAll()
	robotgo.WriteAll(text)
	robotgo.KeyTap("v", cmdKey)
	robotgo.KeyToggle(cmdKey, "up")
	time.Sleep(200 * time.Millisecond)
	robotgo.WriteAll(originalClipboard)
}

func teleportMouse(phrase string) {
	fileName := "mouse_config.json"
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading config")
		return
	}
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if pos, exists := positions[phrase]; exists {
		robotgo.Move(pos.X, pos.Y)
		fmt.Printf("Teleported to %s\n", phrase)
	}
}

func attackMouse(phrase string) {
	fileName := "mouse_config.json"
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return
	}
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if pos, exists := positions[phrase]; exists {
		robotgo.Move(pos.X, pos.Y)
		time.Sleep(time.Millisecond * 50)
		robotgo.Click("left", false)
		fmt.Printf("Attacked %s\n", phrase)
	}
}

func rememberMousePosition(phrase string) {
	x, y := robotgo.Location()
	fileName := "mouse_config.json"
	positions := make(map[string]MousePoint)
	fileBytes, err := os.ReadFile(fileName)
	if err == nil {
		json.Unmarshal(fileBytes, &positions)
	}
	positions[phrase] = MousePoint{X: x, Y: y}
	updatedData, _ := json.MarshalIndent(positions, "", "  ")
	os.WriteFile(fileName, updatedData, 0644)
	fmt.Printf("Remembered %s\n", phrase)
}

func forgetMousePosition(phrase string) {
	fileName := "mouse_config.json"
	if phrase == "all" {
		os.WriteFile(fileName, []byte("{}"), 0644)
		fmt.Println("Forgot all positions")
		return
	}
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return
	}
	positions := make(map[string]MousePoint)
	json.Unmarshal(fileBytes, &positions)
	if _, exists := positions[phrase]; exists {
		delete(positions, phrase)
		updatedData, _ := json.MarshalIndent(positions, "", "  ")
		os.WriteFile(fileName, updatedData, 0644)
		fmt.Printf("Forgot %s\n", phrase)
	}
}

func handleMouse(command string) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return
	}
	direction := parts[0]
	val, _ := strconv.Atoi(strings.TrimPrefix(parts[1], "$"))
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

func handleScroll(command string) {
	parts := strings.Fields(command)
	if len(parts) < 3 {
		fmt.Println("Scroll command requires direction and amount")
		return
	}
	direction := parts[1]
	val, _ := strconv.Atoi(strings.TrimPrefix(parts[2], "$"))
	switch direction {
	case "up":
		robotgo.Scroll(0, -val)
	case "down":
		robotgo.Scroll(0, val)
	case "left":
		robotgo.Scroll(-val, 0)
	case "right":
		robotgo.Scroll(val, 0)
	default:
		fmt.Printf("Unknown scroll direction: %s\n", direction)
	}
}