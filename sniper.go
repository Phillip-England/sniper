package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/Phillip-England/vii"
	"github.com/phillip-england/sniper/sniper"
)

// --- CONFIGURATION ---

const (
	ServerPort = "9090"
)

// --- EMBEDDED FILES ---

//go:embed static
var staticEmbed embed.FS

//go:embed templates
var templatesEmbed embed.FS

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

	// Removed MwCORS since everything is now on the same origin
	app.Use(vii.MwTimeout(10))

	// --- Static Files & Templates ---

	staticFS, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		return err
	}
	app.StaticEmbed("/static", staticFS)

	app.Favicon()

	if err := app.TemplatesFS(templatesEmbed, "templates/*.html", nil); err != nil {
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

	// Endpoint: Minimal JSON (Compact)
	app.At("GET /api/commands/min", func(w http.ResponseWriter, r *http.Request) {
		minStr, _, err := sniper.RegistryToJSON()
		if err != nil {
			http.Error(w, "Failed to encode registry: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(minStr))
	})

	// Endpoint: Full JSON (Pretty Printed)
	app.At("GET /api/commands/full", func(w http.ResponseWriter, r *http.Request) {
		_, fullStr, err := sniper.RegistryToJSON()
		if err != nil {
			http.Error(w, "Failed to encode registry: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fullStr))
	})

	app.At("POST /api/data", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Command string `json:"command"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		engine.Parse(req.Command)

		if err := engine.Execute(); err != nil {
			http.Error(w, "Execution Error: "+err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"executed"}`))
	})

	return app.Serve(ServerPort)
}
