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
	"time"

	"github.com/Phillip-England/vii"
	"github.com/go-vgo/robotgo"
)

const (
	ClientPort = "3000"
	ServerPort = "8000"
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
	Symbols map[string]string
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

// --- Client Side ---
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

		staticCmds := []string{
			"teleport [name]", "attack [name]",
			"remember [name]", "forget [name]",
			"click", "rclick", "tclick",
			"left [dist]", "right [dist]", "up [dist]", "down [dist]",
			"copy", "cut", "paste", "undo", "save",
			"select all", "copy all", "cut all", "paste all",
			"type [text]", "characters [text]", "enter", "shift enter",
		}

		data := MousePageData{
			Locations: positions,
			Commands:  staticCmds,
		}
		vii.ExecuteTemplate(w, r, "mouse.html", data)
	})

	app.At("GET /characters", func(w http.ResponseWriter, r *http.Request) {
		data := CharPageData{
			Symbols: getSymbolMap(),
		}
		vii.ExecuteTemplate(w, r, "characters.html", data)
	})

	return app.Serve(ClientPort)
}

// --- Server Side ---
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

// --- CENTRAL DISPATCHER ---

func handleCommand(rawCommand string) {
	fmt.Printf("Raw Input: %s\n", rawCommand)

	cmdKey := "control"
	if runtime.GOOS == "darwin" {
		cmdKey = "cmd"
	}

	cmd := strings.ToLower(rawCommand)

	// Normalize
	cmd = strings.ReplaceAll(cmd, "right click", "rclick")
	cmd = strings.ReplaceAll(cmd, "triple click", "tclick")
	cmd = strings.ReplaceAll(cmd, "select all", "selectall")
	cmd = strings.ReplaceAll(cmd, "copy all", "copyall")
	cmd = strings.ReplaceAll(cmd, "cut all", "cutall")
	cmd = strings.ReplaceAll(cmd, "paste all", "pasteall")
	cmd = strings.ReplaceAll(cmd, "shift enter", "shiftenter")
	cmd = strings.ReplaceAll(cmd, "control one", "controlone")
	cmd = strings.ReplaceAll(cmd, "control w", "controlw")

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	verb := parts[0]
	args := parts[1:]

	switch verb {

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

	case "click":
		robotgo.Click("left", false)
	case "rclick":
		robotgo.Click("right", false)
	case "tclick":
		robotgo.Click("left", false)
		robotgo.Click("left", false)
		robotgo.Click("left", false)

	case "next":
		robotgo.KeyTap("space")
	case "save":
		robotgo.KeyTap("s", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "undo":
		robotgo.KeyTap("z", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "copy":
		robotgo.KeyTap("c", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "paste":
		robotgo.KeyTap("v", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "cut":
		robotgo.KeyTap("x", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "selectall":
		robotgo.Click()
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "enter":
		robotgo.KeyTap("enter")
	case "shiftenter":
		robotgo.KeyTap("enter", "shift")
	case "controlone":
		robotgo.KeyTap("1", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "controlw":
		robotgo.KeyTap("w", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "copyall":
		robotgo.Click()
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
		time.Sleep(time.Millisecond * 100)
		robotgo.KeyTap("c", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "cutall":
		robotgo.Click()
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
		time.Sleep(time.Millisecond * 100)
		robotgo.KeyTap("x", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
	case "pasteall":
		robotgo.Click()
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")
		time.Sleep(time.Millisecond * 100)
		robotgo.KeyTap("v", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "type":
		pasteText(strings.Join(args, " "))

	case "characters":
		symbols := getSymbolMap()
		// Join args so we have a raw string: "bang bang hash"
		fullArgString := strings.Join(args, " ")

		// Replace the words with symbols
		for phrase, symbol := range symbols {
			fullArgString = strings.ReplaceAll(fullArgString, phrase, symbol)
		}

		// Strip remaining spaces so "bang bang" becomes "!!" instead of "! !"
		// Note: The map below maps "gap" to a unique placeholder if you actually need a space.
		finalString := strings.ReplaceAll(fullArgString, " ", "")
		pasteText(finalString)

	case "log":
		fmt.Println("SYSTEM LOG:", strings.Join(args, " "))

	default:
		fmt.Printf("Unrecognized command: %s\n", verb)
	}
}

// --- SYMBOL DICTIONARY (Sonic / Phonetic Optimized) ---

func getSymbolMap() map[string]string {
	return map[string]string{
		// -- Wrappers --
		// Distinct words for open/close to prevent "close brace" lag errors
		"box":   "[",
		"kit":   "]",
		"curl":  "{",
		"lock":  "}",
		"open":  "(",
		"close": ")",
		"less":  "<",
		"great": ">",

		// -- Math / Logic --
		"plus":  "+",
		"equal": "=",
		"dash":  "-",
		"floor": "_", // "Underscore" is too long
		"star":  "*",
		"per":   "%", // "Percent" often gets heard as "Per scent"
		"hat":   "^", // "Caret" is often confused with "Carrot"

		// -- Punctuation --
		"bang":  "!",
		"at":    "@",
		"hash":  "#",
		"cash":  "$",
		"amp":   "&",
		"pipe":  "|",
		"wall":  "|", // Alternative to pipe
		"back":  "\\",
		"slash": "/",
		"col":   ":",
		"semi":  ";",
		"dub":   "\"",
		"tick":  "'",
		"com":   ",", // "Comma" is fine, but "Com" is punchier
		"dot":   ".",
		"quest": "?",
		"wave":  "~",
		"grave": "`",

		// -- Spacing --
		"gap": " ", // Use this if you specifically want a space inside a character string
	}
}

// --- HELPER LOGIC ---

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