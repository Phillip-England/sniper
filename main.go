package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime" // Imported to detect OS (Windows/Linux vs Mac)
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







	// 1. DETERMINE MODIFIER KEY BASED ON OS
	cmdKey := "control"
	if runtime.GOOS == "darwin" {
		cmdKey = "cmd"
	}

	// 2. NORMALIZE INPUT
	cmd := strings.ToLower(rawCommand)

	// 3. PRE-PROCESS PHRASES
	cmd = strings.ReplaceAll(cmd, "right click", "rclick")
	cmd = strings.ReplaceAll(cmd, "triple click", "tclick")
	cmd = strings.ReplaceAll(cmd, "select all", "selectall")
	cmd = strings.ReplaceAll(cmd, "copy all", "copyall")
	cmd = strings.ReplaceAll(cmd, "cut all", "cutall")
	cmd = strings.ReplaceAll(cmd, "paste all", "pasteall")
	cmd = strings.ReplaceAll(cmd, "shift enter", "shiftenter")
	
	// New logic for Control One
	cmd = strings.ReplaceAll(cmd, "control one", "controlone")

	fmt.Printf("Normalized: %s\n", cmd)

	// 4. PARSE TOKENS
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	verb := parts[0]
	args := parts[1:]

	switch verb {

	// --- Mouse Movement ---
	case "left", "right", "up", "down":
		executeMouseMove(verb, args)

	// --- Mouse Clicks ---
	case "click":
		robotgo.Click("left", false)

	case "rclick":
		robotgo.Click("right", false)

	case "tclick":
		robotgo.Click("left", false)
		robotgo.Click("left", false)
		robotgo.Click("left", false)

	// --- Keyboard Shortcuts (Cross-Platform) ---
	// NOTE: For all shortcut commands, we explicitly call KeyToggle(cmdKey, "up")
	// to ensure the modifier key doesn't get stuck down.

	case "save":
		robotgo.KeyTap("s", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "undo":
		robotgo.KeyTap("z", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "copy":
		robotgo.KeyTap("c", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "copyall":
		// Select All
		robotgo.Click()
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

		time.Sleep(time.Millisecond * 100)

		// Copy
		robotgo.KeyTap("c", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "cut":
		robotgo.KeyTap("x", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "cutall":
		// Select All
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

		time.Sleep(time.Millisecond * 100)

		// Cut
		robotgo.KeyTap("x", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "paste":
		robotgo.KeyTap("v", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "pasteall":
		// Select All
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

		time.Sleep(time.Millisecond * 100)

		// Paste
		robotgo.KeyTap("v", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "selectall":
		robotgo.KeyTap("a", cmdKey)
		robotgo.KeyToggle(cmdKey, "up")

	case "enter":
		robotgo.KeyTap("enter")

	case "shiftenter":
		robotgo.KeyTap("enter", "shift")

	case "controlone":
		// Press "1" with the Modifier key (Cmd/Ctrl)
		robotgo.KeyTap("1", cmdKey)
		// Ensure modifier is released
		robotgo.KeyToggle(cmdKey, "up")

	// --- Text Input ---
	case "type":
		executeTypeCommand(args)

	// --- System ---
	case "log":
		fmt.Println("SYSTEM LOG:", strings.Join(args, " "))

	default:
		fmt.Printf("Unrecognized command: %s\n", verb)
	}
}

// --- LOGIC HANDLERS ---

func executeMouseMove(direction string, args []string) {
	if len(args) < 1 {
		fmt.Println("Movement command requires a distance (e.g., 'left 50')")
		return
	}
	distance, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid distance:", args[0])
		return
	}

	x, y := robotgo.Location()

	switch direction {
	case "left":
		robotgo.Move(x-distance, y)
	case "right":
		robotgo.Move(x+distance, y)
	case "up":
		robotgo.Move(x, y-distance)
	case "down":
		robotgo.Move(x, y+distance)
	}
}

func executeTypeCommand(args []string) {
  // Safety buffer: Ensure previous modifier keys are released
  cmdKey := "control"
  if runtime.GOOS == "darwin" {
    cmdKey = "cmd"
  }
  robotgo.KeyToggle(cmdKey, "up")

  // 1. PRESERVE: Capture the current clipboard content
  originalClipboard, _ := robotgo.ReadAll()

  // Join arguments into a single string
  text := strings.Join(args, " ")

  // 2. OVERWRITE: Copy new text to system clipboard
  robotgo.WriteAll(text + " ")

  // 3. PASTE: Perform the shortcut
  robotgo.KeyTap("v", cmdKey)
  robotgo.KeyToggle(cmdKey, "up")

  // 4. WAIT: This is critical. 
  // If we restore the clipboard immediately, the OS might be too slow 
  // to process the 'Paste' command, and it will accidentally paste 
  // the *restored* text instead of the *new* text.
  time.Sleep(200 * time.Millisecond)

  // 5. RESTORE: Put the original text back
  robotgo.WriteAll(originalClipboard)
}