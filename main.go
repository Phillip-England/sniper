package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Phillip-England/vii"
	"github.com/phillip-england/sniper/sniper"
)

// --- CONFIGURATION ---

const (
	ServerPort = "8000"
)

// --- MAIN APPLICATION ---

func main() {
	// Initialize the new Engine
	engine := sniper.NewEngine()

	fmt.Printf("Server running on port %s\n", ServerPort)
	if err := runServer(engine); err != nil {
		log.Fatal(err)
	}
}

func runServer(engine *sniper.Engine) error {
	app := vii.NewApp()

	// Removed MwCORS since everything is now on the same origin (port 8000)
	app.Use(vii.MwTimeout(10))

	// --- Static Files & Templates ---
	app.Static("./static")
	app.Favicon()

	if err := app.Templates("./templates", nil); err != nil {
		return err
	}

	// --- UI Routes ---
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

	// --- API Routes ---
	app.At("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is healthy"))
	})

	app.At("POST /api/data", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Command string `json:"command"`
		}

		// Decode JSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 1. Parse the input string into Tokens/Commands
		engine.Parse(req.Command)

		// 2. Execute the parsed tokens
		if err := engine.Execute(); err != nil {
			// Send error back to client
			http.Error(w, "Execution Error: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"executed"}`))
	})

	return app.Serve(ServerPort)
}
